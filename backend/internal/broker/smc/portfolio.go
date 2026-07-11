package smc

import (
	"context"

	"github.com/Mars-60/project4/backend/internal/broker"
)

type holdingDataResponse struct {
	Symbol       string  `json:"symbol"`
	Exchange     string  `json:"exchange"`
	ISIN         string  `json:"isin"`
	Quantity     int     `json:"quantity"`
	AveragePrice float64 `json:"average_price"`
	LastPrice    float64 `json:"last_price"`
	PnL          float64 `json:"pnl"`
}

type positionDataResponse struct {
	Symbol       string  `json:"symbol"`
	Exchange     string  `json:"exchange"`
	ProductType  string  `json:"product_type"`
	Quantity     int     `json:"quantity"`
	BuyQuantity  int     `json:"buy_quantity"`
	SellQuantity int     `json:"sell_quantity"`
	AveragePrice float64 `json:"average_price"`
	LastPrice    float64 `json:"last_price"`
	PnL          float64 `json:"pnl"`
}

func (c *Client) GetHoldings(ctx context.Context) ([]broker.Holding, error) {
	var response Envelope[[]holdingDataResponse]
	if err := c.transport.Get(ctx, endpointHoldings, &response); err != nil {
		return nil, err
	}

	if !response.OK() {
		return nil, &APIError{Status: response.Status, Message: response.Message}
	}

	holdings := make([]broker.Holding, 0, len(response.Data))
	for _, item := range response.Data {
		holdings = append(holdings, broker.Holding{
			Symbol:       item.Symbol,
			Exchange:     item.Exchange,
			ISIN:         item.ISIN,
			Quantity:     item.Quantity,
			AveragePrice: item.AveragePrice,
			LastPrice:    item.LastPrice,
			PnL:          item.PnL,
		})
	}

	return holdings, nil
}

func (c *Client) GetPositions(ctx context.Context) ([]broker.Position, error) {
	var response Envelope[[]positionDataResponse]
	if err := c.transport.Get(ctx, endpointPositions, &response); err != nil {
		return nil, err
	}

	if !response.OK() {
		return nil, &APIError{Status: response.Status, Message: response.Message}
	}

	positions := make([]broker.Position, 0, len(response.Data))
	for _, item := range response.Data {
		positions = append(positions, broker.Position{
			Symbol:       item.Symbol,
			Exchange:     item.Exchange,
			ProductType:  item.ProductType,
			Quantity:     item.Quantity,
			BuyQuantity:  item.BuyQuantity,
			SellQuantity: item.SellQuantity,
			AveragePrice: item.AveragePrice,
			LastPrice:    item.LastPrice,
			PnL:          item.PnL,
		})
	}

	return positions, nil
}
