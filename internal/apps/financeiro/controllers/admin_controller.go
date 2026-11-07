package controllers

import (
	"encoding/json"
	"massage-scheduler/internal/common/models"
	"net/http"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func HandleFinanceiroAdminUsers(w http.ResponseWriter, r *http.Request) {
	var users []common_models.User
	if err := common_models.GORM.Order("created_at DESC").Find(&users).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func HandleFinanceiroGetPayments(w http.ResponseWriter, r *http.Request) {
	var payments []common_models.UserPayment
	history := r.URL.Query().Get("history")

	query := common_models.GORM.Preload("User")
	if history != "1" {
		query = query.Where("status != ?", "paid")
	}

	if err := query.Order("due_date DESC").Find(&payments).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func HandleFinanceiroCreatePayment(w http.ResponseWriter, r *http.Request) {
	var payment common_models.UserPayment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "Erro ao decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payment.UserID == 0 {
		http.Error(w, "ID de usuário inválido", http.StatusBadRequest)
		return
	}

	payment.Status = "pending"
	if err := common_models.GORM.Create(&payment).Error; err != nil {
		http.Error(w, "Erro ao salvar no banco: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandleFinanceiroUpdatePayment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID     any    `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updates := map[string]any{"status": req.Status}
	if req.Status == "paid" {
		updates["paid_at"] = new(time.Now())
	} else {
		updates["paid_at"] = nil
	}

	if err := common_models.GORM.Model(&common_models.UserPayment{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleFinanceiroUpdateUserPermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email            string `json:"email"`
		MaxProfessionals int    `json:"max_professionals"`
		FreeScheduling   *bool  `json:"free_scheduling"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updates := map[string]any{
		"max_professionals": req.MaxProfessionals,
	}

	if req.FreeScheduling != nil {
		updates["free_scheduling"] = *req.FreeScheduling
	}

	if err := common_models.GORM.Model(&common_models.User{}).Where("email = ?", req.Email).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleFinanceiroUpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email         string `json:"email"`
		HasAgenda     bool   `json:"has_agenda"`
		HasFinanceiro bool   `json:"has_financeiro"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := common_models.GORM.Model(&common_models.User{}).Where("email = ?", req.Email).
		Updates(map[string]any{
			"has_agenda":     req.HasAgenda,
			"has_financeiro": req.HasFinanceiro,
		}).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func HandleFinanceiroDeleteUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	currentUserEmail := r.URL.Query().Get("admin_email")

	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	if email == "edynobre@gmail.com" {
		http.Error(w, "Não é possível excluir o administrador padrão", http.StatusBadRequest)
		return
	}

	if email == currentUserEmail {
		http.Error(w, "Você não pode excluir a sua própria conta através do painel administrativo", http.StatusBadRequest)
		return
	}

	// Iniciamos uma transação para garantir a integridade da exclusão de todos os dados
	err := common_models.GORM.Transaction(func(tx *gorm.DB) error {
		// 1. Encontrar o usuário para obter o ID
		var user common_models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			return err
		}

		// 2. Excluir dados vinculados pelo E-mail (Agenda)
		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&common_models.ConfigStore{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&common_models.Appointment{}).Error; err != nil {
			return err
		}

		// 3. Excluir dados vinculados pelo UserID (Financeiro)
		if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&common_models.Expense{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&common_models.Income{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&common_models.Bank{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&common_models.PaymentType{}).Error; err != nil {
			return err
		}

		// 4. Excluir o usuário e seus relacionamentos dependentes
		if err := tx.Unscoped().Select(clause.Associations).Delete(&user).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Erro ao excluir usuário e seus dados: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
