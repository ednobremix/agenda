package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"massage-scheduler/internal/apps/agenda/controllers"
	financeiro_controllers "massage-scheduler/internal/apps/financeiro/controllers"
	"massage-scheduler/internal/common/models"
)

func main() {
	// Inicializar Banco de Dados (Model)
	common_models.InitDB()

	// Configurar Rotas da API (Controller)
	http.HandleFunc("POST /api/register", controllers.HandleRegister)
	http.HandleFunc("POST /api/login", controllers.HandleLogin)
	http.HandleFunc("POST /api/auth/google", controllers.HandleGoogleAuth)
	http.HandleFunc("GET /api/config", controllers.HandleGetConfig)
	http.HandleFunc("POST /api/config", controllers.HandleSaveConfig)
	http.HandleFunc("GET /api/appointments", controllers.HandleGetAppointments)
	http.HandleFunc("POST /api/appointments", controllers.HandleSaveAppointments)
	http.HandleFunc("GET /api/admin/users", controllers.HandleAdminUsers)
	http.HandleFunc("DELETE /api/admin/users", controllers.HandleDeleteUser)
	http.HandleFunc("POST /api/admin/update-permission", controllers.HandleUpdateUserPermission)
	http.HandleFunc("POST /api/admin/update-access", controllers.HandleUpdateUserAccess)
	http.HandleFunc("GET /api/admin/payments", controllers.HandleGetPayments)
	http.HandleFunc("POST /api/admin/payments", controllers.HandleCreatePayment)
	http.HandleFunc("PUT /api/admin/payments", controllers.HandleUpdatePayment)
	http.HandleFunc("GET /api/me", controllers.HandleMe)
	http.HandleFunc("POST /api/user/update-password", controllers.HandleUpdatePassword)
	http.HandleFunc("POST /api/user/link-google", controllers.HandleLinkGoogle)

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

	// Para manter compatibilidade com links diretos como /agenda.html
	http.HandleFunc("/agenda.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "internal/apps/agenda/views/agenda.html")
	})
	// Servir outros arquivos da agenda que podem estar na raiz da visualização (js, css)
	// Como o script.js é referenciado como "script.js" no html, ele precisa ser servido.
	// Se acessarmos /agenda.html, ele vai pedir /script.js
	// Vou servir a pasta de views da agenda na raiz se não for outra rota

	// Servir arquivos estáticos da Home e fallback para Agenda
	homeFS := http.FileServer(http.Dir("internal/apps/home/views"))

	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers para evitar erros de COOP no Google Auth
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")

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

		// Se não, tenta na pasta agenda (ex: script.js, agenda.html)
		if _, err := os.Stat("internal/apps/agenda/views" + r.URL.Path); err == nil {
			http.FileServer(http.Dir("internal/apps/agenda/views")).ServeHTTP(w, r)
			return
		}

		// Se não, tenta na pasta financeiro
		if _, err := os.Stat("internal/apps/financeiro/views" + r.URL.Path); err == nil {
			http.FileServer(http.Dir("internal/apps/financeiro/views")).ServeHTTP(w, r)
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
