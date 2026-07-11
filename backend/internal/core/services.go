package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MarketService struct {
	provider MarketDataProvider
	cache    *MarketCache
}

func NewMarketService(provider MarketDataProvider, ttl time.Duration) *MarketService {
	return &MarketService{provider: provider, cache: NewMarketCache(ttl)}
}

func (s *MarketService) Quote(ctx context.Context, exchange string, symbolToken string) (MarketSnapshot, error) {
	key := exchange + ":" + symbolToken
	if snapshot, ok := s.cache.Get(key); ok {
		return snapshot, nil
	}
	snapshot, err := s.provider.Quote(ctx, exchange, symbolToken)
	if err != nil {
		return MarketSnapshot{}, err
	}
	s.cache.Set(key, snapshot)
	return snapshot, nil
}

type MarketCache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]cacheItem
}

type cacheItem struct {
	snapshot  MarketSnapshot
	expiresAt time.Time
}

func NewMarketCache(ttl time.Duration) *MarketCache {
	return &MarketCache{ttl: ttl, items: make(map[string]cacheItem)}
}

func (c *MarketCache) Get(key string) (MarketSnapshot, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(item.expiresAt) {
		return MarketSnapshot{}, false
	}
	return item.snapshot, true
}

func (c *MarketCache) Set(key string, snapshot MarketSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cacheItem{snapshot: snapshot, expiresAt: time.Now().Add(c.ttl)}
}

type StrategyRegistry struct {
	mu         sync.RWMutex
	strategies map[ID]Strategy
}

func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{strategies: make(map[ID]Strategy)}
}

func (r *StrategyRegistry) Register(strategy Strategy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if strategy == nil {
		return fmt.Errorf("strategy is nil")
	}
	if _, exists := r.strategies[strategy.ID()]; exists {
		return fmt.Errorf("strategy already registered: %s", strategy.ID())
	}
	r.strategies[strategy.ID()] = strategy
	return nil
}

func (r *StrategyRegistry) Get(id ID) (Strategy, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	strategy, ok := r.strategies[id]
	return strategy, ok
}

type StrategyManager struct {
	repo     StrategyRepository
	registry *StrategyRegistry
	clock    Clock
	ids      IDGenerator
}

func NewStrategyManager(repo StrategyRepository, registry *StrategyRegistry, clock Clock, ids IDGenerator) *StrategyManager {
	return &StrategyManager{repo: repo, registry: registry, clock: clock, ids: ids}
}

func (m *StrategyManager) Create(ctx context.Context, strategy StrategyDefinition) (StrategyDefinition, error) {
	now := m.clock.Now()
	strategy.ID = m.ids.NewID()
	strategy.Status = StrategyDisabled
	strategy.CreatedAt = now
	strategy.UpdatedAt = now
	return m.repo.CreateStrategy(ctx, strategy)
}

func (m *StrategyManager) Enable(ctx context.Context, userID ID, strategyID ID) error {
	return m.repo.UpdateStrategyStatus(ctx, userID, strategyID, StrategyEnabled)
}

func (m *StrategyManager) Disable(ctx context.Context, userID ID, strategyID ID) error {
	return m.repo.UpdateStrategyStatus(ctx, userID, strategyID, StrategyDisabled)
}

func (m *StrategyManager) Evaluate(ctx context.Context, definition StrategyDefinition, snapshot MarketSnapshot) ([]Signal, error) {
	if definition.Status != StrategyEnabled {
		return nil, nil
	}
	strategy, ok := m.registry.Get(definition.ID)
	if !ok {
		return nil, fmt.Errorf("strategy implementation not registered: %s", definition.ID)
	}
	return strategy.Evaluate(ctx, snapshot, definition.Config)
}

type RiskManager struct {
	rules RiskRule
}

func NewRiskManager(rules RiskRule) *RiskManager {
	return &RiskManager{rules: rules}
}

