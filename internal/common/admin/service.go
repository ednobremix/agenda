package admin

import (
	"errors"
	common_models "massage-scheduler/internal/common/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAllUsers() ([]common_models.User, error) {
	var users []common_models.User
	if err := common_models.GORM.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func UpdateUserAccess(email string, hasAgenda, hasFinanceiro, hasEstoque bool) error {
	return common_models.GORM.Model(&common_models.User{}).Where("email = ?", email).
		Updates(map[string]any{
			"has_agenda":     hasAgenda,
			"has_financeiro": hasFinanceiro,
			"has_estoque":    hasEstoque,
		}).Error
}

func GetPayments() ([]common_models.UserPayment, error) {
	var payments []common_models.UserPayment
	if err := common_models.GORM.Preload("User").Order("due_date DESC").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func CreatePayment(payment *common_models.UserPayment) error {
	if payment.UserID == 0 {
		return errors.New("ID de usuário inválido")
	}
	return common_models.GORM.Create(payment).Error
}

func UpdatePaymentStatus(id any, status string) error {
	updates := map[string]any{"status": status}
	if status == "paid" {
		now := time.Now()
		updates["paid_at"] = &now
	} else {
		updates["paid_at"] = nil
	}

	return common_models.GORM.Model(&common_models.UserPayment{}).Where("id = ?", id).Updates(updates).Error
}

func DeleteUser(email, currentUserEmail string) error {
	if email == "" {
		return errors.New("email missing")
	}

	if email == "edynobre@gmail.com" {
		return errors.New("não é possível excluir o administrador padrão")
	}

	if email == currentUserEmail {
		return errors.New("você não pode excluir a sua própria conta através do painel administrativo")
	}

	return common_models.GORM.Transaction(func(tx *gorm.DB) error {
		var user common_models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&common_models.ConfigStore{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("user_email = ?", email).Delete(&common_models.Appointment{}).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Select(clause.Associations).Delete(&user).Error; err != nil {
			return err
		}

		return nil
	})
}
