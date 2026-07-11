package smc

import (
	"context"

	"github.com/Mars-60/project4/backend/internal/broker"
)

type orderPayload struct {
	Exchange        string  `json:"exchange"`
	SymbolToken     string  `json:"symbol_token"`
	TradingSymbol   string  `json:"trading_symbol"`
	TransactionType string  `json:"transaction_type"`
	OrderType       string  `json:"order_type"`
	ProductType     string  `json:"product_type"`
	Validity        string  `json:"validity"`
	Price           float64 `json:"price,omitempty"`
	TriggerPrice    float64 `json:"trigger_price,omitempty"`
	Quantity        int     `json:"quantity"`
	DisclosedQty    int     `json:"disclosed_qty,omitempty"`
	Tag             string  `json:"tag,omitempty"`
}

type modifyOrderPayload struct {
	OrderID      string  `json:"order_id"`
	OrderType    string  `json:"order_type"`
	Validity     string  `json:"validity"`
	Price        float64 `json:"price,omitempty"`
	TriggerPrice float64 `json:"trigger_price,omitempty"`
	Quantity     int     `json:"quantity"`
}

type cancelOrderPayload struct {
	OrderID string `json:"order_id"`
}

type orderResponseData struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

func (c *Client) PlaceOrder(ctx context.Context, request broker.OrderRequest) (broker.OrderResponse, error) {
	payload := orderPayload{
		Exchange:        request.Exchange,
		SymbolToken:     request.SymbolToken,
		TradingSymbol:   request.TradingSymbol,
		TransactionType: request.TransactionType,
		OrderType:       request.OrderType,
		ProductType:     request.ProductType,
		Validity:        request.Validity,
		Price:           request.Price,
		TriggerPrice:    request.TriggerPrice,
		Quantity:        request.Quantity,
		DisclosedQty:    request.DisclosedQty,
		Tag:             request.Tag,
	}

	var response Envelope[orderResponseData]
	if err := c.transport.Post(ctx, endpointPlaceOrder, payload, &response); err != nil {
		return broker.OrderResponse{}, err
	}

	return mapOrderResponse(response)
}

func (c *Client) ModifyOrder(ctx context.Context, request broker.ModifyOrderRequest) (broker.OrderResponse, error) {
	payload := modifyOrderPayload{
		OrderID:      request.OrderID,
		OrderType:    request.OrderType,
		Validity:     request.Validity,
		Price:        request.Price,
		TriggerPrice: request.TriggerPrice,
		Quantity:     request.Quantity,
	}

	var response Envelope[orderResponseData]
	if err := c.transport.Put(ctx, endpointModifyOrder, payload, &response); err != nil {
		return broker.OrderResponse{}, err
	}

	return mapOrderResponse(response)
}

func (c *Client) CancelOrder(ctx context.Context, request broker.CancelOrderRequest) (broker.OrderResponse, error) {
	var response Envelope[orderResponseData]
	if err := c.transport.Delete(ctx, endpointCancelOrder, cancelOrderPayload{OrderID: request.OrderID}, &response); err != nil {
		return broker.OrderResponse{}, err
	}

	return mapOrderResponse(response)
}

func mapOrderResponse(response Envelope[orderResponseData]) (broker.OrderResponse, error) {
	if !response.OK() {
		return broker.OrderResponse{}, &APIError{Status: response.Status, Message: response.Message}
	}

	return broker.OrderResponse{
		OrderID: response.Data.OrderID,
		Status:  response.Data.Status,
		Message: response.Message,
	}, nil
}
