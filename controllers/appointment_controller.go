package controllers

import (
	"encoding/json"
	"massage-scheduler/models"
	"net/http"
)

func HandleGetAppointments(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	rows, err := models.DB.Query("SELECT id, paciente, massoterapeuta, tipo, valor, data_hora, pago, duracao FROM appointments WHERE user_email = ?", email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var apps []models.Appointment
	for rows.Next() {
		var a models.Appointment
		if err := rows.Scan(&a.ID, &a.Paciente, &a.Massoterapeuta, &a.Tipo, &a.Valor, &a.DataHora, &a.Pago, &a.Duracao); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		apps = append(apps, a)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}

func HandleSaveAppointments(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	var apps []models.Appointment
	if err := json.NewDecoder(r.Body).Decode(&apps); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := models.DB.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM appointments WHERE user_email = ?", email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, a := range apps {
		_, err = tx.Exec(`INSERT INTO appointments (id, user_email, paciente, massoterapeuta, tipo, valor, data_hora, pago, duracao) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			a.ID, email, a.Paciente, a.Massoterapeuta, a.Tipo, a.Valor, a.DataHora, a.Pago, a.Duracao)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
