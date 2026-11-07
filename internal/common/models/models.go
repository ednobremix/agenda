package common_models

import (
	"time"

	"gorm.io/gorm"
)

// User representa o usuário do sistema
type User struct {
	ID               uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt        time.Time      `json:"created_at,omitzero"`
	UpdatedAt        time.Time      `json:"updated_at,omitzero"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
	Name             string         `json:"name"`
	Business         string         `json:"business"`
	Username         string         `json:"username" gorm:"uniqueIndex"`
	Email            string         `json:"email" gorm:"uniqueIndex;not null"`
	Password         string         `json:"-"`
	Role             string         `json:"role" gorm:"default:'user'"`
	MaxProfessionals int            `json:"max_professionals" gorm:"default:1"`
	IsPremium        bool           `json:"is_premium" gorm:"default:false"`
	IsPrivate        bool           `json:"is_private" gorm:"default:false"`
	GoogleID         string         `json:"google_id" gorm:"uniqueIndex;type:varchar(100)"`
	FreeScheduling   bool           `json:"free_scheduling" gorm:"default:true"`
	HasAgenda        bool           `json:"has_agenda" gorm:"default:false"`
	HasFinanceiro    bool           `json:"has_financeiro" gorm:"default:false"`
}

// Estruturas antigas para manter compatibilidade com controllers existentes (se necessário)
// Nota: Em uma migração real, os controllers também seriam atualizados.

// ConfigStore representa as configurações salvas no banco
type ConfigStore struct {
	UserEmail string `gorm:"primaryKey"`
	Data      string `gorm:"type:text"`
}

// Appointment representa um agendamento no sistema
type Appointment struct {
	ID             string `json:"id" gorm:"primaryKey;type:varchar(100)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	UserEmail      string         `json:"user_email" gorm:"index"`
	Cliente        string         `json:"cliente"`
	Massoterapeuta string         `json:"massoterapeuta"`
	Tipo           string         `json:"tipo"`
	Valor          float64        `json:"valor"`
	DataHora       string         `json:"dataHora"`
	Pago           bool           `json:"pago"`
	Duracao        int            `json:"duracao"`
}

// UserPayment representa um pagamento de um usuário (mensalidade/cobrança)
type UserPayment struct {
	ID        uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at,omitzero"`
	UpdatedAt time.Time      `json:"updated_at,omitzero"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UserID    uint           `json:"user_id" gorm:"index"`
	User      User           `json:"user"`
	Amount    float64        `json:"amount"`
	DueDate   time.Time      `json:"due_date"`
	PaidAt    *time.Time     `json:"paid_at"`
	Status    string         `json:"status" gorm:"default:'pending'"` // pending, paid, overdue
	Reference string         `json:"reference"`                       // Ex: "Mensalidade Outubro"
}

// Expense representa um gasto cadastrado pelo usuário
type Expense struct {
	ID           uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt    time.Time      `json:"created_at,omitzero"`
	UpdatedAt    time.Time      `json:"updated_at,omitzero"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
	UserEmail    string         `json:"user_email" gorm:"index"`
	Description  string         `json:"description"`   // Compra/Local
	PaymentMonth string         `json:"payment_month"` // Mês de pagamento
	Installments int            `json:"installments"`  // Número de parcelas
	PurchaseDate time.Time      `json:"purchase_date"` // Data da compra
	Amount       float64        `json:"amount"`        // Valor da parcela
	PaymentType  string         `json:"payment_type"`  // Tipo de pagamento
	Category     string         `json:"category"`      // Compra ou Empréstimo
}

// Income representa um recebimento cadastrado pelo usuário
type Income struct {
	ID            uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt     time.Time      `json:"created_at,omitzero"`
	UpdatedAt     time.Time      `json:"updated_at,omitzero"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	UserEmail     string         `json:"user_email" gorm:"index"`
	Source        string         `json:"source"`         // Procedência
	ReceivedMonth string         `json:"received_month"` // Mês recebido
	ReceivedDate  time.Time      `json:"received_date"`  // Data de recebimento
	Amount        float64        `json:"amount"`         // Valor
	Bank          string         `json:"bank"`           // Banco recebido
	Category      string         `json:"category"`       // Categoria (opcional)
	ExpenseID     *uint          `json:"expense_id"`     // ID do gasto (empréstimo) vinculado
}

// Bank representa um banco cadastrado pelo usuário
type Bank struct {
	ID        uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at,omitzero"`
	UpdatedAt time.Time      `json:"updated_at,omitzero"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UserEmail string         `json:"user_email" gorm:"index"`
	Name      string         `json:"name"`
}

// PaymentType representa um tipo de pagamento cadastrado pelo usuário
type PaymentType struct {
	ID        uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at,omitzero"`
	UpdatedAt time.Time      `json:"updated_at,omitzero"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	UserEmail string         `json:"user_email" gorm:"index"`
	Name      string         `json:"name"`
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
