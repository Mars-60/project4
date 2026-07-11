package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Mars-60/project4/backend/internal/core"
	"github.com/go-chi/chi/v5"
)

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}
	if !decode(w, r, &req) {
		return
	}
	user, tokens, err := a.Auth.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	user.PasswordHash = ""
	WriteJSON(w, http.StatusCreated, map[string]any{"user": user, "tokens": tokens})
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !decode(w, r, &req) {
		return
	}
	user, tokens, err := a.Auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user.PasswordHash = ""
	WriteJSON(w, http.StatusOK, map[string]any{"user": user, "tokens": tokens})
}

func (a *App) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &req) {
		return
	}
	tokens, err := a.Auth.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, tokens)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if !decode(w, r, &req) {
		return
	}
	if err := a.Auth.Logout(r.Context(), req.RefreshToken); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (a *App) Me(w http.ResponseWriter, r *http.Request) {
	claims, _ := ClaimsFromContext(r.Context())
	WriteJSON(w, http.StatusOK, claims)
}

func (a *App) CreateStrategy(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string              `json:"name"`
		Description string              `json:"description"`
		Config      core.StrategyConfig `json:"config"`
	}
	if !decode(w, r, &req) {
		return
	}
	strategy, err := a.Strategies.Create(r.Context(), core.StrategyDefinition{
		UserID: UserIDFromContext(r.Context()), Name: req.Name, Description: req.Description, Config: req.Config,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, strategy)
}

func (a *App) ListStrategies(w http.ResponseWriter, r *http.Request) {
	strategies, err := a.Repo.ListStrategies(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, strategies)
}

func (a *App) EnableStrategy(w http.ResponseWriter, r *http.Request) {
	if err := a.Strategies.Enable(r.Context(), UserIDFromContext(r.Context()), core.ID(chi.URLParam(r, "id"))); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "enabled"})
}

func (a *App) DisableStrategy(w http.ResponseWriter, r *http.Request) {
	if err := a.Strategies.Disable(r.Context(), UserIDFromContext(r.Context()), core.ID(chi.URLParam(r, "id"))); err != nil {
		WriteError(w, http.StatusNotFound, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

func (a *App) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := a.Repo.ListOrders(r.Context(), UserIDFromContext(r.Context()), pageFilter(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, orders)
}

func (a *App) ListTrades(w http.ResponseWriter, r *http.Request) {
	trades, err := a.Repo.ListTrades(r.Context(), UserIDFromContext(r.Context()), pageFilter(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, trades)
}

func (a *App) PlacePaperOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Signal    core.Signal `json:"signal"`
		LastPrice float64     `json:"last_price"`
	}
	if !decode(w, r, &req) {
		return
	}
	order, err := a.Execution.PlacePaperOrder(r.Context(), UserIDFromContext(r.Context()), req.Signal, req.LastPrice)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, order)
}

func (a *App) PortfolioSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := a.Portfolio.Summary(r.Context(), UserIDFromContext(r.Context()), queryBool(r, "paper"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

func (a *App) Positions(w http.ResponseWriter, r *http.Request) {
	positions, err := a.Portfolio.Positions(r.Context(), UserIDFromContext(r.Context()), queryBool(r, "paper"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, positions)
}

func (a *App) Holdings(w http.ResponseWriter, r *http.Request) {
	holdings, err := a.Portfolio.Holdings(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, holdings)
}

func (a *App) Funds(w http.ResponseWriter, r *http.Request) {
	funds, err := a.Repo.GetFunds(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, funds)
}

func (a *App) Quote(w http.ResponseWriter, r *http.Request) {
	quote, err := a.Market.Quote(r.Context(), r.URL.Query().Get("exchange"), r.URL.Query().Get("symbol_token"))
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, quote)
}

func (a *App) AskAI(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if !decode(w, r, &req) {
		return
	}
	conversation, err := a.AI.Ask(r.Context(), UserIDFromContext(r.Context()), req.Prompt)
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, conversation)
}

func (a *App) AnalyzePortfolio(w http.ResponseWriter, r *http.Request) {
	summary, err := a.Portfolio.Summary(r.Context(), UserIDFromContext(r.Context()), queryBool(r, "paper"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	conversation, err := a.AI.AnalyzePortfolio(r.Context(), UserIDFromContext(r.Context()), summary)
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, conversation)
}

func (a *App) ExplainMarket(w http.ResponseWriter, r *http.Request) {
	quote, err := a.Market.Quote(r.Context(), r.URL.Query().Get("exchange"), r.URL.Query().Get("symbol_token"))
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	conversation, err := a.AI.ExplainMarket(r.Context(), UserIDFromContext(r.Context()), quote)
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, conversation)
}

func (a *App) ExplainRisk(w http.ResponseWriter, r *http.Request) {
	summary, err := a.Portfolio.Summary(r.Context(), UserIDFromContext(r.Context()), queryBool(r, "paper"))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	conversation, err := a.AI.ExplainRisk(r.Context(), UserIDFromContext(r.Context()), summary)
	if err != nil {
		WriteError(w, http.StatusBadGateway, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, conversation)
}

func (a *App) ExplainStrategy(w http.ResponseWriter, r *http.Request) {
	strategies, err := a.Repo.ListStrategies(r.Context(), UserIDFromContext(r.Context()))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	id := core.ID(chi.URLParam(r, "id"))
	for _, strategy := range strategies {
		if strategy.ID == id {
			conversation, err := a.AI.ExplainStrategy(r.Context(), UserIDFromContext(r.Context()), strategy)
			if err != nil {
				WriteError(w, http.StatusBadGateway, err.Error())
				return
			}
			WriteJSON(w, http.StatusOK, conversation)
			return
		}
	}
	WriteError(w, http.StatusNotFound, "strategy not found")
}

func (a *App) ListNotifications(w http.ResponseWriter, r *http.Request) {
	items, err := a.Repo.ListNotifications(r.Context(), UserIDFromContext(r.Context()), pageFilter(r))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	WriteJSON(w, http.StatusOK, items)
}

func (a *App) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if !decode(w, r, &req) {
		return
	}
	notification, err := a.Notifications.Notify(r.Context(), UserIDFromContext(r.Context()), req.Channel, req.Subject, req.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	WriteJSON(w, http.StatusCreated, notification)
}

func (a *App) SystemMetrics(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]any{"time": time.Now().UTC(), "status": "ok"})
}

func decode(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid json body")
		return false
	}
	return true
}

func pageFilter(r *http.Request) core.PageFilter {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	return core.NormalizePageFilter(core.PageFilter{Limit: limit, Offset: offset, Sort: r.URL.Query().Get("sort")})
}

func queryBool(r *http.Request, key string) bool {
	value, _ := strconv.ParseBool(r.URL.Query().Get(key))
	return value
}