func (r *RiskManager) Validate(order Order, funds Funds, positions []Position) RiskDecision {
	if order.Quantity <= 0 {
		return RiskDecision{Allowed: false, Reason: "quantity must be greater than zero"}
	}
	if order.Exchange == "" || order.SymbolToken == "" || order.TradingSymbol == "" {
		return RiskDecision{Allowed: false, Reason: "exchange, symbol token and trading symbol are required"}
	}
	if order.TransactionType != string(SignalBuy) && order.TransactionType != string(SignalSell) {
		return RiskDecision{Allowed: false, Reason: "transaction type must be BUY or SELL"}
	}
	if order.Price <= 0 {
		return RiskDecision{Allowed: false, Reason: "price must be greater than zero"}
	}
	if order.StopLoss > 0 && order.TransactionType == string(SignalBuy) && order.StopLoss >= order.Price {
		return RiskDecision{Allowed: false, Reason: "buy stop loss must be below entry price"}
	}
	if order.Target > 0 && order.TransactionType == string(SignalBuy) && order.Target <= order.Price {
		return RiskDecision{Allowed: false, Reason: "buy target must be above entry price"}
	}
	if order.TrailingStop > 0 && order.TrailBy <= 0 {
		return RiskDecision{Allowed: false, Reason: "trail amount is required when trailing stop is enabled"}
	}
	orderValue := float64(order.Quantity) * order.Price
	if r.rules.MaxOrderValue > 0 && orderValue > r.rules.MaxOrderValue {
		return RiskDecision{Allowed: false, Reason: "order value exceeds risk limit"}
	}
	if r.rules.MaxSymbolExposure > 0 {
		exposure := orderValue
		for _, position := range positions {
			if position.TradingSymbol == order.TradingSymbol {
				exposure += abs(float64(position.Quantity) * position.LastPrice)
			}
		}
		if exposure > r.rules.MaxSymbolExposure {
			return RiskDecision{Allowed: false, Reason: "symbol exposure exceeds risk limit"}
		}
	}
	if r.rules.MaxPositionValue > 0 && CalculateExposure(positions)+orderValue > r.rules.MaxPositionValue {
		return RiskDecision{Allowed: false, Reason: "portfolio exposure exceeds risk limit"}
	}
	if order.TransactionType == "SELL" && !r.rules.AllowShortSelling {
		for _, position := range positions {
			if position.TradingSymbol == order.TradingSymbol && position.Quantity < order.Quantity {
				return RiskDecision{Allowed: false, Reason: "short selling is disabled"}
			}
		}
	}
	if orderValue > funds.Available {
		return RiskDecision{Allowed: false, Reason: "insufficient funds"}
	}
	return RiskDecision{Allowed: true}
}

func CalculateUnrealizedPnL(quantity int, averagePrice float64, lastPrice float64) float64 {
	return float64(quantity) * (lastPrice - averagePrice)
}

func CalculateExposure(positions []Position) float64 {
	var exposure float64
	for _, position := range positions {
		exposure += abs(float64(position.Quantity) * position.LastPrice)
	}
	return exposure
}

