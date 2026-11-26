package services

import (
	"errors"
	"massage-scheduler/internal/apps/financeiro/models"
	common_models "massage-scheduler/internal/common/models"
	"gorm.io/gorm"
)

var ErrRecordNotFound = errors.New("User not found")

func UpdateUserPermission(input models.UpdateUserPermissionInput) error {
	var user common_models.User
	if err := common_models.GORM.First(&user, input.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRecordNotFound
		}
		return err
	}

	user.HasFinanceiro = input.HasFinanceiro
	user.HasEstoque = input.HasEstoque
	user.Role = input.Role

	return common_models.GORM.Save(&user).Error
}

func FindUserByEmail(email string) (common_models.User, error) {
	var user common_models.User
	if err := common_models.GORM.Where("email = ?", email).First(&user).Error; err != nil {
		return common_models.User{}, err
	}
	return user, nil
}
