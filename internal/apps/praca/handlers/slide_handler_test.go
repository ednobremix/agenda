package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"massage-scheduler/internal/apps/praca/models"
	"massage-scheduler/internal/common/auth"
	common_models "massage-scheduler/internal/common/models"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	common_models.InitDBWithName(dbPath)
}

func makeTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 74, G: 124, B: 89, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}
	return buf.Bytes()
}

func makeSessionCookie(t *testing.T, email, role string) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if err := auth.SetSessionCookie(rec, req, email, role); err != nil {
		t.Fatalf("failed to set session cookie: %v", err)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}
	return cookies[0]
}

func TestSlidesAdminAuthAndMultiUpload(t *testing.T) {
	setupTestDB(t)

	// Create a regular user in DB
	regularUser := common_models.User{
		Name:     "Visitante",
		Username: "visitante",
		Email:    "user@example.com",
		GoogleID: "google-visitante",
		Role:     "user",
	}
	if err := common_models.GORM.Create(&regularUser).Error; err != nil {
		t.Fatalf("failed to create regular user: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /praca/api/me", HandleMe)
	mux.HandleFunc("GET /praca/api/slides", HandleListSlides)
	mux.HandleFunc("GET /praca/api/slides/image", HandleGetSlideImage)
	mux.HandleFunc("POST /praca/api/slides", auth.RequireAdmin(HandleCreateSlide))
	mux.HandleFunc("DELETE /praca/api/slides", auth.RequireAdmin(HandleDeleteSlide))

	pngBytes := makeTestPNG(t)

	buildMultiUpload := func() (*bytes.Buffer, string) {
		var body bytes.Buffer
		mw := multipart.NewWriter(&body)
		for i := 1; i <= 2; i++ {
			fw, err := mw.CreateFormFile("images", fmt.Sprintf("foto%d.png", i))
			if err != nil {
				t.Fatalf("CreateFormFile: %v", err)
			}
			fw.Write(pngBytes)
		}
		mw.WriteField("title", "Praça Nobre")
		mw.WriteField("caption", "Lote de fotos")
		mw.Close()
		return &body, mw.FormDataContentType()
	}

	// 1. Public list works without auth
	req := httptest.NewRequest(http.MethodGet, "/praca/api/slides", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on public list, got %d", rec.Code)
	}

	// 2. Unauthenticated upload is rejected (401)
	body, ct := buildMultiUpload()
	req = httptest.NewRequest(http.MethodPost, "/praca/api/slides", body)
	req.Header.Set("Content-Type", ct)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated upload, got %d", rec.Code)
	}

	// 3. Regular user upload is forbidden (403)
	userCookie := makeSessionCookie(t, "user@example.com", "user")
	body, ct = buildMultiUpload()
	req = httptest.NewRequest(http.MethodPost, "/praca/api/slides", body)
	req.Header.Set("Content-Type", ct)
	req.AddCookie(userCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin upload, got %d", rec.Code)
	}

	// 4. Admin upload with multiple photos succeeds (201)
	adminCookie := makeSessionCookie(t, "edynobre@gmail.com", "admin")
	body, ct = buildMultiUpload()
	req = httptest.NewRequest(http.MethodPost, "/praca/api/slides", body)
	req.Header.Set("Content-Type", ct)
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for admin multi-upload, got %d: %s", rec.Code, rec.Body.String())
	}

	var created []models.SlideMeta
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode created slice: %v", err)
	}
	if len(created) != 2 {
		t.Fatalf("expected 2 created slides, got %d", len(created))
	}

	// 5. Public image fetch works without auth
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/praca/api/slides/image?id=%d", created[0].ID), nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !bytes.Equal(rec.Body.Bytes(), pngBytes) {
		t.Fatalf("expected 200 and matching png bytes, got %d", rec.Code)
	}

	// 6. Unauthenticated delete fails (401), admin delete succeeds (200)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/praca/api/slides?id=%d", created[0].ID), nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on unauthenticated delete, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/praca/api/slides?id=%d", created[0].ID), nil)
	req.AddCookie(adminCookie)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on admin delete, got %d", rec.Code)
	}
}
