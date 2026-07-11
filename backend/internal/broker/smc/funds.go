package smc

import (
	"context"

	"github.com/Mars-60/project4/backend/internal/broker"
)

type fundsDataResponse struct {
	AvailableCash  float64 `json:"available_cash"`
	UsedMargin     float64 `json:"used_margin"`
	OpeningBalance float64 `json:"opening_balance"`
	NetBalance     float64 `json:"net_balance"`
	PayIn          float64 `json:"payin"`
	PayOut         float64 `json:"payout"`
}

func (c *Client) GetFunds(ctx context.Context) (broker.Funds, error) {
	var response Envelope[fundsDataResponse]
	if err := c.transport.Get(ctx, endpointFunds, &response); err != nil {
		return broker.Funds{}, err
	}

	if !response.OK() {
		return broker.Funds{}, &APIError{Status: response.Status, Message: response.Message}
	}

	return broker.Funds{
		AvailableCash:  response.Data.AvailableCash,
		UsedMargin:     response.Data.UsedMargin,
		OpeningBalance: response.Data.OpeningBalance,
		NetBalance:     response.Data.NetBalance,
		PayIn:          response.Data.PayIn,
		PayOut:         response.Data.PayOut,
	}, nil
}
