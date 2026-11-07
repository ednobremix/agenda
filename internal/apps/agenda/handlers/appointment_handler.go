package handlers

import (
	"encoding/json"
	"errors"
	"massage-scheduler/internal/apps/agenda/services"
	"massage-scheduler/internal/common/models"
	"net/http"

	"gorm.io/gorm"
)

func HandleGetAppointments(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	apps, err := services.GetAppointments(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

func HandleSaveAppointments(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var apps []common_models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&apps); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := services.SaveAppointments(email, apps)
	if err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			http.Error(w, "Agendamento livre desativado: limite de 1 agendamento por dia excedido.", http.StatusBadRequest)
			return
		}
		if err.Error() == "acesso negado ao App Agenda" || err.Error() == "não é permitido cancelar sessões que já passaram e foram pagas" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
