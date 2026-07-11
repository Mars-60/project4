package smc

import (
	"context"

	"github.com/Mars-60/project4/backend/internal/broker"
)

type orderDataResponse struct {
	OrderID         string  `json:"order_id"`
	Exchange        string  `json:"exchange"`
	SymbolToken     string  `json:"symbol_token"`
	TradingSymbol   string  `json:"trading_symbol"`
	TransactionType string  `json:"transaction_type"`
	OrderType       string  `json:"order_type"`
	ProductType     string  `json:"product_type"`
	Status          string  `json:"status"`
	Quantity        int     `json:"quantity"`
	FilledQuantity  int     `json:"filled_quantity"`
	PendingQuantity int     `json:"pending_quantity"`
	Price           float64 `json:"price"`
	AveragePrice    float64 `json:"average_price"`
	TriggerPrice    float64 `json:"trigger_price"`
	OrderTime       string  `json:"order_time"`
	UpdateTime      string  `json:"update_time"`
	RejectionReason string  `json:"rejection_reason"`
	Tag             string  `json:"tag"`
}

type tradeDataResponse struct {
	TradeID         string  `json:"trade_id"`
	OrderID         string  `json:"order_id"`
	Exchange        string  `json:"exchange"`
	TradingSymbol   string  `json:"trading_symbol"`
	TransactionType string  `json:"transaction_type"`
	Quantity        int     `json:"quantity"`
	Price           float64 `json:"price"`
	TradeTime       string  `json:"trade_time"`
}

func (c *Client) GetOrderBook(ctx context.Context) ([]broker.Order, error) {
	var response Envelope[[]orderDataResponse]
	if err := c.transport.Get(ctx, endpointOrderBook, &response); err != nil {
		return nil, err
	}

	if !response.OK() {
		return nil, &APIError{Status: response.Status, Message: response.Message}
	}

	orders := make([]broker.Order, 0, len(response.Data))
	for _, item := range response.Data {
		orders = append(orders, broker.Order{
			OrderID:         item.OrderID,
			Exchange:        item.Exchange,
			SymbolToken:     item.SymbolToken,
			TradingSymbol:   item.TradingSymbol,
			TransactionType: item.TransactionType,
			OrderType:       item.OrderType,
			ProductType:     item.ProductType,
			Status:          item.Status,
			Quantity:        item.Quantity,
			FilledQuantity:  item.FilledQuantity,
			PendingQuantity: item.PendingQuantity,
			Price:           item.Price,
			AveragePrice:    item.AveragePrice,
			TriggerPrice:    item.TriggerPrice,
			OrderTime:       parseTime(item.OrderTime),
			UpdateTime:      parseTime(item.UpdateTime),
			RejectionReason: item.RejectionReason,
			Tag:             item.Tag,
		})
	}

	return orders, nil
}

func (c *Client) GetTradeBook(ctx context.Context) ([]broker.Trade, error) {
	var response Envelope[[]tradeDataResponse]
	if err := c.transport.Get(ctx, endpointTradeBook, &response); err != nil {
		return nil, err
	}

	if !response.OK() {
		return nil, &APIError{Status: response.Status, Message: response.Message}
	}

	trades := make([]broker.Trade, 0, len(response.Data))
	for _, item := range response.Data {
		trades = append(trades, broker.Trade{
			TradeID:         item.TradeID,
			OrderID:         item.OrderID,
			Exchange:        item.Exchange,
			TradingSymbol:   item.TradingSymbol,
			TransactionType: item.TransactionType,
			Quantity:        item.Quantity,
			Price:           item.Price,
			TradeTime:       parseTime(item.TradeTime),
		})
	}

	return trades, nil
}
