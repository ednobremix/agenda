package controllers

import (
	"encoding/json"
	"massage-scheduler/models"
	"net/http"
)

func HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var configStore models.ConfigStore
	err := models.GORM.Where("user_email = ?", email).First(&configStore).Error
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"businessType":"Massoterapia","professionals":[],"services":[]}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(configStore.Data))
}

func HandleSaveConfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var cfg models.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bytes, _ := json.Marshal(cfg)
	configStore := models.ConfigStore{
		UserEmail: email,
		Data:      string(bytes),
	}

	if err := models.GORM.Save(&configStore).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
