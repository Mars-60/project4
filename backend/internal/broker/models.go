package broker

import "time"

type LoginRequest struct {
	ClientID string
	Password string
}

type LoginResponse struct {
	RequestToken string
	NeedTOTP     bool
	Success      bool
	Message      string
}

type TokenRequest struct {
	RequestToken string
}

type TokenResponse struct {
	AccessToken string
	FeedToken   string
	Success     bool
	Message     string
}
type Profile struct {
	ClientID string
	Name     string
	Email    string
	Mobile   string

	Exchanges []string
	Products  []string
}

type Funds struct {
	AvailableCash  float64
	UsedMargin     float64
	OpeningBalance float64
	NetBalance     float64
	PayIn          float64
	PayOut         float64
}

type Holding struct {
	Symbol       string
	Exchange     string
	ISIN         string
	Quantity     int
	AveragePrice float64
	LastPrice    float64
	PnL          float64
}

type Position struct {
	Symbol       string
	Exchange     string
	ProductType  string
	Quantity     int
	BuyQuantity  int
	SellQuantity int
	AveragePrice float64
	LastPrice    float64
	PnL          float64
}

type OrderRequest struct {
	Exchange        string
	SymbolToken     string
	TradingSymbol   string
	TransactionType string
	OrderType       string
	ProductType     string
	Validity        string
	Price           float64
	TriggerPrice    float64
	Quantity        int
	DisclosedQty    int
	Tag             string
}

type ModifyOrderRequest struct {
	OrderID      string
	OrderType    string
	Validity     string
	Price        float64
	TriggerPrice float64
	Quantity     int
}

type CancelOrderRequest struct {
	OrderID string
}

type OrderResponse struct {
	OrderID string
	Status  string
	Message string
}

type Order struct {
	OrderID         string
	Exchange        string
	SymbolToken     string
	TradingSymbol   string
	TransactionType string
	OrderType       string
	ProductType     string
	Status          string
	Quantity        int
	FilledQuantity  int
	PendingQuantity int
	Price           float64
	AveragePrice    float64
	TriggerPrice    float64
	OrderTime       time.Time
	UpdateTime      time.Time
	RejectionReason string
	Tag             string
}

type Trade struct {
	TradeID         string
	OrderID         string
	Exchange        string
	TradingSymbol   string
	TransactionType string
	Quantity        int
	Price           float64
	TradeTime       time.Time
}

type QuoteRequest struct {
	Exchange    string
	SymbolToken string
}

type Quote struct {
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

type CandleRequest struct {
	Exchange    string
	SymbolToken string
	Interval    string
	From        time.Time
	To          time.Time
}

type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int64
}
