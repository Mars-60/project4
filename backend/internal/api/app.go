package api

import (
	"context"
	"fmt"

	"github.com/Mars-60/project4/backend/configs"
	"github.com/Mars-60/project4/backend/internal/ai"
	"github.com/Mars-60/project4/backend/internal/auth"
	"github.com/Mars-60/project4/backend/internal/broker"
	"github.com/Mars-60/project4/backend/internal/broker/smc"
	"github.com/Mars-60/project4/backend/internal/core"
	"github.com/Mars-60/project4/backend/internal/database"
	"github.com/Mars-60/project4/backend/internal/notifications"
)

type App struct {
	Auth          *auth.Service
	Strategies    *core.StrategyManager
	Market        *core.MarketService
	Execution     *core.ExecutionService
	Portfolio     *core.PortfolioService
	AI            *ai.ConversationService
	Notifications *notifications.Service
	Repo          Repository
	DB            *database.DB
}

type Repository interface {
	core.UserRepository
	core.StrategyRepository
	core.OrderRepository
	core.TradeRepository
	core.PositionRepository
	core.HoldingRepository
	core.FundsRepository
	core.ConversationRepository
	core.NotificationRepository
	core.RefreshSessionRepository
	core.TransactionManager
}

func NewApp(ctx context.Context) (*App, error) {
	repo, db, err := newRepository(ctx)
	if err != nil {
		return nil, err
	}
	clock := core.SystemClock{}
	ids := core.TimeIDGenerator{}

	authService := auth.NewService(repo, repo, clock, ids, configs.App.Auth.JWTSecret, configs.App.Auth.AccessTokenTTL, configs.App.Auth.RefreshTokenTTL, configs.App.Auth.PasswordIterations)
	registry := core.NewStrategyRegistry()
	strategyManager := core.NewStrategyManager(repo, registry, clock, ids)
	risk := core.NewRiskManager(core.RiskRule{
		MaxOrderValue:     500000,
		MaxPositionValue:  1000000,
		MaxDailyLoss:      50000,
		MaxSymbolExposure: 250000,
	})
	execution := core.NewExecutionService(repo, repo, repo, repo, repo, risk, clock, ids)
	portfolio := core.NewPortfolioService(repo, repo, repo)
	aiClient := ai.NewGroqClient(configs.App.Groq.BaseURL, configs.App.Groq.APIKey, configs.App.Groq.Model, configs.App.Groq.Timeout, configs.App.Groq.MaxRetries)
	conversations := ai.NewConversationService(aiClient, repo, clock, ids)
	notificationService := notifications.NewService(repo, clock, ids)
	notificationService.Register(notifications.NewLogProvider("email"))
	notificationService.Register(notifications.NewLogProvider("telegram"))
	notificationService.Register(notifications.NewLogProvider("whatsapp"))
	notificationService.Register(notifications.NewLogProvider("push"))

	marketProvider := NewBrokerMarketProvider(newSMCClient())
	market := core.NewMarketService(marketProvider, configs.App.SMC.Timeout)

	return &App{
		Auth: authService, Strategies: strategyManager, Market: market, Execution: execution,
		Portfolio: portfolio, AI: conversations, Notifications: notificationService, Repo: repo, DB: db,
	}, nil
}

func newRepository(ctx context.Context) (Repository, *database.DB, error) {
	if configs.App.DB.DSN == "" {
		return database.NewMemoryRepository(), nil, nil
	}
	db, err := database.Open(ctx, database.Config{
		Driver:          configs.App.DB.Driver,
		DSN:             configs.App.DB.DSN,
		MaxOpenConns:    configs.App.DB.MaxOpenConns,
		MaxIdleConns:    configs.App.DB.MaxIdleConns,
		ConnMaxLifetime: configs.App.DB.ConnMaxLifetime,
	})
	if err != nil {
		return nil, nil, err
	}
	return database.NewPostgresRepository(db), db, nil
}

func newSMCClient() *smc.Client {
	return smc.NewClient(
		configs.App.SMC.BaseURL,
		configs.App.SMC.APIKey,
		configs.App.SMC.APISecret,
		smc.WithTimeout(configs.App.SMC.Timeout),
		smc.WithMaxRetries(configs.App.SMC.MaxRetries),
		smc.WithWebSocketURL(configs.App.SMC.WebSocketURL),
	)
}

type BrokerMarketProvider struct {
	broker broker.Broker
}

func NewBrokerMarketProvider(broker broker.Broker) BrokerMarketProvider {
	return BrokerMarketProvider{broker: broker}
}

func (p BrokerMarketProvider) Quote(ctx context.Context, exchange string, symbolToken string) (core.MarketSnapshot, error) {
	if exchange == "" || symbolToken == "" {
		return core.MarketSnapshot{}, fmt.Errorf("exchange and symbol_token are required")
	}
	quote, err := p.broker.GetQuote(ctx, broker.QuoteRequest{Exchange: exchange, SymbolToken: symbolToken})
	if err != nil {
		return core.MarketSnapshot{}, err
	}
	return core.MarketSnapshot{
		Exchange: quote.Exchange, SymbolToken: quote.SymbolToken, TradingSymbol: quote.TradingSymbol,
		LastPrice: quote.LastPrice, Open: quote.Open, High: quote.High, Low: quote.Low, Close: quote.Close,
		Volume: quote.Volume, Timestamp: quote.Timestamp,
	}, nil
}
