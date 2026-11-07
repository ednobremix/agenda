package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"massage-scheduler/internal/apps/agenda/routers"
	estoque_controllers "massage-scheduler/internal/apps/estoque/controllers"
	financeiro_controllers "massage-scheduler/internal/apps/financeiro/controllers"
	"massage-scheduler/internal/common/models"
)

func main() {
	// Inicializar Banco de Dados (Model)
	common_models.InitDB()

	// Configurar Rotas da API Agenda
	routers.RegisterAgendaRoutes()

	// Rotas do Financeiro Admin
	http.HandleFunc("GET /api/financeiro/admin/users", financeiro_controllers.HandleFinanceiroAdminUsers)
	http.HandleFunc("DELETE /api/financeiro/admin/users", financeiro_controllers.HandleFinanceiroDeleteUser)
	http.HandleFunc("POST /api/financeiro/admin/update-permission", financeiro_controllers.HandleFinanceiroUpdateUserPermission)
	http.HandleFunc("POST /api/financeiro/admin/update-access", financeiro_controllers.HandleFinanceiroUpdateUserAccess)
	http.HandleFunc("GET /api/financeiro/admin/payments", financeiro_controllers.HandleFinanceiroGetPayments)
	http.HandleFunc("POST /api/financeiro/admin/payments", financeiro_controllers.HandleFinanceiroCreatePayment)
	http.HandleFunc("PUT /api/financeiro/admin/payments", financeiro_controllers.HandleFinanceiroUpdatePayment)

	// Rotas do Financeiro
	http.HandleFunc("GET /api/financeiro/expenses", financeiro_controllers.HandleGetExpenses)
	http.HandleFunc("POST /api/financeiro/expenses", financeiro_controllers.HandleCreateExpense)
	http.HandleFunc("DELETE /api/financeiro/expenses", financeiro_controllers.HandleDeleteExpense)
	http.HandleFunc("GET /api/financeiro/incomes", financeiro_controllers.HandleGetIncomes)
	http.HandleFunc("POST /api/financeiro/incomes", financeiro_controllers.HandleCreateIncome)
	http.HandleFunc("DELETE /api/financeiro/incomes", financeiro_controllers.HandleDeleteIncome)

	// Rotas do Estoque
	http.HandleFunc("GET /api/estoque/products", estoque_controllers.AuthEstoqueMiddleware(estoque_controllers.HandleGetProducts))
	http.HandleFunc("POST /api/estoque/products", estoque_controllers.AuthEstoqueMiddleware(estoque_controllers.HandleCreateProduct))
	http.HandleFunc("PUT /api/estoque/products", estoque_controllers.AuthEstoqueMiddleware(estoque_controllers.HandleUpdateProduct))
	http.HandleFunc("DELETE /api/estoque/products", estoque_controllers.AuthEstoqueMiddleware(estoque_controllers.HandleDeleteProduct))
	http.HandleFunc("POST /api/estoque/movements", estoque_controllers.AuthEstoqueMiddleware(estoque_controllers.HandleStockMovement))
	http.HandleFunc("GET /api/estoque/movements", estoque_controllers.AuthEstoqueMiddleware(estoque_controllers.HandleGetMovements))
	http.HandleFunc("GET /api/estoque/stats", estoque_controllers.AuthEstoqueMiddleware(estoque_controllers.HandleGetStockStats))

	// Rotas de Configuração do Financeiro
	http.HandleFunc("GET /api/financeiro/banks", financeiro_controllers.GetBanks)
	http.HandleFunc("POST /api/financeiro/banks", financeiro_controllers.CreateBank)
	http.HandleFunc("DELETE /api/financeiro/banks", financeiro_controllers.DeleteBank)
	http.HandleFunc("GET /api/financeiro/payment-types", financeiro_controllers.GetPaymentTypes)
	http.HandleFunc("POST /api/financeiro/payment-types", financeiro_controllers.CreatePaymentType)
	http.HandleFunc("DELETE /api/financeiro/payment-types", financeiro_controllers.DeletePaymentType)

	// Servir arquivos estáticos da Agenda
	agendaFS := http.FileServer(http.Dir("internal/apps/agenda/views"))
	http.Handle("/agenda/", http.StripPrefix("/agenda/", agendaFS))

	// Servir arquivos estáticos do Financeiro
	financeiroFS := http.FileServer(http.Dir("internal/apps/financeiro/views"))
	http.Handle("/financeiro/", http.StripPrefix("/financeiro/", financeiroFS))

	// Servir arquivos estáticos do Estoque
	estoqueFS := http.FileServer(http.Dir("internal/apps/estoque/views"))
	http.Handle("/estoque/", http.StripPrefix("/estoque/", estoqueFS))

	http.HandleFunc("/admin.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/agenda/admin.html", http.StatusMovedPermanently)
	})
	http.HandleFunc("/ajuda.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/agenda/ajuda.html", http.StatusMovedPermanently)
	})
	http.HandleFunc("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/agenda/login.html", http.StatusMovedPermanently)
	})

	// Servir arquivos estáticos da Home e fallback para Agenda
	homeFS := http.FileServer(http.Dir("internal/apps/home/views"))

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers para evitar erros de COOP no Google Auth
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")

		// Servir manifest.json e sw.js na raiz
		if r.URL.Path == "/manifest.json" {
			http.ServeFile(w, r, "internal/apps/agenda/views/manifest.json")
			return
		}
		if r.URL.Path == "/sw.js" {
			http.ServeFile(w, r, "internal/apps/agenda/views/sw.js")
			return
		}

		// Rota raiz "/" serve home.html da pasta home
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "internal/apps/home/views/home.html")
			return
		}

		// Se o arquivo existir na pasta home (ex: style.css da home)
		if _, err := os.Stat("internal/apps/home/views" + r.URL.Path); err == nil {
			homeFS.ServeHTTP(w, r)
			return
		}

		// Se não, tenta na pasta financeiro
		if _, err := os.Stat("internal/apps/financeiro/views" + r.URL.Path); err == nil {
			http.FileServer(http.Dir("internal/apps/financeiro/views")).ServeHTTP(w, r)
			return
		}

		// Se não, tenta na pasta estoque
		if _, err := os.Stat("internal/apps/estoque/views" + r.URL.Path); err == nil {
			http.FileServer(http.Dir("internal/apps/estoque/views")).ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})
	http.Handle("/", staticHandler)

	// Middleware para API
	apiHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
			next.ServeHTTP(w, r)
		})
	}

	server := &http.Server{
		Addr:    ":8080",
		Handler: apiHandler(http.DefaultServeMux),
	}

	fmt.Printf("Servidor iniciado em http://localhost%s\n", server.Addr)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
