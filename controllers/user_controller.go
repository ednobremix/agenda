package controllers

import (
	"database/sql"
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

	res, err := models.DB.Exec("INSERT INTO users (name, business, email, password, role) VALUES (?, ?, ?, ?, 'user')", u.Name, u.Business, u.Email, u.Password)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			http.Error(w, "E-mail já cadastrado", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := res.LastInsertId()
	u.ID = int(id)
	u.Role = "user"
	models.DB.QueryRow("SELECT created_at FROM users WHERE id = ?", u.ID).Scan(&u.CreatedAt)

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
	err := models.DB.QueryRow("SELECT id, name, business, email, password, role, max_professionals, created_at FROM users WHERE email = ? AND password = ?", credentials.Email, credentials.Password).Scan(&u.ID, &u.Name, &u.Business, &u.Email, &u.Password, &u.Role, &u.MaxProfessionals, &u.CreatedAt)
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
	err = models.DB.QueryRow("SELECT id, name, business, email, password, role, max_professionals, created_at FROM users WHERE email = ?", claims.Email).Scan(&u.ID, &u.Name, &u.Business, &u.Email, &u.Password, &u.Role, &u.MaxProfessionals, &u.CreatedAt)

	if err == sql.ErrNoRows {
		res, err := models.DB.Exec("INSERT INTO users (name, business, email, password, role) VALUES (?, ?, ?, ?, 'user')", claims.Name, "Google User", claims.Email, "google-auth-"+claims.Sub)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		u.ID = int(id)
		u.Name = claims.Name
		u.Business = "Google User"
		u.Email = claims.Email
		u.Role = "user"
		u.MaxProfessionals = 1
		models.DB.QueryRow("SELECT created_at FROM users WHERE id = ?", u.ID).Scan(&u.CreatedAt)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
	err := models.DB.QueryRow("SELECT id, name, business, email, role, max_professionals, created_at FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Name, &user.Business, &user.Email, &user.Role, &user.MaxProfessionals, &user.CreatedAt)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
