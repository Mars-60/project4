package smc

import (
	"context"

	"github.com/Mars-60/project4/backend/internal/broker"
)

type profileDataResponse struct {
	ClientID  string   `json:"client_id"`
	Name      string   `json:"name"`
	Email     string   `json:"email_id"`
	Mobile    string   `json:"mobile_no"`
	Exchanges []string `json:"exchanges"`
	Products  []string `json:"products"`
}

func (c *Client) GetProfile(ctx context.Context) (broker.Profile, error) {
	var response Envelope[profileDataResponse]
	if err := c.transport.Get(ctx, endpointProfile, &response); err != nil {
		return broker.Profile{}, err
	}

	if !response.OK() {
		return broker.Profile{}, &APIError{Status: response.Status, Message: response.Message}
	}

	return broker.Profile{
		ClientID:  response.Data.ClientID,
		Name:      response.Data.Name,
		Email:     response.Data.Email,
		Mobile:    response.Data.Mobile,
		Exchanges: response.Data.Exchanges,
		Products:  response.Data.Products,
	}, nil
}
