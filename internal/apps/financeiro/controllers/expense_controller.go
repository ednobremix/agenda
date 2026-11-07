package controllers

import (
	"encoding/json"
	"fmt"
	"massage-scheduler/internal/common/models"
	"net/http"
)

func HandleGetExpenses(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	var expenses []common_models.Expense
	if err := common_models.GORM.Where("user_email = ?", email).Order("purchase_date DESC").Find(&expenses).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}

func HandleCreateExpense(w http.ResponseWriter, r *http.Request) {
	var expense common_models.Expense
	if err := json.NewDecoder(r.Body).Decode(&expense); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if expense.UserEmail == "" {
		http.Error(w, "User email missing", http.StatusBadRequest)
		return
	}

	// Verificar permissão
	var user common_models.User
	if err := common_models.GORM.Where("email = ?", expense.UserEmail).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusForbidden)
		return
	}
	if user.Role != "admin" && !user.HasFinanceiro {
		http.Error(w, "Acesso negado ao App Financeiro", http.StatusForbidden)
		return
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

	// Se não achou o mês (erro), usa Janeiro
	if startMonthIdx == -1 {
		startMonthIdx = 0
	}

	originalDescription := expense.Description
	baseDate := expense.PurchaseDate

	for i := 0; i < installments; i++ {
		currentExpense := expense
		currentExpense.ID = 0 // Reset ID for GORM

		// Ajusta descrição se houver mais de uma parcela
		if installments > 1 {
			currentExpense.Description = fmt.Sprintf("%s (%d/%d)", originalDescription, i+1, installments)
		}

		// Ajusta mês de pagamento
		monthIdx := (startMonthIdx + i) % 12
		currentExpense.PaymentMonth = months[monthIdx]

		// Ajusta data de compra (incrementa meses)
		currentExpense.PurchaseDate = baseDate.AddDate(0, i, 0)

		if err := common_models.GORM.Create(&currentExpense).Error; err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(expense)
}

func HandleDeleteExpense(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID missing", http.StatusBadRequest)
		return
	}

	// Verificar se o gasto está vinculado a um recebimento (empréstimo pago)
	var count int64
	common_models.GORM.Model(&common_models.Income{}).Where("expense_id = ?", id).Count(&count)
	if count > 0 {
		http.Error(w, "Não é possível excluir um empréstimo que já possui recebimento vinculado (está pago).", http.StatusForbidden)
		return
	}

	if err := common_models.GORM.Delete(&common_models.Expense{}, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
