package routers

import (
	"massage-scheduler/internal/apps/estoque/handlers"
	"massage-scheduler/internal/common/auth"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	requireEst := func(next http.HandlerFunc) http.HandlerFunc {
		return auth.RequireAppAccess(auth.AppEstoque, next)
	}

	mux.HandleFunc("GET /estoque/api/products", requireEst(handlers.HandleGetProducts))
	mux.HandleFunc("POST /estoque/api/products", requireEst(handlers.HandleCreateProduct))
	mux.HandleFunc("PUT /estoque/api/products", requireEst(handlers.HandleUpdateProduct))
	mux.HandleFunc("DELETE /estoque/api/products", requireEst(handlers.HandleDeleteProduct))
	mux.HandleFunc("POST /estoque/api/movement", requireEst(handlers.HandleStockMovement))
	mux.HandleFunc("GET /estoque/api/movements", requireEst(handlers.HandleGetMovements))
	mux.HandleFunc("GET /estoque/api/stats", requireEst(handlers.HandleGetStockStats))
}
