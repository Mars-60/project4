package api

import (
	"github.com/Mars-60/project4/backend/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router *chi.Mux, app *App) {

	router.Get("/", handlers.HomeHandler)

	router.Route("/api/v1", func(r chi.Router) {

		r.Get("/health", handlers.HealthHandler)
		r.Get("/version", handlers.VersionHandler)
		r.Post("/auth/register", app.Register)
		r.Post("/auth/login", app.Login)
		r.Post("/auth/refresh", app.Refresh)
		r.Post("/auth/logout", app.Logout)

		r.Group(func(protected chi.Router) {
			protected.Use(AuthMiddleware(app.Auth))

			protected.Get("/auth/me", app.Me)
			protected.Get("/strategies", app.ListStrategies)
			protected.Post("/strategies", app.CreateStrategy)
			protected.Post("/strategies/{id}/enable", app.EnableStrategy)
			protected.Post("/strategies/{id}/disable", app.DisableStrategy)
			protected.Post("/strategies/{id}/explain", app.ExplainStrategy)
			protected.Get("/orders", app.ListOrders)
			protected.Get("/trades", app.ListTrades)
			protected.Post("/paper/orders", app.PlacePaperOrder)
			protected.Get("/portfolio/summary", app.PortfolioSummary)
			protected.Get("/portfolio/positions", app.Positions)
			protected.Get("/portfolio/holdings", app.Holdings)
			protected.Get("/funds", app.Funds)
			protected.Get("/market/quote", app.Quote)
			protected.Post("/ai/ask", app.AskAI)
			protected.Post("/ai/portfolio", app.AnalyzePortfolio)
			protected.Post("/ai/market", app.ExplainMarket)
			protected.Post("/ai/risk", app.ExplainRisk)
			protected.Get("/notifications", app.ListNotifications)
			protected.Post("/notifications", app.CreateNotification)
			protected.Get("/system/metrics", app.SystemMetrics)
		})

	})

}
