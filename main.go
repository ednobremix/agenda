package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	agenda_routers "massage-scheduler/internal/apps/agenda/routers"
	estoque_routers "massage-scheduler/internal/apps/estoque/routers"
	financeiro_routers "massage-scheduler/internal/apps/financeiro/routers"
	praca_routers "massage-scheduler/internal/apps/praca/routers"
	"massage-scheduler/internal/common/models"
	common_utils "massage-scheduler/internal/common/utils"
)

func main() {
	// Inicializar Banco de Dados (Model)
	common_models.InitDB()

	// Configurar Rotas das APIs
	mux := http.DefaultServeMux
	agenda_routers.RegisterAgendaRoutes(mux)
	financeiro_routers.RegisterRoutes(mux)
	estoque_routers.RegisterRoutes(mux)
	praca_routers.RegisterRoutes(mux)

	// Pré-instanciar servidores de arquivos estáticos
	agendaFS := http.FileServer(http.Dir("internal/apps/agenda/views"))
	financeiroFS := http.FileServer(http.Dir("internal/apps/financeiro/views"))
	estoqueFS := http.FileServer(http.Dir("internal/apps/estoque/views"))
	pracaFS := http.FileServer(http.Dir("internal/apps/praca/views"))
	homeFS := http.FileServer(http.Dir("internal/apps/home/views"))

	http.Handle("/agenda/", http.StripPrefix("/agenda/", agendaFS))
	http.Handle("/financeiro/", http.StripPrefix("/financeiro/", financeiroFS))
	http.Handle("/estoque/", http.StripPrefix("/estoque/", estoqueFS))
	http.HandleFunc("/praca/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/praca/" {
			http.Redirect(w, r, "/praca/html/index.html", http.StatusMovedPermanently)
			return
		}
		http.StripPrefix("/praca/", pracaFS).ServeHTTP(w, r)
	})

	// Redirecionamentos de rotas legadas
	redirects := map[string]string{
		"/admin.html":            "/agenda/html/admin.html",
		"/ajuda.html":            "/agenda/html/ajuda.html",
		"/login.html":            "/agenda/html/login.html",
		"/agenda/login.html":     "/agenda/html/login.html",
		"/agenda/admin.html":     "/agenda/html/admin.html",
		"/agenda/ajuda.html":     "/agenda/html/ajuda.html",
		"/agenda/index.html":     "/agenda/html/index.html",
		"/financeiro/login.html": "/financeiro/html/login.html",
		"/financeiro/admin.html": "/financeiro/html/admin.html",
		"/financeiro/ajuda.html": "/financeiro/html/ajuda.html",
		"/financeiro/index.html": "/financeiro/html/index.html",
		"/estoque/login.html":    "/estoque/html/login.html",
		"/estoque/ajuda.html":    "/estoque/html/ajuda.html",
		"/estoque/index.html":    "/estoque/html/index.html",
		"/praca":                 "/praca/html/index.html",
		"/praca/index.html":      "/praca/html/index.html",
		"/praca/ajuda.html":      "/praca/html/ajuda.html",
		"/praca/slides.html":     "/praca/html/slides.html",
		"/praca/galeria3d.html":  "/praca/html/galeria3d.html",
	}
	for from, to := range redirects {
		target := to
		http.HandleFunc(from, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		})
	}

	// Servir arquivos estáticos da Home e fallback
	staticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers para evitar erros de COOP no Google Auth
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")

		switch r.URL.Path {
		case "/manifest.json":
			http.ServeFile(w, r, "internal/apps/agenda/views/manifest.json")
			return
		case "/sw.js":
			http.ServeFile(w, r, "internal/apps/agenda/views/js/sw.js")
			return
		case "/":
			http.ServeFile(w, r, "internal/apps/home/views/html/home.html")
			return
		}

		if _, err := os.Stat("internal/apps/home/views" + r.URL.Path); err == nil {
			homeFS.ServeHTTP(w, r)
			return
		}
		if _, err := os.Stat("internal/apps/financeiro/views" + r.URL.Path); err == nil {
			financeiroFS.ServeHTTP(w, r)
			return
		}
		if _, err := os.Stat("internal/apps/estoque/views" + r.URL.Path); err == nil {
			estoqueFS.ServeHTTP(w, r)
			return
		}
		if _, err := os.Stat("internal/apps/praca/views" + r.URL.Path); err == nil {
			pracaFS.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})
	http.Handle("/", staticHandler)

	// Middleware global de cabeçalhos de segurança HTTP
	securityHeadersMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cross-Origin-Opener-Policy", "same-origin-allow-popups")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' https://accounts.google.com; style-src 'self' 'unsafe-inline' https://accounts.google.com; frame-src 'self' https://accounts.google.com; connect-src 'self' https://accounts.google.com; img-src 'self' data: https:;")
			next.ServeHTTP(w, r)
		})
	}

	server := &http.Server{
		Addr:              ":8080",
		Handler:           securityHeadersMiddleware(http.DefaultServeMux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("Servidor iniciado em http://localhost%s\n", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()
	log.Println("Encerrando servidor...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Erro no shutdown do servidor HTTP: %v", err)
	}
	common_utils.ShutdownNotifications(shutdownCtx)
	log.Println("Servidor encerrado com segurança.")
}
