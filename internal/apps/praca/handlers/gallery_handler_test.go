package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"massage-scheduler/internal/apps/praca/models"
	"massage-scheduler/internal/common/auth"
	common_models "massage-scheduler/internal/common/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGallery3DAdminAndVisitorAccess(t *testing.T) {
	setupTestDB(t)

	regularUser := common_models.User{
		Name:     "Visitante",
		Username: "visitante_gal",
		Email:    "visitor@example.com",
		GoogleID: "google-visitor-gal",
		Role:     "user",
	}
	if err := common_models.GORM.Create(&regularUser).Error; err != nil {
		t.Fatalf("failed to create regular user: %v", err)
	}

	slide := common_models.PracaSlide{
		Title:     "Obra Teste",
		MimeType:  "image/png",
		ImageData: makeTestPNG(t),
	}
	if err := common_models.GORM.Create(&slide).Error; err != nil {
		t.Fatalf("failed to seed slide: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /praca/api/gallery", HandleGetGallery)
	mux.HandleFunc("PUT /praca/api/gallery/config", auth.RequireAdmin(HandleSaveGalleryConfig))
	mux.HandleFunc("POST /praca/api/gallery/rooms", auth.RequireAdmin(HandleCreateRoom))
	mux.HandleFunc("PUT /praca/api/gallery/rooms", auth.RequireAdmin(HandleUpdateRoom))
	mux.HandleFunc("DELETE /praca/api/gallery/rooms", auth.RequireAdmin(HandleDeleteRoom))
	mux.HandleFunc("POST /praca/api/gallery/artworks", auth.RequireAdmin(HandleCreateArtwork))
	mux.HandleFunc("PUT /praca/api/gallery/artworks", auth.RequireAdmin(HandleUpdateArtwork))
	mux.HandleFunc("DELETE /praca/api/gallery/artworks", auth.RequireAdmin(HandleDeleteArtwork))

	// 1. Visitor can read gallery without auth
	req := httptest.NewRequest(http.MethodGet, "/praca/api/gallery", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on public gallery, got %d", rec.Code)
	}

	// 2. Visitor cannot update config (401)
	cfgBody, _ := json.Marshal(models.GalleryConfigDTO{Title: "Hack", Template: "dark"})
	req = httptest.NewRequest(http.MethodPut, "/praca/api/gallery/config", bytes.NewReader(cfgBody))
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on unauthenticated config, got %d", rec.Code)
	}

	// 3. Regular user cannot create artwork (403)
	userCookie := makeSessionCookie(t, "visitor@example.com", "user")
	artBody, _ := json.Marshal(models.ArtworkDTO{SlideID: slide.ID, Title: "Teste"})
	req = httptest.NewRequest(http.MethodPost, "/praca/api/gallery/artworks", bytes.NewReader(artBody))
	req.AddCookie(userCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin artwork create, got %d", rec.Code)
	}

	// 4. Admin updates config
	adminCookie := makeSessionCookie(t, "edynobre@gmail.com", "admin")
	cfgBody, _ = json.Marshal(models.GalleryConfigDTO{
		Title:    "Exposição Primavera",
		Subtitle: "Bem-vindos ao museu",
		Template: "modern",
	})
	req = httptest.NewRequest(http.MethodPut, "/praca/api/gallery/config", bytes.NewReader(cfgBody))
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on admin config update, got %d: %s", rec.Code, rec.Body.String())
	}

	// 5. Admin creates artwork (and javascript: URL is sanitized)
	artBody, _ = json.Marshal(models.ArtworkDTO{
		SlideID:      slide.ID,
		Wall:         "east",
		PosX:         40,
		PosY:         50,
		Scale:        "large",
		FrameStyle:   "dark-wood",
		Title:        "Aurora",
		Artist:       "Ed Nobre",
		Year:         "2026",
		Technique:    "Óleo sobre tela",
		Dimensions:   "80 x 60 cm",
		Description:  "Descrição da obra",
		ExternalLink: "javascript:alert(1)",
	})
	req = httptest.NewRequest(http.MethodPost, "/praca/api/gallery/artworks", bytes.NewReader(artBody))
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on admin artwork create, got %d: %s", rec.Code, rec.Body.String())
	}
	var created models.ArtworkDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created: %v", err)
	}
	if created.ExternalLink != "" {
		t.Fatalf("expected javascript: link to be stripped, got %q", created.ExternalLink)
	}
	if created.Wall != "east" || created.FrameStyle != "dark-wood" {
		t.Fatalf("unexpected artwork fields: %+v", created)
	}

	// 6. Admin updates artwork with valid https link and custom width_px
	created.ExternalLink = "https://example.com/obra"
	created.ExternalLinkLabel = "Catálogo"
	created.WidthPx = 320
	updBody, _ := json.Marshal(created)
	req = httptest.NewRequest(http.MethodPut, "/praca/api/gallery/artworks", bytes.NewReader(updBody))
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on artwork update, got %d", rec.Code)
	}

	// 7. Visitor reads updated gallery
	req = httptest.NewRequest(http.MethodGet, "/praca/api/gallery", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	var gal models.GalleryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &gal); err != nil {
		t.Fatalf("decode gallery: %v", err)
	}
	if gal.Config.Template != "modern" || len(gal.Artworks) != 1 || gal.Artworks[0].ExternalLink != "https://example.com/obra" || gal.Artworks[0].WidthPx != 320 {
		t.Fatalf("unexpected gallery response: %+v", gal)
	}

	// 8. Admin creates a second room, links doorway, and deletes it
	roomBody, _ := json.Marshal(models.RoomDTO{
		Name:            "Ala Moderna",
		Template:        "dark",
		Width:           1600,
		Depth:           1000,
		Height:          650,
		DoorSouthRoomID: 1,
	})
	req = httptest.NewRequest(http.MethodPost, "/praca/api/gallery/rooms", bytes.NewReader(roomBody))
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 on room create, got %d: %s", rec.Code, rec.Body.String())
	}
	var createdRoom models.RoomDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &createdRoom); err != nil {
		t.Fatalf("decode room: %v", err)
	}
	if createdRoom.Width != 1600 || createdRoom.DoorSouthRoomID != 1 {
		t.Fatalf("unexpected room: %+v", createdRoom)
	}

	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/praca/api/gallery/rooms?id=%d", createdRoom.ID), nil)
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on room delete, got %d", rec.Code)
	}

	// 9. Admin deletes artwork
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/praca/api/gallery/artworks?id=%d", created.ID), nil)
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on artwork delete, got %d", rec.Code)
	}
}
