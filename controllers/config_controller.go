package controllers

import (
	"encoding/json"
	"massage-scheduler/models"
	"net/http"
)

func HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var data string
	err := models.DB.QueryRow("SELECT data FROM configs WHERE user_email = ?", email).Scan(&data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"businessType":"Massoterapia","professionals":[],"services":[]}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(data))
}

func HandleSaveConfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var cfg models.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bytes, _ := json.Marshal(cfg)
	_, err := models.DB.Exec("INSERT INTO configs (user_email, data) VALUES (?, ?) ON CONFLICT(user_email) DO UPDATE SET data = excluded.data", email, string(bytes))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
