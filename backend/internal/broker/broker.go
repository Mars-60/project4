package broker

import "context"

type Broker interface {
	Login(
		ctx context.Context,
		request LoginRequest,
	) (LoginResponse, error)

	GenerateAccessToken(
		ctx context.Context,
		request TokenRequest,
	) (TokenResponse, error)

	GetProfile(ctx context.Context) (Profile, error)
	GetFunds(ctx context.Context) (Funds, error)
	GetHoldings(ctx context.Context) ([]Holding, error)
	GetPositions(ctx context.Context) ([]Position, error)
	PlaceOrder(ctx context.Context, request OrderRequest) (OrderResponse, error)
	ModifyOrder(ctx context.Context, request ModifyOrderRequest) (OrderResponse, error)
	CancelOrder(ctx context.Context, request CancelOrderRequest) (OrderResponse, error)
	GetOrderBook(ctx context.Context) ([]Order, error)
	GetTradeBook(ctx context.Context) ([]Trade, error)
	GetQuote(ctx context.Context, request QuoteRequest) (Quote, error)
	GetCandles(ctx context.Context, request CandleRequest) ([]Candle, error)
}
