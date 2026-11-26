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
	HasEstoque       bool           `json:"has_estoque" gorm:"default:false"`
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

// Product representa um produto no estoque
type Product struct {
	ID             uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt      time.Time      `json:"created_at,omitzero"`
	UpdatedAt      time.Time      `json:"updated_at,omitzero"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
	UserEmail      string         `json:"user_email" gorm:"index"`
	Name           string         `json:"name"`
	Quantity       float64        `json:"quantity"` // Quantidade total somando todos os lotes
	MinQuantity    float64        `json:"min_quantity"`
	Type           string         `json:"type"`  // "comprado" ou "criado"
	Value          float64        `json:"value"` // Valor do lote mais recente ou custo médio (opcional)
	ProductionCost float64        `json:"production_cost"`
	Location       string         `json:"location"`
	Batches        []ProductBatch `json:"batches" gorm:"foreignKey:ProductID"`
}

// ProductBatch representa um lote de produto com um valor específico
type ProductBatch struct {
	ID             uint      `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt      time.Time `json:"created_at,omitzero"`
	ProductID      uint      `json:"product_id" gorm:"index"`
	Quantity       float64   `json:"quantity"`
	Value          float64   `json:"value"`           // Valor pago (para comprados)
	ProductionCost float64   `json:"production_cost"` // Custo de produção (para criados)
}

// StockMovement representa a entrada ou saída de um produto
type StockMovement struct {
	ID         uint          `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt  time.Time     `json:"created_at,omitzero"`
	ProductID  uint          `json:"product_id" gorm:"index"`
	Product    Product       `json:"product"`
	BatchID    *uint         `json:"batch_id" gorm:"index"` // Lote afetado (opcional)
	Batch      *ProductBatch `json:"batch" gorm:"foreignKey:BatchID"`
	UserEmail  string        `json:"user_email" gorm:"index"`
	Type       string        `json:"type"` // "entrada", "saida", "venda", "baixa", "cadastro"
	Quantity   float64       `json:"quantity"`
	SaleValue  float64       `json:"sale_value"`
	BatchValue float64       `json:"batch_value"` // Valor do novo lote (para entrada)
	Reason     string        `json:"reason"`
}

// PracaSlide representa uma foto exibida no expositor de slides da Praça
type PracaSlide struct {
	ID        uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at,omitzero"`
	UpdatedAt time.Time      `json:"updated_at,omitzero"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	Title     string         `json:"title"`
	Caption   string         `json:"caption"`
	MimeType  string         `json:"mime_type" gorm:"size:100;not null"`
	ImageData []byte         `json:"-" gorm:"not null"`
}

// PracaGalleryConfig representa a configuração do espaço 3D da Galeria/Museu Virtual
type PracaGalleryConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title" gorm:"default:'Museu Virtual da Praça'"`
	Subtitle  string    `json:"subtitle" gorm:"default:'Explore as obras pelas paredes da galeria 3D'"`
	Template  string    `json:"template" gorm:"default:'classic'"` // classic, modern, dark
}

// PracaRoom representa uma sala modelável do Museu Virtual
type PracaRoom struct {
	ID              uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt       time.Time      `json:"created_at,omitzero"`
	UpdatedAt       time.Time      `json:"updated_at,omitzero"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
	Name            string         `json:"name" gorm:"not null"`
	Description     string         `json:"description"`
	Template        string         `json:"template" gorm:"size:20;default:'classic'"`
	Width           int            `json:"width" gorm:"default:1200"`
	Depth           int            `json:"depth" gorm:"default:1200"`
	Height          int            `json:"height" gorm:"default:600"`
	SortOrder       int            `json:"sort_order" gorm:"default:0"`
	DoorNorthRoomID uint           `json:"door_north_room_id" gorm:"default:0"`
	DoorEastRoomID  uint           `json:"door_east_room_id" gorm:"default:0"`
	DoorSouthRoomID uint           `json:"door_south_room_id" gorm:"default:0"`
	DoorWestRoomID  uint           `json:"door_west_room_id" gorm:"default:0"`
	DoorNorthPosX   float64        `json:"door_north_pos_x" gorm:"default:82"`
	DoorEastPosX    float64        `json:"door_east_pos_x" gorm:"default:82"`
	DoorSouthPosX   float64        `json:"door_south_pos_x" gorm:"default:82"`
	DoorWestPosX    float64        `json:"door_west_pos_x" gorm:"default:82"`
}

// PracaArtwork representa uma obra posicionada em uma parede da galeria 3D
type PracaArtwork struct {
	ID                uint           `json:"id,omitzero" gorm:"primaryKey"`
	CreatedAt         time.Time      `json:"created_at,omitzero"`
	UpdatedAt         time.Time      `json:"updated_at,omitzero"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
	RoomID            uint           `json:"room_id" gorm:"index;default:1"`
	SlideID           uint           `json:"slide_id" gorm:"index;not null"`
	Wall              string         `json:"wall" gorm:"size:20;default:'north'"`       // north, east, south, west
	PosX              float64        `json:"pos_x" gorm:"default:50"`                   // 10..90 (%)
	PosY              float64        `json:"pos_y" gorm:"default:50"`                   // 25..75 (%)
	Scale             string         `json:"scale" gorm:"size:20;default:'medium'"`     // small, medium, large
	WidthPx           float64        `json:"width_px" gorm:"default:0"`                 // 0 = usa Scale; >0 = largura customizada (100..440)
	FrameStyle        string         `json:"frame_style" gorm:"size:30;default:'classic-gold'"`
	Title             string         `json:"title"`
	Artist            string         `json:"artist"`
	Year              string         `json:"year"`
	Technique         string         `json:"technique"`
	Dimensions        string         `json:"dimensions"`
	Description       string         `json:"description" gorm:"type:text"`
	ExternalLink      string         `json:"external_link"`
	ExternalLinkLabel string         `json:"external_link_label"`
}


