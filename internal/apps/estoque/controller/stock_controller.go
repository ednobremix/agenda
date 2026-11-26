package controller

import (
	"errors"
	"massage-scheduler/internal/apps/estoque/models"
	"massage-scheduler/internal/apps/estoque/services"
	common_models "massage-scheduler/internal/common/models"
)

var (
	ErrUserNotFound = errors.New("Usuário não encontrado")
	ErrAccessDenied = errors.New("Acesso negado ao Controle de Estoque")
)

func ValidateUserAccess(email string) error {
	var user common_models.User
	if err := common_models.GORM.Where("email = ?", email).First(&user).Error; err != nil {
		return ErrUserNotFound
	}

	if user.Role != "admin" && !user.HasEstoque {
		return ErrAccessDenied
	}

	return nil
}

func GetProducts(email string) ([]common_models.Product, error) {
	return services.GetProducts(email)
}

func CreateProduct(input common_models.Product) (common_models.Product, error) {
	return services.CreateProduct(input)
}

func UpdateProduct(product common_models.Product) (common_models.Product, error) {
	return services.UpdateProduct(product)
}

func DeleteProduct(id uint, email string) error {
	return services.DeleteProduct(id, email)
}

func RegisterMovement(movement common_models.StockMovement) error {
	return services.StockMovement(movement)
}

func GetMovements(email, productID, startDate, endDate string) ([]common_models.StockMovement, error) {
	return services.GetMovements(email, productID, startDate, endDate)
}

func GetStockStats(email, startDate, endDate string) (models.StockStats, error) {
	return services.GetStockStats(email, startDate, endDate)
}
