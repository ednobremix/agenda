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
	defer models.DB.Close()

	// Configurar Rotas da API (Controller)
	http.HandleFunc("POST /api/register", controllers.HandleRegister)
	http.HandleFunc("POST /api/login", controllers.HandleLogin)
	http.HandleFunc("POST /api/auth/google", controllers.HandleGoogleAuth)
	http.HandleFunc("GET /api/config", controllers.HandleGetConfig)
	http.HandleFunc("POST /api/config", controllers.HandleSaveConfig)
	http.HandleFunc("GET /api/appointments", controllers.HandleGetAppointments)
	http.HandleFunc("POST /api/appointments", controllers.HandleSaveAppointments)
	http.HandleFunc("GET /api/admin/users", controllers.HandleAdminUsers)
	http.HandleFunc("POST /api/admin/update-permission", controllers.HandleUpdateUserPermission)
	http.HandleFunc("GET /api/me", controllers.HandleMe)
	http.HandleFunc("GET /api/debug/users", controllers.HandleDebugUsers)

	// Servir arquivos estáticos (View)
	// Como os arquivos foram para a pasta 'views', apontamos o FileServer para lá
	fs := http.FileServer(http.Dir("views"))
	http.Handle("/", fs)

	port := ":8080"
	fmt.Printf("Servidor iniciado em http://localhost%s\n", port)

	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
