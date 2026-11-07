package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"massage-scheduler/internal/apps/agenda/controllers"
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
	http.HandleFunc("GET /api/admin/payments", controllers.HandleGetPayments)
	http.HandleFunc("POST /api/admin/payments", controllers.HandleCreatePayment)
	http.HandleFunc("PUT /api/admin/payments", controllers.HandleUpdatePayment)
	http.HandleFunc("GET /api/me", controllers.HandleMe)
	http.HandleFunc("POST /api/user/update-password", controllers.HandleUpdatePassword)
	http.HandleFunc("POST /api/user/link-google", controllers.HandleLinkGoogle)

	// Servir arquivos estáticos da Agenda
	agendaFS := http.FileServer(http.Dir("internal/apps/agenda/views"))
	http.Handle("/agenda/", http.StripPrefix("/agenda/", agendaFS))

	// Servir arquivos estáticos do Financeiro (se a pasta existir)
	financeiroPath := "../financeiro/frontend"
	if _, err := os.Stat(financeiroPath); err == nil {
		financeiroFS := http.FileServer(http.Dir(financeiroPath))
		http.Handle("/financeiro/", http.StripPrefix("/financeiro/", financeiroFS))
	}

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
