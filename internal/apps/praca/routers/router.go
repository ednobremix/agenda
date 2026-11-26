package routers

import (
	"massage-scheduler/internal/apps/praca/handlers"
	"massage-scheduler/internal/common/auth"
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /praca/api/me", handlers.HandleMe)
	mux.HandleFunc("GET /praca/api/slides", handlers.HandleListSlides)
	mux.HandleFunc("GET /praca/api/slides/image", handlers.HandleGetSlideImage)
	mux.HandleFunc("POST /praca/api/slides", auth.RequireAdmin(handlers.HandleCreateSlide))
	mux.HandleFunc("DELETE /praca/api/slides", auth.RequireAdmin(handlers.HandleDeleteSlide))

	mux.HandleFunc("GET /praca/api/gallery", handlers.HandleGetGallery)
	mux.HandleFunc("PUT /praca/api/gallery/config", auth.RequireAdmin(handlers.HandleSaveGalleryConfig))
	mux.HandleFunc("POST /praca/api/gallery/rooms", auth.RequireAdmin(handlers.HandleCreateRoom))
	mux.HandleFunc("PUT /praca/api/gallery/rooms", auth.RequireAdmin(handlers.HandleUpdateRoom))
	mux.HandleFunc("DELETE /praca/api/gallery/rooms", auth.RequireAdmin(handlers.HandleDeleteRoom))
	mux.HandleFunc("POST /praca/api/gallery/artworks", auth.RequireAdmin(handlers.HandleCreateArtwork))
	mux.HandleFunc("PUT /praca/api/gallery/artworks", auth.RequireAdmin(handlers.HandleUpdateArtwork))
	mux.HandleFunc("DELETE /praca/api/gallery/artworks", auth.RequireAdmin(handlers.HandleDeleteArtwork))
}
