package smc

const (
	endpointLogin       = "/auth/login"
	endpointToken       = "/auth/token"
	endpointProfile     = "/auth/profile"
	endpointFunds       = "/user/funds"
	endpointHoldings    = "/portfolio/holdings"
	endpointPositions   = "/portfolio/positions"
	endpointPlaceOrder  = "/orders"
	endpointModifyOrder = "/orders/modify"
	endpointCancelOrder = "/orders/cancel"
	endpointOrderBook   = "/orders/book"
	endpointTradeBook   = "/orders/trades"
	endpointQuote       = "/market/quote"
	endpointCandles     = "/market/candles"
)

type LoginPayload struct {
	Platform string    `json:"platform"`
	Data     LoginData `json:"data"`
}

type LoginData struct {
	ClientID string `json:"client_id"`
	Password string `json:"password"`
}

type AccessTokenPayload struct {
	APIKey    string `json:"api_key"`
	Signature string `json:"signature"`
	ReqToken  string `json:"req_token"`
}
