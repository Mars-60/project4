package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Mars-60/project4/backend/configs"
	"github.com/Mars-60/project4/backend/internal/broker"
	"github.com/Mars-60/project4/backend/internal/broker/smc"
)

func main() {

	if err := configs.Load(); err != nil {
		log.Fatal(err)
	}

	client := smc.NewClient(
		configs.App.SMC.BaseURL,
		configs.App.SMC.APIKey,
		configs.App.SMC.APISecret,
	)

	loginResponse, err := client.Login(
		context.Background(),
		broker.LoginRequest{
			ClientID: configs.App.SMC.ClientID,
			Password: configs.App.SMC.Password,
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("========== LOGIN ==========")
	fmt.Println("Message :", loginResponse.Message)
	fmt.Println("Success :", loginResponse.Success)
	fmt.Println("Need TOTP :", loginResponse.NeedTOTP)
	fmt.Println("Request Token :", loginResponse.RequestToken)

	tokenResponse, err := client.GenerateAccessToken(
		context.Background(),
		broker.TokenRequest{
			RequestToken: loginResponse.RequestToken,
		},
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n========== TOKEN ==========")
	fmt.Println("Access Token :", tokenResponse.AccessToken)
	fmt.Println("Feed Token :", tokenResponse.FeedToken)

	profile, err := client.GetProfile(
		context.Background(),
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n========== PROFILE ==========")
	fmt.Println("Client :", profile.ClientID)
	fmt.Println("Name :", profile.Name)
	fmt.Println("Email :", profile.Email)
	fmt.Println("Mobile :", profile.Mobile)
}