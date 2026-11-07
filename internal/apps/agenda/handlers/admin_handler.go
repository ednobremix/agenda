package handlers

import (
	"encoding/json"
	"massage-scheduler/internal/apps/agenda/models"
	"massage-scheduler/internal/apps/agenda/services"
	"massage-scheduler/internal/common/models"
	"net/http"
)

func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	users, err := services.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func HandleUpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateUserAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := services.UpdateUserAccess(req.Email, req.HasAgenda, req.HasFinanceiro, req.HasEstoque); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleGetPayments(w http.ResponseWriter, r *http.Request) {
	payments, err := services.GetPayments()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func HandleCreatePayment(w http.ResponseWriter, r *http.Request) {
	var payment common_models.UserPayment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "Erro ao decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := services.CreatePayment(&payment); err != nil {
		if err.Error() == "ID de usuário inválido" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Erro ao salvar no banco: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandleUpdatePayment(w http.ResponseWriter, r *http.Request) {
	var req models.UpdatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := services.UpdatePaymentStatus(req.ID, req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleUpdateUserPermission(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateUserPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := services.UpdateUserPermission(req.Email, req.MaxProfessionals, req.FreeScheduling); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	currentUserEmail := r.URL.Query().Get("admin_email")

	if err := services.DeleteUser(email, currentUserEmail); err != nil {
		if err.Error() == "email missing" ||
			err.Error() == "não é possível excluir o administrador padrão" ||
			err.Error() == "você não pode excluir a sua própria conta através do painel administrativo" {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Erro ao excluir usuário e seus dados: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusOK)
}
