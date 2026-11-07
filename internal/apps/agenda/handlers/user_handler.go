package handlers

import (
	"encoding/json"
	"massage-scheduler/internal/apps/agenda/models"
	"massage-scheduler/internal/apps/agenda/services"
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

	user, err := services.AuthenticateGoogle(data.Credential)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	user, err := services.GetUserByEmail(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "A alteração de senha manual está desativada. Sua conta é gerida via Google Auth.", http.StatusForbidden)
}

func HandleLinkGoogle(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var data models.LinkGoogleRequest
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := services.LinkGoogleAccount(email, data.Credential)
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
