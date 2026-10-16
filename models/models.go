package models

type User struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Business         string `json:"business"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	Role             string `json:"role"`
	MaxProfessionals int    `json:"max_professionals"`
	CreatedAt        string `json:"created_at"`
}

type Service struct {
	Name     string  `json:"name"`
	Value    float64 `json:"value"`
	Duration int     `json:"duration"`
}

type Config struct {
	BusinessType  string    `json:"businessType"`
	Professionals []string  `json:"professionals"`
	Services      []Service `json:"services"`
}

type Appointment struct {
	ID             string  `json:"id"`
	Paciente       string  `json:"paciente"`
	Massoterapeuta string  `json:"massoterapeuta"`
	Tipo           string  `json:"tipo"`
	Valor          float64 `json:"valor"`
	DataHora       string  `json:"dataHora"`
	Pago           bool    `json:"pago"`
	Duracao        int     `json:"duracao"`
}
