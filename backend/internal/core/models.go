package core

import "time"

type ID string

type User struct {
	ID           ID
	Email        string
	Name         string
	Role         string
	PasswordHash string
	Active       bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type StrategyStatus string

const (
	StrategyEnabled  StrategyStatus = "enabled"
	StrategyDisabled StrategyStatus = "disabled"
)

type StrategyConfig map[string]string

type StrategyDefinition struct {
	ID          ID
	UserID      ID
	Name        string
	Description string
	Status      StrategyStatus
	Config      StrategyConfig
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MarketSnapshot struct {
	Exchange      string
	SymbolToken   string
	TradingSymbol string
	LastPrice     float64
	Open          float64
	High          float64
	Low           float64
	Close         float64
	Volume        int64
	Timestamp     time.Time
}

type SignalAction string

const (
	SignalBuy  SignalAction = "BUY"
	SignalSell SignalAction = "SELL"
	SignalHold SignalAction = "HOLD"
)

type Signal struct {
	StrategyID    ID
	Exchange      string
	SymbolToken   string
	TradingSymbol string
	Action        SignalAction
	ProductType   string
	OrderType     string
	Quantity      int
	Price         float64
	StopLoss      float64
	Target        float64
	Confidence    float64
	Reason        string
	GeneratedAt   time.Time
}

type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderValidated OrderStatus = "validated"
	OrderRejected  OrderStatus = "rejected"
	OrderPlaced    OrderStatus = "placed"
	OrderFilled    OrderStatus = "filled"
	OrderCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID              ID
	UserID          ID
	StrategyID      ID
	BrokerOrderID   string
	Exchange        string
	SymbolToken     string
	TradingSymbol   string
	TransactionType string
	OrderType       string
	ProductType     string
	Quantity        int
	FilledQuantity  int
	Price           float64
	AveragePrice    float64
	StopLoss        float64
	Target          float64
	TrailingStop    float64
	TrailBy         float64
	Status          OrderStatus
	RejectReason    string
	Paper           bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Trade struct {
	ID              ID
	UserID          ID
	OrderID         ID
	Exchange        string
	TradingSymbol   string
	TransactionType string
	Quantity        int
	Price           float64
	Paper           bool
	TradedAt        time.Time
}

type Position struct {
	ID            ID
	UserID        ID
	Exchange      string
	TradingSymbol string
	ProductType   string
	Quantity      int
	AveragePrice  float64
	LastPrice     float64
	RealizedPnL   float64
	UnrealizedPnL float64
	Paper         bool
	UpdatedAt     time.Time
}

type Holding struct {
	ID            ID
	UserID        ID
	Exchange      string
	TradingSymbol string
	ISIN          string
	Quantity      int
	AveragePrice  float64
	LastPrice     float64
	PnL           float64
	UpdatedAt     time.Time
}

type Funds struct {
	UserID       ID
	Available    float64
	UsedMargin   float64
	Opening      float64
	Net          float64
	PaperBalance float64
	UpdatedAt    time.Time
}

type RefreshSession struct {
	Token     string
	UserID    ID
	Role      string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RiskRule struct {
	MaxOrderValue     float64
	MaxPositionValue  float64
	MaxDailyLoss      float64
	MaxSymbolExposure float64
	AllowShortSelling bool
}

type RiskDecision struct {
	Allowed bool
	Reason  string
}

type PortfolioSummary struct {
	UserID        ID
	NetValue      float64
	Cash          float64
	UsedMargin    float64
	RealizedPnL   float64
	UnrealizedPnL float64
	Exposure      float64
	UpdatedAt     time.Time
}

type Notification struct {
	ID        ID
	UserID    ID
	Channel   string
	Subject   string
	Body      string
	Status    string
	CreatedAt time.Time
}

type AIConversation struct {
	ID        ID
	UserID    ID
	Title     string
	Messages  []AIMessage
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AIMessage struct {
	Role      string
	Content   string
	CreatedAt time.Time
}
