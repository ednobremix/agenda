package routers

import (
	"massage-scheduler/internal/apps/financeiro/handlers"
	"massage-scheduler/internal/common/auth"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	requireFin := func(next http.HandlerFunc) http.HandlerFunc {
		return auth.RequireAppAccess(auth.AppFinanceiro, next)
	}

	// API Gastos
	mux.HandleFunc("GET /financeiro/api/expenses", requireFin(handlers.HandleGetExpenses))
	mux.HandleFunc("POST /financeiro/api/expenses", requireFin(handlers.HandleCreateExpense))
	mux.HandleFunc("DELETE /financeiro/api/expenses", requireFin(handlers.HandleDeleteExpense))

	// API Recebimentos
	mux.HandleFunc("GET /financeiro/api/incomes", requireFin(handlers.HandleGetIncomes))
	mux.HandleFunc("POST /financeiro/api/incomes", requireFin(handlers.HandleCreateIncome))
	mux.HandleFunc("DELETE /financeiro/api/incomes", requireFin(handlers.HandleDeleteIncome))

	// API Bancos
	mux.HandleFunc("GET /financeiro/api/banks", requireFin(handlers.GetBanks))
	mux.HandleFunc("POST /financeiro/api/banks", requireFin(handlers.CreateBank))
	mux.HandleFunc("DELETE /financeiro/api/banks", requireFin(handlers.DeleteBank))

	// API Tipos de Pagamento
	mux.HandleFunc("GET /financeiro/api/payment-types", requireFin(handlers.GetPaymentTypes))
	mux.HandleFunc("POST /financeiro/api/payment-types", requireFin(handlers.CreatePaymentType))
	mux.HandleFunc("DELETE /financeiro/api/payment-types", requireFin(handlers.DeletePaymentType))

	// Admin
	mux.HandleFunc("GET /financeiro/api/admin/users", auth.RequireAdmin(handlers.HandleFinanceiroAdminUsers))
	mux.HandleFunc("GET /financeiro/api/admin/payments", auth.RequireAdmin(handlers.HandleFinanceiroGetPayments))
	mux.HandleFunc("POST /financeiro/api/admin/payments", auth.RequireAdmin(handlers.HandleFinanceiroCreatePayment))
	mux.HandleFunc("PUT /financeiro/api/admin/payments", auth.RequireAdmin(handlers.HandleFinanceiroUpdatePayment))
	mux.HandleFunc("POST /financeiro/api/admin/user/permissions", auth.RequireAdmin(handlers.HandleFinanceiroUpdateUserPermission))
	mux.HandleFunc("POST /financeiro/api/admin/user/access", auth.RequireAdmin(handlers.HandleFinanceiroUpdateUserAccess))
	mux.HandleFunc("DELETE /financeiro/api/admin/user/delete", auth.RequireAdmin(handlers.HandleFinanceiroDeleteUser))
}
