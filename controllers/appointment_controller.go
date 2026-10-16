package controllers

import (
	"encoding/json"
	"errors"
	"massage-scheduler/models"
	"net/http"
	"time"

	"gorm.io/gorm"
)

func HandleGetAppointments(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var apps []models.Appointment
	if err := models.GORM.Where("user_email = ?", email).Find(&apps).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

func HandleSaveAppointments(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var apps []models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&apps); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := models.GORM.Transaction(func(tx *gorm.DB) error {
		// Verificar se o usuário tem agendamento livre ou restrição de 1 por dia
		var user models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			return err
		}

		if !user.FreeScheduling {
			// Agrupar novos agendamentos por data (YYYY-MM-DD)
			newAppsByDate := make(map[string]int)
			for _, app := range apps {
				if len(app.DataHora) >= 10 {
					date := app.DataHora[:10]
					newAppsByDate[date]++
					if newAppsByDate[date] > 1 {
						return gorm.ErrInvalidData // Simplificando erro para transação, trataremos fora
					}
				}
			}
		}

		// Limpar agendamentos existentes.
		// Nota: Usamos Unscoped() para realmente deletar se for o caso,
		// mas aqui estamos apenas limpando para recriar a lista completa vinda do frontend.
		// No entanto, não permitimos o cancelamento de sessões que já passaram e foram pagas.
		// Precisamos garantir que elas permaneçam se existirem.

		var existingApps []models.Appointment
		if err := tx.Where("user_email = ?", email).Find(&existingApps).Error; err != nil {
			return err
		}

		// Identificar quais IDs estão sendo removidos
		newAppIDs := make(map[string]bool)
		for _, app := range apps {
			if app.ID != "" {
				newAppIDs[app.ID] = true
			}
		}

		now := time.Now()
		for _, oldApp := range existingApps {
			if _, exists := newAppIDs[oldApp.ID]; !exists {
				// Este agendamento está sendo removido.
				// Verificar se ele já passou e foi pago.
				// A dataHora no banco é string, pode vir em ISO8601 (JS) ou custom "YYYY-MM-DDTHH:MM"
				var appTime time.Time
				appTime, err := time.Parse(time.RFC3339, oldApp.DataHora)
				if err != nil {
					appTime, err = time.Parse("2006-01-02T15:04", oldApp.DataHora)
				}

				if err == nil {
					if oldApp.Pago && appTime.Before(now) {
						return errors.New("não é permitido cancelar sessões que já passaram e foram pagas")
					}
				}
			}
		}

		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&models.Appointment{}).Error; err != nil {
			return err
		}

		for i := range apps {
			apps[i].UserEmail = email
			// Garantir que agendamentos novos tenham ID se por acaso vierem sem
			if apps[i].ID == "" {
				// Gerar um ID simples se estiver vazio (fallback)
				apps[i].ID = time.Now().Format("20060102150405") + "-" + email
			}
		}

		if len(apps) > 0 {
			// Usamos CreateInBatches para eficiência se houver muitos agendamentos
			if err := tx.CreateInBatches(&apps, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrInvalidData) {
			http.Error(w, "Agendamento livre desativado: limite de 1 agendamento por dia excedido.", http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
