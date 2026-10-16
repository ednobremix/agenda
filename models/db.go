package models

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "./agenda.db")
	if err != nil {
		log.Fatal(err)
	}

	// Criar tabelas
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			business TEXT,
			email TEXT UNIQUE,
			password TEXT,
			role TEXT DEFAULT 'user',
			max_professionals INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS configs (
			user_email TEXT PRIMARY KEY,
			data TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS appointments (
			id TEXT PRIMARY KEY,
			user_email TEXT,
			paciente TEXT,
			massoterapeuta TEXT,
			tipo TEXT,
			valor REAL,
			data_hora TEXT,
			pago BOOLEAN,
			duracao INTEGER
		);`,
	}

	for _, q := range queries {
		_, err = DB.Exec(q)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Adicionar colunas se não existirem (para DBs legados)
	_, _ = DB.Exec("ALTER TABLE users ADD COLUMN allow_multiple_professionals BOOLEAN DEFAULT 0")
	_, _ = DB.Exec("ALTER TABLE users ADD COLUMN max_professionals INTEGER DEFAULT 1")

	// Criar usuário admin se não existir
	var adminExists bool
	err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = 'adm@teste.com')").Scan(&adminExists)
	if err != nil {
		log.Printf("Erro ao verificar admin: %v", err)
	} else if !adminExists {
		_, err = DB.Exec("INSERT INTO users (name, business, email, password, role) VALUES (?, ?, ?, ?, ?)",
			"Administrador", "Sistema", "adm@teste.com", "q1w2e3", "admin")
		if err != nil {
			log.Printf("Erro ao criar admin: %v", err)
		} else {
			log.Println("Usuário admin criado com sucesso.")
		}
	}
}
