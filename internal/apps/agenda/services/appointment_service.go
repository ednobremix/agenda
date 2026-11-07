package services

import (
	"errors"
	"massage-scheduler/internal/common/models"
	"time"

	"gorm.io/gorm"
)

func GetAppointments(email string) ([]common_models.Appointment, error) {
	var apps []common_models.Appointment
	if err := common_models.GORM.Where("user_email = ?", email).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func SaveAppointments(email string, apps []common_models.Appointment) error {
	return common_models.GORM.Transaction(func(tx *gorm.DB) error {
		var user common_models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			return err
		}

		if user.Role != "admin" && !user.HasAgenda {
			return errors.New("acesso negado ao App Agenda")
		}

		if !user.FreeScheduling {
			newAppsByDate := make(map[string]int)
			for _, app := range apps {
				if len(app.DataHora) >= 10 {
					date := app.DataHora[:10]
					newAppsByDate[date]++
					if newAppsByDate[date] > 1 {
						return gorm.ErrInvalidData
					}
				}
			}
		}

		var existingApps []common_models.Appointment
		if err := tx.Where("user_email = ?", email).Find(&existingApps).Error; err != nil {
			return err
		}

		newAppIDs := make(map[string]bool)
		for _, app := range apps {
			if app.ID != "" {
				newAppIDs[app.ID] = true
			}
		}

		now := time.Now()
		for _, oldApp := range existingApps {
			if _, exists := newAppIDs[oldApp.ID]; !exists {
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

		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&common_models.Appointment{}).Error; err != nil {
			return err
		}

		for i := range apps {
			apps[i].UserEmail = email
			if apps[i].ID == "" {
				apps[i].ID = time.Now().Format("20060102150405") + "-" + email
			}
		}

		if len(apps) > 0 {
			if err := tx.CreateInBatches(&apps, 100).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
