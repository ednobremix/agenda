package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Define o diretório que contém os arquivos estáticos (.)
	fs := http.FileServer(http.Dir("."))

	// Configura a rota principal para servir os arquivos
	http.Handle("/", fs)

	port := ":8080"
	fmt.Printf("Servidor iniciado em http://localhost%s\n", port)

	// Inicia o servidor na porta 8080
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
