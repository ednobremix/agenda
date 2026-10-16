package models

import (
	"database/sql"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *sql.DB
var GORM *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	if dsn != "" {
		log.Println("Conectando ao PostgreSQL via DATABASE_URL...")
		GORM, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	} else {
		log.Println("DATABASE_URL não definida. Usando SQLite local (agenda.db) como fallback...")
		GORM, err = gorm.Open(sqlite.Open("agenda.db"), &gorm.Config{})
	}

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Expor o *sql.DB em models.DB para compatibilidade com código legado
	DB, err = GORM.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
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
	result := GORM.Where("email = ?", "adm@teste.com").First(&admin)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		admin = User{
			Name:     "Administrador",
			Business: "Sistema",
			Email:    "adm@teste.com",
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
