package auth

import (
	common_models "massage-scheduler/internal/common/models"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionCookieRoundTripAndTamperDetection(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8080/agenda/api/auth/google", nil)
	rec := httptest.NewRecorder()

	if err := SetSessionCookie(rec, req, "user@example.com", "user"); err != nil {
		t.Fatalf("SetSessionCookie failed: %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected 1 cookie, got %d", len(cookies))
	}
	cookie := cookies[0]
	if !cookie.HttpOnly {
		t.Fatalf("expected HttpOnly cookie")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
	}

	authReq := httptest.NewRequest(http.MethodGet, "http://localhost:8080/agenda/api/me", nil)
	authReq.AddCookie(cookie)

	claims, err := GetSessionClaims(authReq)
	if err != nil {
		t.Fatalf("GetSessionClaims failed: %v", err)
	}
	if claims.Email != "user@example.com" || claims.Role != "user" {
		t.Fatalf("unexpected claims: %+v", claims)
	}

	// Tampered signature must be rejected
	tampered := &http.Cookie{
		Name:  sessionCookieName,
		Value: cookie.Value + "x",
	}
	tamperReq := httptest.NewRequest(http.MethodGet, "http://localhost:8080/agenda/api/me", nil)
	tamperReq.AddCookie(tampered)

	if _, err := GetSessionClaims(tamperReq); err == nil {
		t.Fatalf("expected tampered session cookie to be rejected")
	}
}

func TestRequireAuthAndRequireAdminRejectUnauthenticated(t *testing.T) {
	nextCalled := false
	handler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})

	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/agenda/api/me?email=victim@example.com", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized || nextCalled {
		t.Fatalf("expected 401 and next not called, got %d (called=%v)", rec.Code, nextCalled)
	}

	adminHandler := RequireAdmin(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})
	rec2 := httptest.NewRecorder()
	adminHandler.ServeHTTP(rec2, req)

	if rec2.Code != http.StatusUnauthorized || nextCalled {
		t.Fatalf("expected 401 for admin route without cookie, got %d", rec2.Code)
	}
}

func TestHasAppAccess(t *testing.T) {
	if HasAppAccess(nil, AppAgenda) {
		t.Fatalf("expected nil user to be denied")
	}

	admin := &common_models.User{Role: "admin"}
	for _, app := range []string{AppAgenda, AppFinanceiro, AppEstoque} {
		if !HasAppAccess(admin, app) {
			t.Fatalf("expected admin to have access to %s", app)
		}
	}

	user := &common_models.User{Role: "user", HasAgenda: true, HasEstoque: true}
	if !HasAppAccess(user, AppAgenda) || HasAppAccess(user, AppFinanceiro) || !HasAppAccess(user, AppEstoque) {
		t.Fatalf("unexpected app access evaluation for user: %+v", user)
	}
	if HasAppAccess(user, "desconhecido") {
		t.Fatalf("expected unknown app to be denied")
	}
}
