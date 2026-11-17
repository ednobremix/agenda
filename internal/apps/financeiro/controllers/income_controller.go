package controllers

import (
	"encoding/json"
	"massage-scheduler/internal/common/models"
	"net/http"
)

func HandleGetIncomes(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	var incomes []common_models.Income
	if err := common_models.GORM.Where("user_email = ?", email).Order("received_date DESC").Find(&incomes).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incomes)
}

func HandleCreateIncome(w http.ResponseWriter, r *http.Request) {
	var income common_models.Income
	if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if income.UserEmail == "" {
		http.Error(w, "User email missing", http.StatusBadRequest)
		return
	}

	if err := common_models.GORM.Create(&income).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(income)
}

func HandleDeleteIncome(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID missing", http.StatusBadRequest)
		return
	}

	if err := common_models.GORM.Delete(&common_models.Income{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
