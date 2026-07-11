package main

import (
	"github.com/Mars-60/project4/backend/configs"
	"github.com/Mars-60/project4/backend/internal/api"
	"github.com/Mars-60/project4/backend/internal/logger"
)

func main() {

	if err := configs.Load(); err != nil {
		panic(err)
	}

	if err := logger.InitWithLevel(configs.App.Env, configs.App.Log.Level); err != nil {
		panic(err)
	}

	defer logger.Log.Sync()

	logger.Log.Info(
		"Starting TradePilot AI",
	)

	if err := api.StartServer(); err != nil {
		logger.Log.Fatal(
			"Server crashed",
		)
	}
}
