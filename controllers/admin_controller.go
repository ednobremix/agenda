package controllers

import (
	"encoding/json"
	"massage-scheduler/models"
	"net/http"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	if err := models.GORM.Order("created_at DESC").Find(&users).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func HandleGetPayments(w http.ResponseWriter, r *http.Request) {
	var payments []models.UserPayment
	if err := models.GORM.Preload("User").Order("due_date DESC").Find(&payments).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

func HandleCreatePayment(w http.ResponseWriter, r *http.Request) {
	var payment models.UserPayment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "Erro ao decodificar JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if payment.UserID == 0 {
		http.Error(w, "ID de usuário inválido", http.StatusBadRequest)
		return
	}

	if err := models.GORM.Create(&payment).Error; err != nil {
		http.Error(w, "Erro ao salvar no banco: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func HandleUpdatePayment(w http.ResponseWriter, r *http.Request) {
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
		updates["paid_at"] = new(time.Time)
		*updates["paid_at"].(*time.Time) = time.Now()
	} else {
		updates["paid_at"] = nil
	}

	if err := models.GORM.Model(&models.UserPayment{}).Where("id = ?", req.ID).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleUpdateUserPermission(w http.ResponseWriter, r *http.Request) {
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

	if err := models.GORM.Model(&models.User{}).Where("email = ?", req.Email).Updates(updates).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	currentUserEmail := r.URL.Query().Get("admin_email")

	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	if email == "adm@teste.com" {
		http.Error(w, "Não é possível excluir o administrador padrão", http.StatusBadRequest)
		return
	}

	if email == currentUserEmail {
		http.Error(w, "Você não pode excluir a sua própria conta através do painel administrativo", http.StatusBadRequest)
		return
	}

	// Iniciamos uma transação para garantir a integridade da exclusão de todos os dados
	err := models.GORM.Transaction(func(tx *gorm.DB) error {
		// 1. Encontrar o usuário para obter o ID
		var user models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			return err
		}

		// 2. Excluir dados vinculados pelo E-mail (ConfigStore e Appointment)
		// Usamos Unscoped() para exclusão física completa se desejado, ou mantemos soft delete.
		// O requisito diz "excluir todos os dados", então usaremos exclusão física para garantir.
		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&models.ConfigStore{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&models.Appointment{}).Error; err != nil {
			return err
		}

		// 3. Excluir o usuário e seus relacionamentos dependentes (Photo, Post, etc)
		// Por causa do constraint:OnDelete:CASCADE definido nos modelos e o fato de usarmos GORM,
		// precisamos usar Select() para GORM lidar com as associações ou deletar fisicamente o User
		// O Unscoped().Delete() no GORM para um objeto com associações configuradas com CASCADE no banco
		// funcionará se o banco suportar. Caso contrário, o GORM pode precisar de ajuda.
		// Vamos usar Select(clause.Associations) para garantir que o GORM tente limpar.
		// Mas como adicionamos CASCADE no nível de banco via tags GORM (AutoMigrate), o banco deve cuidar disso.
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
