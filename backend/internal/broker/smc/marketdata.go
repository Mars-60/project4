package smc

import (
	"context"

	"github.com/Mars-60/project4/backend/internal/broker"
)

type quotePayload struct {
	Exchange    string `json:"exchange"`
	SymbolToken string `json:"symbol_token"`
}

type quoteDataResponse struct {
	Exchange      string  `json:"exchange"`
	SymbolToken   string  `json:"symbol_token"`
	TradingSymbol string  `json:"trading_symbol"`
	LastPrice     float64 `json:"last_price"`
	Open          float64 `json:"open"`
	High          float64 `json:"high"`
	Low           float64 `json:"low"`
	Close         float64 `json:"close"`
	Volume        int64   `json:"volume"`
	Timestamp     string  `json:"timestamp"`
}

type candlePayload struct {
	Exchange    string `json:"exchange"`
	SymbolToken string `json:"symbol_token"`
	Interval    string `json:"interval"`
	From        string `json:"from"`
	To          string `json:"to"`
}

type candleDataResponse struct {
	Timestamp string  `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    int64   `json:"volume"`
}

func (c *Client) GetQuote(ctx context.Context, request broker.QuoteRequest) (broker.Quote, error) {
	payload := quotePayload{
		Exchange:    request.Exchange,
		SymbolToken: request.SymbolToken,
	}

	var response Envelope[quoteDataResponse]
	if err := c.transport.Post(ctx, endpointQuote, payload, &response); err != nil {
		return broker.Quote{}, err
	}

	if !response.OK() {
		return broker.Quote{}, &APIError{Status: response.Status, Message: response.Message}
	}

	return broker.Quote{
		Exchange:      response.Data.Exchange,
		SymbolToken:   response.Data.SymbolToken,
		TradingSymbol: response.Data.TradingSymbol,
		LastPrice:     response.Data.LastPrice,
		Open:          response.Data.Open,
		High:          response.Data.High,
		Low:           response.Data.Low,
		Close:         response.Data.Close,
		Volume:        response.Data.Volume,
		Timestamp:     parseTime(response.Data.Timestamp),
	}, nil
}

func (c *Client) GetCandles(ctx context.Context, request broker.CandleRequest) ([]broker.Candle, error) {
	payload := candlePayload{
		Exchange:    request.Exchange,
		SymbolToken: request.SymbolToken,
		Interval:    request.Interval,
		From:        request.From.Format("2006-01-02 15:04:05"),
		To:          request.To.Format("2006-01-02 15:04:05"),
	}

	var response Envelope[[]candleDataResponse]
	if err := c.transport.Post(ctx, endpointCandles, payload, &response); err != nil {
		return nil, err
	}

	if !response.OK() {
		return nil, &APIError{Status: response.Status, Message: response.Message}
	}

	candles := make([]broker.Candle, 0, len(response.Data))
	for _, item := range response.Data {
		candles = append(candles, broker.Candle{
			Timestamp: parseTime(item.Timestamp),
			Open:      item.Open,
			High:      item.High,
			Low:       item.Low,
			Close:     item.Close,
			Volume:    item.Volume,
		})
	}

	return candles, nil
}
