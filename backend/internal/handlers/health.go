package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Mars-60/project4/backend/configs"
)

type HealthResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
}

type VersionResponse struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {

	response := HealthResponse{
		Status:      "healthy",
		Service:     configs.App.Name,
		Version:     configs.App.Version,
		Environment: configs.App.Env,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
	}
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	response := VersionResponse{
		Service: configs.App.Name,
		Version: configs.App.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"failed to encode response",
			http.StatusInternalServerError,
		)
	}
}
