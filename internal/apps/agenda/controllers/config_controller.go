package controllers

import (
	"encoding/json"
	"massage-scheduler/internal/common/models"
	"net/http"
)

func HandleGetConfig(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var configStore common_models.ConfigStore
	err := common_models.GORM.Where("user_email = ?", email).First(&configStore).Error
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
	var cfg common_models.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bytes, _ := json.Marshal(cfg)
	configStore := common_models.ConfigStore{
		UserEmail: email,
		Data:      string(bytes),
	}

	if err := common_models.GORM.Save(&configStore).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
