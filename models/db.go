package models

import (
	"bufio"
	"database/sql"
	"log"
	"os"
	"strings"

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
	loadEnv()
	dsn := os.Getenv("DATABASE_URL")
	var err error
	if dsn != "" {
		log.Println("Conectando ao PostgreSQL via DATABASE_URL...")
		GORM, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		sqliteFile := os.Getenv("SQLITE_DB")
		if sqliteFile == "" {
			sqliteFile = "sqlite.db"
		}
		log.Printf("DATABASE_URL não definida. Usando SQLite local (%s) como fallback...\n", sqliteFile)
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
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate: %v", err)
	}

	log.Println("Database connection established and migration completed.")

	// Criar usuário admin se não existir (Exemplo de semente/seed)
	seedAdmin()
}

func seedAdmin() {
	var admin User
	result := GORM.Where("email = ?", "edynobre@gmail.com").First(&admin)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		admin = User{
			Name:     "Administrador",
			Business: "Sistema",
			Email:    "edynobre@gmail.com",
			Password: "q1w2e3", // Em produção, use hash de senha!
			Role:     "admin",
		}
		if err := GORM.Create(&admin).Error; err != nil {
			log.Printf("Erro ao criar admin: %v", err)
		} else {
			log.Println("Usuário admin criado com sucesso.")
		}
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
