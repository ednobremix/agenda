package handlers

import (
	"errors"
	"massage-scheduler/internal/apps/financeiro/controller"
	"massage-scheduler/internal/apps/financeiro/decode"
	"massage-scheduler/internal/common/auth"
	common_utils "massage-scheduler/internal/common/utils"
	"net/http"
)

func HandleFinanceiroAdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := controller.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = common_utils.WriteJSON(w, http.StatusOK, users)
}

func HandleFinanceiroGetPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := controller.GetPayments()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = common_utils.WriteJSON(w, http.StatusOK, payments)
}

func HandleFinanceiroCreatePayment(w http.ResponseWriter, r *http.Request) {
	payment, err := decode.DecodePayment(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := controller.CreatePayment(&payment); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandleFinanceiroUpdatePayment(w http.ResponseWriter, r *http.Request) {
	input, err := decode.DecodeUpdatePayment(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := controller.UpdatePaymentStatus(input.ID, input.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleFinanceiroUpdateUserPermission(w http.ResponseWriter, r *http.Request) {
	input, err := decode.DecodeUpdateUserPermission(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := controller.UpdateUserPermission(input); err != nil {
		if errors.Is(err, controller.ErrTargetUserNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleFinanceiroUpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	input, err := decode.DecodeUpdateUserAccess(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := controller.UpdateUserAccess(input); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleFinanceiroDeleteUser(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}
	email, _, err := decode.ParseDeleteUserParams(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := controller.DeleteUser(email, currentUser.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
