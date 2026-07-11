package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Mars-60/project4/backend/internal/core"
)

type MemoryRepository struct {
	mu            sync.RWMutex
	users         map[core.ID]core.User
	usersByEmail  map[string]core.ID
	strategies    map[core.ID]core.StrategyDefinition
	orders        map[core.ID]core.Order
	trades        map[core.ID]core.Trade
	positions     map[string]core.Position
	holdings      map[string]core.Holding
	funds         map[core.ID]core.Funds
	conversations map[core.ID]core.AIConversation
	notifications map[core.ID]core.Notification
	sessions      map[string]core.RefreshSession
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		users:         map[core.ID]core.User{},
		usersByEmail:  map[string]core.ID{},
		strategies:    map[core.ID]core.StrategyDefinition{},
		orders:        map[core.ID]core.Order{},
		trades:        map[core.ID]core.Trade{},
		positions:     map[string]core.Position{},
		holdings:      map[string]core.Holding{},
		funds:         map[core.ID]core.Funds{},
		conversations: map[core.ID]core.AIConversation{},
		notifications: map[core.ID]core.Notification{},
		sessions:      map[string]core.RefreshSession{},
	}
}

func (r *MemoryRepository) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (r *MemoryRepository) CreateUser(ctx context.Context, user core.User) (core.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.usersByEmail[user.Email]; exists {
		return core.User{}, fmt.Errorf("email already exists")
	}
	r.users[user.ID] = user
	r.usersByEmail[user.Email] = user.ID
	return user, nil
}

func (r *MemoryRepository) FindUserByEmail(ctx context.Context, email string) (core.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.usersByEmail[email]
	if !ok {
		return core.User{}, ErrNotFound
	}
	return r.users[id], nil
}

func (r *MemoryRepository) FindUserByID(ctx context.Context, id core.ID) (core.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[id]
	if !ok {
		return core.User{}, ErrNotFound
	}
	return user, nil
}

func (r *MemoryRepository) CreateStrategy(ctx context.Context, strategy core.StrategyDefinition) (core.StrategyDefinition, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategies[strategy.ID] = strategy
	return strategy, nil
}

func (r *MemoryRepository) ListStrategies(ctx context.Context, userID core.ID) ([]core.StrategyDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []core.StrategyDefinition
	for _, strategy := range r.strategies {
		if strategy.UserID == userID {
			out = append(out, strategy)
		}
	}
	return out, nil
}

func (r *MemoryRepository) UpdateStrategyStatus(ctx context.Context, userID core.ID, strategyID core.ID, status core.StrategyStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	strategy, ok := r.strategies[strategyID]
	if !ok || strategy.UserID != userID {
		return ErrNotFound
	}
	strategy.Status = status
	r.strategies[strategyID] = strategy
	return nil
}

func (r *MemoryRepository) CreateOrder(ctx context.Context, order core.Order) (core.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return order, nil
}

func (r *MemoryRepository) UpdateOrder(ctx context.Context, order core.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
	return nil
}

func (r *MemoryRepository) ListOrders(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []core.Order
	for _, order := range r.orders {
		if order.UserID == userID {
			out = append(out, order)
		}
	}
	return out, nil
}

func (r *MemoryRepository) CreateTrade(ctx context.Context, trade core.Trade) (core.Trade, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.trades[trade.ID] = trade
	return trade, nil
}

func (r *MemoryRepository) ListTrades(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.Trade, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []core.Trade
	for _, trade := range r.trades {
		if trade.UserID == userID {
			out = append(out, trade)
		}
	}
	return out, nil
}

func (r *MemoryRepository) UpsertPosition(ctx context.Context, position core.Position) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.positions[string(position.UserID)+":"+position.TradingSymbol+":"+position.ProductType+":"+fmt.Sprint(position.Paper)] = position
	return nil
}

func (r *MemoryRepository) ListPositions(ctx context.Context, userID core.ID, paper bool) ([]core.Position, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []core.Position
	for _, position := range r.positions {
		if position.UserID == userID && position.Paper == paper {
			out = append(out, position)
		}
	}
	return out, nil
}

func (r *MemoryRepository) UpsertHolding(ctx context.Context, holding core.Holding) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.holdings[string(holding.UserID)+":"+holding.TradingSymbol] = holding
	return nil
}

func (r *MemoryRepository) ListHoldings(ctx context.Context, userID core.ID) ([]core.Holding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []core.Holding
	for _, holding := range r.holdings {
		if holding.UserID == userID {
			out = append(out, holding)
		}
	}
	return out, nil
}

func (r *MemoryRepository) GetFunds(ctx context.Context, userID core.ID) (core.Funds, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	funds, ok := r.funds[userID]
	if !ok {
		return core.Funds{UserID: userID, Available: 1000000, Net: 1000000, PaperBalance: 1000000}, nil
	}
	return funds, nil
}

func (r *MemoryRepository) SaveFunds(ctx context.Context, funds core.Funds) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.funds[funds.UserID] = funds
	return nil
}

func (r *MemoryRepository) SaveConversation(ctx context.Context, conversation core.AIConversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conversations[conversation.ID] = conversation
	return nil
}

func (r *MemoryRepository) ListConversations(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.AIConversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []core.AIConversation
	for _, conversation := range r.conversations {
		if conversation.UserID == userID {
			out = append(out, conversation)
		}
	}
	return out, nil
}

func (r *MemoryRepository) CreateNotification(ctx context.Context, notification core.Notification) (core.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.notifications[notification.ID] = notification
	return notification, nil
}

func (r *MemoryRepository) ListNotifications(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []core.Notification
	for _, notification := range r.notifications {
		if notification.UserID == userID {
			out = append(out, notification)
		}
	}
	return out, nil
}

func (r *MemoryRepository) SaveRefreshSession(ctx context.Context, session core.RefreshSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.Token] = session
	return nil
}

func (r *MemoryRepository) FindRefreshSession(ctx context.Context, token string) (core.RefreshSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[token]
	if !ok {
		return core.RefreshSession{}, ErrNotFound
	}
	return session, nil
}

func (r *MemoryRepository) RevokeRefreshSession(ctx context.Context, token string, revokedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.sessions[token]
	if !ok {
		return ErrNotFound
	}
	session.RevokedAt = &revokedAt
	r.sessions[token] = session
	return nil
}
