package controllers

import (
	"encoding/json"
	"massage-scheduler/models"
	"net/http"
)

func HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := models.DB.Query("SELECT id, name, business, email, password, role, max_professionals, created_at FROM users ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Business, &u.Email, &u.Password, &u.Role, &u.MaxProfessionals, &u.CreatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func HandleUpdateUserPermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email            string `json:"email"`
		MaxProfessionals int    `json:"max_professionals"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := models.DB.Exec("UPDATE users SET max_professionals = ? WHERE email = ?", req.MaxProfessionals, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func HandleDebugUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := models.DB.Query("SELECT email, password FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type debugUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var users []debugUser
	for rows.Next() {
		var u debugUser
		if err := rows.Scan(&u.Email, &u.Password); err != nil {
			continue
		}
		users = append(users, u)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
