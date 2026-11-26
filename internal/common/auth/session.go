package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	common_models "massage-scheduler/internal/common/models"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	sessionCookieName = "portal_session"
	sessionDuration   = 7 * 24 * time.Hour
)

type contextKey string

const userContextKey contextKey = "authenticated_user"

type SessionClaims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	Exp   int64  `json:"exp"`
}

var (
	secretOnce sync.Once
	secretKey  []byte
)

func getSessionSecret() []byte {
	secretOnce.Do(func() {
		if envSecret := strings.TrimSpace(os.Getenv("SESSION_SECRET")); envSecret != "" {
			secretKey = []byte(envSecret)
			return
		}
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("erro ao gerar chave de sessão: %v", err)
		}
		secretKey = buf
		log.Println("AVISO: SESSION_SECRET não definido; usando chave aleatória em memória (sessões expiram ao reiniciar).")
	})
	return secretKey
}

func signPayload(payloadB64 string) string {
	mac := hmac.New(sha256.New, getSessionSecret())
	mac.Write([]byte(payloadB64))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func isSecureRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func SetSessionCookie(w http.ResponseWriter, r *http.Request, email, role string) error {
	claims := SessionClaims{
		Email: email,
		Role:  role,
		Exp:   time.Now().Add(sessionDuration).Unix(),
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return err
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	sig := signPayload(payloadB64)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    payloadB64 + "." + sig,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func GetSessionClaims(r *http.Request) (*SessionClaims, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil, errors.New("sessão ausente")
	}

	parts := strings.Split(cookie.Value, ".")
	if len(parts) != 2 {
		return nil, errors.New("token de sessão inválido")
	}

	expectedSig := signPayload(parts[0])
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return nil, errors.New("assinatura de sessão inválida")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("payload de sessão inválido")
	}

	var claims SessionClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("claims de sessão inválidos")
	}

	if claims.Email == "" || time.Now().Unix() > claims.Exp {
		return nil, errors.New("sessão expirada")
	}

	return &claims, nil
}

func AuthenticateRequest(r *http.Request) (*common_models.User, error) {
	claims, err := GetSessionClaims(r)
	if err != nil {
		return nil, err
	}

	var user common_models.User
	if err := common_models.GORM.Where("email = ?", claims.Email).First(&user).Error; err != nil {
		return nil, errors.New("usuário não encontrado")
	}
	return &user, nil
}

func WithUser(ctx context.Context, user *common_models.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*common_models.User, bool) {
	user, ok := ctx.Value(userContextKey).(*common_models.User)
	return user, ok && user != nil
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := AuthenticateRequest(r)
		if err != nil {
			http.Error(w, "Não autenticado", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	}
}

const (
	AppAgenda     = "agenda"
	AppFinanceiro = "financeiro"
	AppEstoque    = "estoque"
)

func HasAppAccess(user *common_models.User, app string) bool {
	if user == nil {
		return false
	}
	if user.Role == "admin" {
		return true
	}
	switch app {
	case AppAgenda:
		return user.HasAgenda
	case AppFinanceiro:
		return user.HasFinanceiro
	case AppEstoque:
		return user.HasEstoque
	default:
		return false
	}
}

func RequireAppAccess(app string, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok {
			http.Error(w, "Não autenticado", http.StatusUnauthorized)
			return
		}
		if !HasAppAccess(user, app) {
			http.Error(w, "Acesso negado ao aplicativo", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := AuthenticateRequest(r)
		if err != nil {
			http.Error(w, "Não autenticado", http.StatusUnauthorized)
			return
		}
		if user.Role != "admin" {
			http.Error(w, "Acesso restrito a administradores", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithUser(r.Context(), user)))
	}
}
