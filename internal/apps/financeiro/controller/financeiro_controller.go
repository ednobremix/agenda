package controller

import (
	"errors"
	"massage-scheduler/internal/apps/financeiro/models"
	"massage-scheduler/internal/apps/financeiro/services"
	common_admin "massage-scheduler/internal/common/admin"
	common_models "massage-scheduler/internal/common/models"
)

var (
	ErrUserNotFound       = errors.New("Usuário não encontrado")
	ErrAccessDenied       = errors.New("Acesso negado ao Financeiro")
	ErrTargetUserNotFound = services.ErrRecordNotFound
)

func ValidateUserAccess(email string) error {
	user, err := services.FindUserByEmail(email)
	if err != nil {
		return ErrUserNotFound
	}
	if user.Role != "admin" && !user.HasFinanceiro {
		return ErrAccessDenied
	}
	return nil
}

func GetExpenses(email string) ([]common_models.Expense, error) {
	return services.GetExpenses(email)
}

func CreateExpense(expense common_models.Expense) (common_models.Expense, error) {
	return services.CreateExpense(expense)
}

func DeleteExpense(id, userEmail string) error {
	return services.DeleteExpense(id, userEmail)
}

func GetIncomes(email string) ([]common_models.Income, error) {
	return services.GetIncomes(email)
}

func CreateIncome(income common_models.Income) (common_models.Income, error) {
	return services.CreateIncome(income)
}

func DeleteIncome(id, userEmail string) error {
	return services.DeleteIncome(id, userEmail)
}

func GetBanks(email string) ([]common_models.Bank, error) {
	return services.GetBanks(email)
}

func CreateBank(bank common_models.Bank) (common_models.Bank, error) {
	return services.CreateBank(bank)
}

func DeleteBank(id, userEmail string) error {
	return services.DeleteBank(id, userEmail)
}

func GetPaymentTypes(email string) ([]common_models.PaymentType, error) {
	return services.GetPaymentTypes(email)
}

func CreatePaymentType(pt common_models.PaymentType) (common_models.PaymentType, error) {
	return services.CreatePaymentType(pt)
}

func DeletePaymentType(id, userEmail string) error {
	return services.DeletePaymentType(id, userEmail)
}

func GetAllUsers() ([]common_models.User, error) {
	return common_admin.GetAllUsers()
}

func GetPayments() ([]common_models.UserPayment, error) {
	return common_admin.GetPayments()
}

func CreatePayment(payment *common_models.UserPayment) error {
	return common_admin.CreatePayment(payment)
}

func UpdatePaymentStatus(id uint, status string) error {
	return common_admin.UpdatePaymentStatus(id, status)
}

func UpdateUserPermission(input models.UpdateUserPermissionInput) error {
	return services.UpdateUserPermission(input)
}

func UpdateUserAccess(input models.UpdateUserAccessInput) error {
	return common_admin.UpdateUserAccess(input.Email, input.HasAgenda, input.HasFinanceiro, input.HasEstoque)
}

func DeleteUser(email, currentUserEmail string) error {
	return common_admin.DeleteUser(email, currentUserEmail)
}
