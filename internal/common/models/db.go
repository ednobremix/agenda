package common_models

import (
	"bufio"
	"database/sql"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *sql.DB
var GORM *gorm.DB

func loadEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Se a variável já estiver definida no ambiente, não sobrescreve
			if os.Getenv(key) == "" {
				os.Setenv(key, val)
			}
		}
	}
}

func InitDB() {
	InitDBWithName("")
}

func InitDBWithName(dbName string) {
	loadEnv()
	dsn := os.Getenv("DATABASE_URL")
	var err error
	if dsn != "" {
		log.Println("Conectando ao PostgreSQL via DATABASE_URL...")
		GORM, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		sqliteFile := dbName
		if sqliteFile == "" {
			sqliteFile = os.Getenv("SQLITE_DB")
		}
		if sqliteFile == "" {
			sqliteFile = "sqlite.db"
		}
		log.Printf("DATABASE_URL não definida. Usando SQLite local (%s)...\n", sqliteFile)
		GORM, err = gorm.Open(sqlite.Open(sqliteFile), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Expor o *sql.DB em models.DB para compatibilidade com código legado
	DB, err = GORM.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	if GORM.Dialector.Name() == "sqlite" {
		DB.SetMaxOpenConns(1)
		if err := GORM.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
			log.Printf("Aviso: erro ao ativar WAL no SQLite: %v", err)
		}
		if err := GORM.Exec("PRAGMA busy_timeout=5000;").Error; err != nil {
			log.Printf("Aviso: erro ao configurar busy_timeout no SQLite: %v", err)
		}
	}

	// Corrige bancos antigos onde appointments.id foi criado como bigint
	if GORM.Dialector.Name() == "postgres" {
		fixAppointmentIDColumn()
	}

	// Automigrate irá criar as tabelas se elas não existirem e atualizar o esquema
	err = GORM.AutoMigrate(
		&User{},
		&ConfigStore{},
		&Appointment{},
		&UserPayment{},
		&Expense{},
		&Income{},
		&Bank{},
		&PaymentType{},
		&Product{},
		&ProductBatch{},
		&StockMovement{},
		&PracaSlide{},
		&PracaGalleryConfig{},
		&PracaRoom{},
		&PracaArtwork{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	log.Println("Database connection established and migration completed.")

	// Criar usuário admin se não existir (Exemplo de semente/seed)
	seedAdmin()

	// Garantir que todos os produtos tenham pelo menos um lote (para produtos legados)
	FixLegacyProducts()
}

func FixLegacyProducts() {
	res := GORM.Exec(`
		INSERT INTO product_batches (created_at, product_id, quantity, value, production_cost)
		SELECT ?, p.id, p.quantity, p.value, p.production_cost
		FROM products p
		WHERE p.quantity > 0
		  AND p.deleted_at IS NULL
		  AND NOT EXISTS (
			SELECT 1 FROM product_batches pb WHERE pb.product_id = p.id
		  )
	`, time.Now())
	if res.Error != nil {
		log.Printf("Erro ao corrigir lotes de produtos legados: %v", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		log.Printf("Lotes iniciais criados em lote para %d produto(s) legado(s).", res.RowsAffected)
	}
}

func seedAdmin() {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "edynobre@gmail.com"
	}

	var admin User
	result := GORM.Where("email = ?", adminEmail).First(&admin)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		admin = User{
			Name:     "Administrador",
			Business: "Sistema",
			Email:    adminEmail,
			Role:     "admin",
		}
		if err := GORM.Create(&admin).Error; err != nil {
			log.Printf("Erro ao criar admin: %v", err)
		} else {
			log.Println("Usuário admin criado com sucesso.")
		}
	} else if admin.Role != "admin" {
		GORM.Model(&admin).Update("role", "admin")
		log.Println("Role de administrador verificada/atualizada.")
	}
}

// fixAppointmentIDColumn converte appointments.id de bigint para varchar(100),
// pois os IDs são gerados no frontend (UUID).
func fixAppointmentIDColumn() {
	var dataType string
	GORM.Raw(`SELECT data_type FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'appointments' AND column_name = 'id'`).Scan(&dataType)
	if dataType != "bigint" && dataType != "integer" {
		return
	}
	err := GORM.Exec(`ALTER TABLE appointments ALTER COLUMN id DROP DEFAULT,
		ALTER COLUMN id TYPE varchar(100) USING id::text`).Error
	if err != nil {
		log.Printf("Erro ao converter appointments.id para varchar: %v", err)
		return
	}
	log.Println("appointments.id convertido de bigint para varchar(100).")
}
