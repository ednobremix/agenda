package main

import (
	"fmt"
	"log"
	"net/http"

	"massage-scheduler/controllers"
	"massage-scheduler/models"
)

func main() {
	// Inicializar Banco de Dados (Model)
	models.InitDB()

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

	// Servir arquivos estáticos (View)
	// Como os arquivos foram para a pasta 'views', apontamos o FileServer para lá
	fs := http.FileServer(http.Dir("views"))
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers para evitar erros de COOP no Google Auth
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
		// w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp") // Opcional, dependendo da necessidade

		// Se for uma requisição de API, não deve cair aqui, mas por precaução garantimos cabeçalhos
		fs.ServeHTTP(w, r)
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
