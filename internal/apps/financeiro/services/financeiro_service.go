package services

import (
	"fmt"
	common_models "massage-scheduler/internal/common/models"
)

func GetExpenses(email string) ([]common_models.Expense, error) {
	var expenses []common_models.Expense
	if err := common_models.GORM.Where("user_email = ?", email).Order("purchase_date DESC").Find(&expenses).Error; err != nil {
		return nil, err
	}
	return expenses, nil
}

func CreateExpense(expense common_models.Expense) (common_models.Expense, error) {
	var user common_models.User
	if err := common_models.GORM.Where("email = ?", expense.UserEmail).First(&user).Error; err != nil {
		return common_models.Expense{}, fmt.Errorf("user not found: %w", err)
	}

	installments := expense.Installments
	if installments <= 0 {
		installments = 1
	}

	months := []string{"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}
	startMonthIdx := -1
	for i, m := range months {
		if m == expense.PaymentMonth {
			startMonthIdx = i
			break
		}
	}
	if startMonthIdx == -1 {
		startMonthIdx = 0
	}

	originalDescription := expense.Description
	baseDate := expense.PurchaseDate

	for i := 0; i < installments; i++ {
		currentExpense := expense
		currentExpense.ID = 0
		if installments > 1 {
			currentExpense.Description = fmt.Sprintf("%s (%d/%d)", originalDescription, i+1, installments)
		}
		monthIdx := (startMonthIdx + i) % 12
		currentExpense.PaymentMonth = months[monthIdx]
		currentExpense.PurchaseDate = baseDate.AddDate(0, i, 0)

		if err := common_models.GORM.Create(&currentExpense).Error; err != nil {
			return common_models.Expense{}, err
		}
	}

	return expense, nil
}

func DeleteExpense(id, userEmail string) error {
	var count int64
	common_models.GORM.Model(&common_models.Income{}).Where("expense_id = ? AND user_email = ?", id, userEmail).Count(&count)
	if count > 0 {
		return fmt.Errorf("não é possível excluir um empréstimo que já possui recebimento vinculado")
	}

	res := common_models.GORM.Where("id = ? AND user_email = ?", id, userEmail).Delete(&common_models.Expense{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func GetIncomes(email string) ([]common_models.Income, error) {
	var incomes []common_models.Income
	if err := common_models.GORM.Where("user_email = ?", email).Order("received_date DESC").Find(&incomes).Error; err != nil {
		return nil, err
	}
	return incomes, nil
}

func CreateIncome(income common_models.Income) (common_models.Income, error) {
	if err := common_models.GORM.Create(&income).Error; err != nil {
		return common_models.Income{}, err
	}
	return income, nil
}

func DeleteIncome(id, userEmail string) error {
	res := common_models.GORM.Where("id = ? AND user_email = ?", id, userEmail).Delete(&common_models.Income{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func GetBanks(email string) ([]common_models.Bank, error) {
	var banks []common_models.Bank
	err := common_models.GORM.Where("user_email = ?", email).Find(&banks).Error
	return banks, err
}

func CreateBank(bank common_models.Bank) (common_models.Bank, error) {
	err := common_models.GORM.Create(&bank).Error
	return bank, err
}

func DeleteBank(id, userEmail string) error {
	res := common_models.GORM.Where("id = ? AND user_email = ?", id, userEmail).Delete(&common_models.Bank{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func GetPaymentTypes(email string) ([]common_models.PaymentType, error) {
	var types []common_models.PaymentType
	err := common_models.GORM.Where("user_email = ?", email).Find(&types).Error
	return types, err
}

func CreatePaymentType(pt common_models.PaymentType) (common_models.PaymentType, error) {
	err := common_models.GORM.Create(&pt).Error
	return pt, err
}

func DeletePaymentType(id, userEmail string) error {
	res := common_models.GORM.Where("id = ? AND user_email = ?", id, userEmail).Delete(&common_models.PaymentType{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