func abs(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

type PortfolioService struct {
	positions PositionRepository
	holdings  HoldingRepository
	funds     FundsRepository
}

func NewPortfolioService(positions PositionRepository, holdings HoldingRepository, funds FundsRepository) *PortfolioService {
	return &PortfolioService{positions: positions, holdings: holdings, funds: funds}
}

func (s *PortfolioService) Summary(ctx context.Context, userID ID, paper bool) (PortfolioSummary, error) {
	positions, err := s.positions.ListPositions(ctx, userID, paper)
	if err != nil {
		return PortfolioSummary{}, err
	}
	funds, err := s.funds.GetFunds(ctx, userID)
	if err != nil {
		return PortfolioSummary{}, err
	}
	var realized, unrealized float64
	for _, position := range positions {
		realized += position.RealizedPnL
		unrealized += position.UnrealizedPnL
	}
	return PortfolioSummary{
		UserID:        userID,
		NetValue:      funds.Available + unrealized,
		Cash:          funds.Available,
		UsedMargin:    funds.UsedMargin,
		RealizedPnL:   realized,
		UnrealizedPnL: unrealized,
		Exposure:      CalculateExposure(positions),
		UpdatedAt:     time.Now().UTC(),
	}, nil
}

func (s *PortfolioService) Positions(ctx context.Context, userID ID, paper bool) ([]Position, error) {
	return s.positions.ListPositions(ctx, userID, paper)
}

func (s *PortfolioService) Holdings(ctx context.Context, userID ID) ([]Holding, error) {
	return s.holdings.ListHoldings(ctx, userID)
}

type ExecutionService struct {
	orders    OrderRepository
	trades    TradeRepository
	positions PositionRepository
	funds     FundsRepository
	tx        TransactionManager
	risk      *RiskManager
	clock     Clock
	ids       IDGenerator
}

func NewExecutionService(orders OrderRepository, trades TradeRepository, positions PositionRepository, funds FundsRepository, tx TransactionManager, risk *RiskManager, clock Clock, ids IDGenerator) *ExecutionService {
	return &ExecutionService{orders: orders, trades: trades, positions: positions, funds: funds, tx: tx, risk: risk, clock: clock, ids: ids}
}

func (s *ExecutionService) PlacePaperOrder(ctx context.Context, userID ID, signal Signal, lastPrice float64) (Order, error) {
	var order Order
	err := s.withTx(ctx, func(txCtx context.Context) error {
		placed, err := s.placePaperOrder(txCtx, userID, signal, lastPrice)
		if err != nil {
			return err
		}
		order = placed
		return nil
	})
	return order, err
}

func (s *ExecutionService) placePaperOrder(ctx context.Context, userID ID, signal Signal, lastPrice float64) (Order, error) {
	now := s.clock.Now()
	order := Order{
		ID:              s.ids.NewID(),
		UserID:          userID,
		StrategyID:      signal.StrategyID,
		Exchange:        signal.Exchange,
		SymbolToken:     signal.SymbolToken,
		TradingSymbol:   signal.TradingSymbol,
		TransactionType: string(signal.Action),
		OrderType:       signal.OrderType,
		ProductType:     signal.ProductType,
		Quantity:        signal.Quantity,
		Price:           choosePrice(signal.Price, lastPrice),
		AveragePrice:    choosePrice(signal.Price, lastPrice),
		StopLoss:        signal.StopLoss,
		Target:          signal.Target,
		TrailingStop:    signal.StopLoss,
		TrailBy:         0,
		Status:          OrderFilled,
		Paper:           true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	funds, err := s.funds.GetFunds(ctx, userID)
	if err != nil {
		return Order{}, err
	}
	positions, err := s.positions.ListPositions(ctx, userID, true)
	if err != nil {
		return Order{}, err
	}
	if decision := s.risk.Validate(order, funds, positions); !decision.Allowed {
		order.Status = OrderRejected
		order.RejectReason = decision.Reason
		return s.orders.CreateOrder(ctx, order)
	}
	order, err = s.orders.CreateOrder(ctx, order)
	if err != nil {
		return Order{}, err
	}
	_, err = s.trades.CreateTrade(ctx, Trade{
		ID:              s.ids.NewID(),
		UserID:          userID,
		OrderID:         order.ID,
		Exchange:        order.Exchange,
		TradingSymbol:   order.TradingSymbol,
		TransactionType: order.TransactionType,
		Quantity:        order.Quantity,
		Price:           order.AveragePrice,
		Paper:           true,
		TradedAt:        now,
	})
	if err != nil {
		return Order{}, err
	}
	position := applyFill(existingPosition(positions, order), order, lastPrice, s.ids.NewID(), now)
	if err := s.positions.UpsertPosition(ctx, position); err != nil {
		return Order{}, err
	}
	funds.Available -= float64(signedQuantity(order.TransactionType, order.Quantity)) * order.AveragePrice
	funds.PaperBalance = funds.Available
	funds.Net = funds.Available + position.UnrealizedPnL
	funds.UpdatedAt = now
	return order, s.funds.SaveFunds(ctx, funds)
}

func (s *ExecutionService) withTx(ctx context.Context, fn func(context.Context) error) error {
	if s.tx == nil {
		return fn(ctx)
	}
	return s.tx.WithinTx(ctx, fn)
}

func choosePrice(preferred float64, fallback float64) float64 {
	if preferred > 0 {
		return preferred
	}
	return fallback
}

func signedQuantity(side string, quantity int) int {
	if side == "SELL" {
		return -quantity
	}
	return quantity
}

func existingPosition(positions []Position, order Order) Position {
	for _, position := range positions {
		if position.TradingSymbol == order.TradingSymbol && position.ProductType == order.ProductType && position.Paper == order.Paper {
			return position
		}
	}
	return Position{ID: "", UserID: order.UserID, Exchange: order.Exchange, TradingSymbol: order.TradingSymbol, ProductType: order.ProductType, Paper: order.Paper}
}

func applyFill(position Position, order Order, lastPrice float64, fallbackID ID, now time.Time) Position {
	fillQty := signedQuantity(order.TransactionType, order.Quantity)
	if position.ID == "" {
		position.ID = fallbackID
	}
	oldQty := position.Quantity
	newQty := oldQty + fillQty
	if oldQty == 0 || (oldQty > 0 && fillQty > 0) || (oldQty < 0 && fillQty < 0) {
		totalCost := abs(float64(oldQty))*position.AveragePrice + abs(float64(fillQty))*order.AveragePrice
		if newQty != 0 {
			position.AveragePrice = totalCost / abs(float64(newQty))
		}
	} else {
		closedQty := minInt(absInt(oldQty), absInt(fillQty))
		if oldQty > 0 {
			position.RealizedPnL += float64(closedQty) * (order.AveragePrice - position.AveragePrice)
		} else {
			position.RealizedPnL += float64(closedQty) * (position.AveragePrice - order.AveragePrice)
		}
		if newQty == 0 {
			position.AveragePrice = 0
		} else if (oldQty > 0 && newQty < 0) || (oldQty < 0 && newQty > 0) {
			position.AveragePrice = order.AveragePrice
		}
	}
	position.Quantity = newQty
	position.LastPrice = lastPrice
	position.UnrealizedPnL = CalculateUnrealizedPnL(position.Quantity, position.AveragePrice, lastPrice)
	position.UpdatedAt = now
	return position
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
