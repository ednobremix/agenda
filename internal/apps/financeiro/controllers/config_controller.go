package controllers

import (
	"encoding/json"
	"massage-scheduler/internal/common/models"
	"net/http"
)

func GetBanks(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var banks []common_models.Bank
	common_models.GORM.Where("user_email = ?", email).Find(&banks)
	json.NewEncoder(w).Encode(banks)
}

func CreateBank(w http.ResponseWriter, r *http.Request) {
	var bank common_models.Bank
	if err := json.NewDecoder(r.Body).Decode(&bank); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	common_models.GORM.Create(&bank)
	json.NewEncoder(w).Encode(bank)
}

func DeleteBank(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	common_models.GORM.Delete(&common_models.Bank{}, id)
	w.WriteHeader(http.StatusOK)
}

func GetPaymentTypes(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var types []common_models.PaymentType
	common_models.GORM.Where("user_email = ?", email).Find(&types)
	json.NewEncoder(w).Encode(types)
}

func CreatePaymentType(w http.ResponseWriter, r *http.Request) {
	var pt common_models.PaymentType
	if err := json.NewDecoder(r.Body).Decode(&pt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	common_models.GORM.Create(&pt)
	json.NewEncoder(w).Encode(pt)
}

func DeletePaymentType(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	common_models.GORM.Delete(&common_models.PaymentType{}, id)
	w.WriteHeader(http.StatusOK)
}
