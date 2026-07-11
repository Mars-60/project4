package smc

import (
	"context"
	"fmt"

	"github.com/Mars-60/project4/backend/internal/broker"
)

type loginDataResponse struct {
	RequestToken string `json:"request_token"`
	Is2FAEnabled bool   `json:"is_2fa_enabled"`
}

type accessTokenDataResponse struct {
	AccessToken string `json:"access_token"`
	FeedToken   string `json:"feed_token"`
}

func (c *Client) Login(ctx context.Context, request broker.LoginRequest) (broker.LoginResponse, error) {
	payload := LoginPayload{
		Platform: "api",
		Data: LoginData{
			ClientID: request.ClientID,
			Password: request.Password,
		},
	}

	var response Envelope[loginDataResponse]
	if err := c.transport.Post(ctx, endpointLogin, payload, &response); err != nil {
		return broker.LoginResponse{}, err
	}

	if !response.OK() {
		return broker.LoginResponse{}, &APIError{Status: response.Status, Message: response.Message}
	}

	return broker.LoginResponse{
		RequestToken: response.Data.RequestToken,
		NeedTOTP:     response.Data.Is2FAEnabled,
		Success:      true,
		Message:      response.Message,
	}, nil
}

func (c *Client) GenerateAccessToken(ctx context.Context, request broker.TokenRequest) (broker.TokenResponse, error) {
	if request.RequestToken == "" {
		return broker.TokenResponse{}, fmt.Errorf("request token is required")
	}

	payload := AccessTokenPayload{
		APIKey:    c.apiKey,
		Signature: c.GenerateSignature(request.RequestToken),
		ReqToken:  request.RequestToken,
	}

	var response Envelope[accessTokenDataResponse]
	if err := c.transport.Post(ctx, endpointToken, payload, &response); err != nil {
		return broker.TokenResponse{}, err
	}

	if !response.OK() {
		return broker.TokenResponse{}, &APIError{Status: response.Status, Message: response.Message}
	}

	c.SetTokens(response.Data.AccessToken, response.Data.FeedToken)

	return broker.TokenResponse{
		AccessToken: response.Data.AccessToken,
		FeedToken:   response.Data.FeedToken,
		Success:     true,
		Message:     response.Message,
	}, nil
}
