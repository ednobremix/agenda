package handlers

import (
	"encoding/json"
	"massage-scheduler/internal/apps/agenda/models"
	"massage-scheduler/internal/apps/agenda/services"
	"massage-scheduler/internal/common/auth"
	"net/http"
)

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "O cadastro manual está desativado. Utilize o login via Google na página inicial.", http.StatusForbidden)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "O login manual está desativado. Utilize o login via Google.", http.StatusForbidden)
}

func HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	var data models.GoogleAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := services.AuthenticateGoogle(r.Context(), data.Credential)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := auth.SetSessionCookie(w, r, user.Email, user.Role); err != nil {
		http.Error(w, "Erro ao criar sessão", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func HandleLogout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w, r)
	w.WriteHeader(http.StatusOK)
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "A alteração de senha manual está desativada. Sua conta é gerida via Google Auth.", http.StatusForbidden)
}

func HandleLinkGoogle(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		http.Error(w, "Não autenticado", http.StatusUnauthorized)
		return
	}

	var data models.LinkGoogleRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := services.LinkGoogleAccount(r.Context(), user.Email, data.Credential)
	if err != nil {
		if err.Error() == "usuário não encontrado" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if err.Error() == "esta conta Google já está vinculada a outro usuário" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
