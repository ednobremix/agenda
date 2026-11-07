package handlers

import (
	"encoding/json"
	"massage-scheduler/internal/apps/agenda/services"
	"massage-scheduler/internal/common/models"
	"net/http"
)

func HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	configData, err := services.GetConfig(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(configData))
}

func HandleSaveConfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var cfg common_models.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := services.SaveConfig(email, cfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
