package routers

import (
	"massage-scheduler/internal/apps/agenda/handlers"
	"net/http"
)

func RegisterAgendaRoutes() {
	http.HandleFunc("POST /api/register", handlers.HandleRegister)
	http.HandleFunc("POST /api/login", handlers.HandleLogin)
	http.HandleFunc("POST /api/auth/google", handlers.HandleGoogleAuth)
	http.HandleFunc("GET /api/config", handlers.HandleGetConfig)
	http.HandleFunc("POST /api/config", handlers.HandleSaveConfig)
	http.HandleFunc("GET /api/appointments", handlers.HandleGetAppointments)
	http.HandleFunc("POST /api/appointments", handlers.HandleSaveAppointments)
	http.HandleFunc("GET /api/admin/users", handlers.HandleAdminUsers)
	http.HandleFunc("DELETE /api/admin/users", handlers.HandleDeleteUser)
	http.HandleFunc("POST /api/admin/update-permission", handlers.HandleUpdateUserPermission)
	http.HandleFunc("POST /api/admin/update-access", handlers.HandleUpdateUserAccess)
	http.HandleFunc("GET /api/admin/payments", handlers.HandleGetPayments)
	http.HandleFunc("POST /api/admin/payments", handlers.HandleCreatePayment)
	http.HandleFunc("PUT /api/admin/payments", handlers.HandleUpdatePayment)
	http.HandleFunc("GET /api/me", handlers.HandleMe)
	http.HandleFunc("POST /api/user/update-password", handlers.HandleUpdatePassword)
	http.HandleFunc("POST /api/user/link-google", handlers.HandleLinkGoogle)
}
