package services

import (
	common_admin "massage-scheduler/internal/common/admin"
	"massage-scheduler/internal/common/models"
)

func GetAllUsers() ([]common_models.User, error) {
	return common_admin.GetAllUsers()
}

func UpdateUserAccess(email string, hasAgenda, hasFinanceiro, hasEstoque bool) error {
	return common_admin.UpdateUserAccess(email, hasAgenda, hasFinanceiro, hasEstoque)
}

func GetPayments() ([]common_models.UserPayment, error) {
	return common_admin.GetPayments()
}

func CreatePayment(payment *common_models.UserPayment) error {
	return common_admin.CreatePayment(payment)
}

func UpdatePaymentStatus(id any, status string) error {
	return common_admin.UpdatePaymentStatus(id, status)
}

func UpdateUserPermission(email string, maxProfessionals int, freeScheduling *bool) error {
	updates := map[string]any{
		"max_professionals": maxProfessionals,
	}

	if freeScheduling != nil {
		updates["free_scheduling"] = *freeScheduling
	}

	return common_models.GORM.Model(&common_models.User{}).Where("email = ?", email).Updates(updates).Error
}

func DeleteUser(email, currentUserEmail string) error {
	return common_admin.DeleteUser(email, currentUserEmail)
}
