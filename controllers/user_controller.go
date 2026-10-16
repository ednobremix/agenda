package controllers

import (
	"encoding/base64"
	"encoding/json"
	"massage-scheduler/models"
	"net/http"
	"strings"
)

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u.Role = "user"
	if err := models.GORM.Create(&u).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "E-mail ou Usuário já cadastrado", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var u models.User
	err := models.GORM.Where("email = ? AND password = ?", credentials.Email, credentials.Password).First(&u).Error
	if err != nil {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(u)
}

func HandleGoogleAuth(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Credential string `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parts := strings.Split(data.Credential, ".")
	if len(parts) < 2 {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	payloadPart := parts[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		http.Error(w, "Error decoding token: "+err.Error(), http.StatusBadRequest)
		return
	}

	var claims struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Sub   string `json:"sub"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		http.Error(w, "Error parsing token claims", http.StatusBadRequest)
		return
	}

	if claims.Email == "" {
		http.Error(w, "Email not found in token", http.StatusBadRequest)
		return
	}

	var u models.User
	err = models.GORM.Where("google_id = ?", claims.Sub).First(&u).Error

	if err != nil {
		// Se não achar por GoogleID, tenta por e-mail
		err = models.GORM.Where("email = ?", claims.Email).First(&u).Error
		if err == nil {
			// Vincula o GoogleID se a conta existir mas não estiver vinculada
			u.GoogleID = claims.Sub
			models.GORM.Save(&u)
		} else {
			// Cria novo usuário
			u = models.User{
				Name:             claims.Name,
				Business:         "Google User",
				Email:            claims.Email,
				Username:         claims.Email, // Usar e-mail como username padrão
				Password:         "google-auth-" + claims.Sub,
				Role:             "user",
				MaxProfessionals: 1,
				GoogleID:         claims.Sub,
			}
			if err := models.GORM.Create(&u).Error; err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	json.NewEncoder(w).Encode(u)
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email missing", http.StatusBadRequest)
		return
	}

	var user models.User
	err := models.GORM.Where("email = ?", email).First(&user).Error

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var data struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var u models.User
	err := models.GORM.Where("email = ? AND password = ?", email, data.CurrentPassword).First(&u).Error
	if err != nil {
		http.Error(w, "Senha atual incorreta", http.StatusUnauthorized)
		return
	}

	u.Password = data.NewPassword
	if err := models.GORM.Save(&u).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleLinkGoogle(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var data struct {
		Credential string `json:"credential"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	parts := strings.Split(data.Credential, ".")
	if len(parts) < 2 {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	payloadPart := parts[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		http.Error(w, "Error decoding token", http.StatusBadRequest)
		return
	}

	var claims struct {
		Email string `json:"email"`
		Sub   string `json:"sub"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		http.Error(w, "Error parsing token claims", http.StatusBadRequest)
		return
	}

	var u models.User
	if err := models.GORM.Where("email = ?", email).First(&u).Error; err != nil {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	u.GoogleID = claims.Sub
	if claims.Email != "" {
		u.Email = claims.Email
		u.Username = claims.Email
	}

	if err := models.GORM.Save(&u).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "Esta conta Google já está vinculada a outro usuário", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
