package routers

import (
	"massage-scheduler/internal/apps/agenda/handlers"
	"massage-scheduler/internal/common/auth"
	"net/http"
)

func RegisterAgendaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/agenda/api/register", handlers.HandleRegister)
	mux.HandleFunc("/agenda/api/login", handlers.HandleLogin)
	mux.HandleFunc("/agenda/api/auth/google", handlers.HandleGoogleAuth)
	mux.HandleFunc("/agenda/api/logout", handlers.HandleLogout)
	mux.HandleFunc("/agenda/api/config", auth.RequireAppAccess(auth.AppAgenda, handlers.HandleGetConfig))
	mux.HandleFunc("POST /agenda/api/config", auth.RequireAppAccess(auth.AppAgenda, handlers.HandleSaveConfig))
	mux.HandleFunc("/agenda/api/appointments", auth.RequireAppAccess(auth.AppAgenda, handlers.HandleGetAppointments))
	mux.HandleFunc("POST /agenda/api/appointments", auth.RequireAppAccess(auth.AppAgenda, handlers.HandleSaveAppointments))
	mux.HandleFunc("/agenda/api/admin/users", auth.RequireAdmin(handlers.HandleAdminUsers))
	mux.HandleFunc("DELETE /agenda/api/admin/users", auth.RequireAdmin(handlers.HandleDeleteUser))
	mux.HandleFunc("/agenda/api/admin/update-permission", auth.RequireAdmin(handlers.HandleUpdateUserPermission))
	mux.HandleFunc("/agenda/api/admin/update-access", auth.RequireAdmin(handlers.HandleUpdateUserAccess))
	mux.HandleFunc("/agenda/api/admin/payments", auth.RequireAdmin(handlers.HandleGetPayments))
	mux.HandleFunc("POST /agenda/api/admin/payments", auth.RequireAdmin(handlers.HandleCreatePayment))
	mux.HandleFunc("PUT /agenda/api/admin/payments", auth.RequireAdmin(handlers.HandleUpdatePayment))
	mux.HandleFunc("/agenda/api/me", auth.RequireAuth(handlers.HandleMe))
	mux.HandleFunc("/agenda/api/user/update-password", auth.RequireAuth(handlers.HandleUpdatePassword))
	mux.HandleFunc("/agenda/api/user/link-google", auth.RequireAuth(handlers.HandleLinkGoogle))
}
