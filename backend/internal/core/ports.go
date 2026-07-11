package core

import (
	"context"
	"time"
)

type MarketDataProvider interface {
	Quote(ctx context.Context, exchange string, symbolToken string) (MarketSnapshot, error)
}

type Strategy interface {
	ID() ID
	Name() string
	Evaluate(ctx context.Context, snapshot MarketSnapshot, config StrategyConfig) ([]Signal, error)
}

type StrategyRepository interface {
	CreateStrategy(ctx context.Context, strategy StrategyDefinition) (StrategyDefinition, error)
	ListStrategies(ctx context.Context, userID ID) ([]StrategyDefinition, error)
	UpdateStrategyStatus(ctx context.Context, userID ID, strategyID ID, status StrategyStatus) error
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order Order) (Order, error)
	UpdateOrder(ctx context.Context, order Order) error
	ListOrders(ctx context.Context, userID ID, filter PageFilter) ([]Order, error)
}

type TradeRepository interface {
	CreateTrade(ctx context.Context, trade Trade) (Trade, error)
	ListTrades(ctx context.Context, userID ID, filter PageFilter) ([]Trade, error)
}

type PositionRepository interface {
	UpsertPosition(ctx context.Context, position Position) error
	ListPositions(ctx context.Context, userID ID, paper bool) ([]Position, error)
}

type HoldingRepository interface {
	ListHoldings(ctx context.Context, userID ID) ([]Holding, error)
	UpsertHolding(ctx context.Context, holding Holding) error
}

type FundsRepository interface {
	GetFunds(ctx context.Context, userID ID) (Funds, error)
	SaveFunds(ctx context.Context, funds Funds) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	FindUserByEmail(ctx context.Context, email string) (User, error)
	FindUserByID(ctx context.Context, id ID) (User, error)
}

type ConversationRepository interface {
	SaveConversation(ctx context.Context, conversation AIConversation) error
	ListConversations(ctx context.Context, userID ID, filter PageFilter) ([]AIConversation, error)
}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification Notification) (Notification, error)
	ListNotifications(ctx context.Context, userID ID, filter PageFilter) ([]Notification, error)
}

type RefreshSessionRepository interface {
	SaveRefreshSession(ctx context.Context, session RefreshSession) error
	FindRefreshSession(ctx context.Context, token string) (RefreshSession, error)
	RevokeRefreshSession(ctx context.Context, token string, revokedAt time.Time) error
}

type TransactionManager interface {
	WithinTx(ctx context.Context, fn func(context.Context) error) error
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

type IDGenerator interface {
	NewID() ID
}

type TimeIDGenerator struct{}

func (TimeIDGenerator) NewID() ID {
	return ID(time.Now().UTC().Format("20060102150405.000000000"))
}

type PageFilter struct {
	Limit  int
	Offset int
	Sort   string
}

func NormalizePageFilter(filter PageFilter) PageFilter {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return filter
}
