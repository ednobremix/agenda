package decode

import (
	"context"
	"errors"
	"os"
	"strings"

	"google.golang.org/api/idtoken"
)

const defaultGoogleClientID = "665567298487-am8coo2krrd2msdno64p3ng5om1pfr75.apps.googleusercontent.com"

type GoogleClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}

func DecodeGoogleToken(ctx context.Context, credential string) (*GoogleClaims, error) {
	if strings.TrimSpace(credential) == "" {
		return nil, errors.New("missing google token")
	}

	clientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	if clientID == "" {
		clientID = defaultGoogleClientID
	}

	payload, err := idtoken.Validate(ctx, credential, clientID)
	if err != nil {
		return nil, errors.New("invalid google token: " + err.Error())
	}

	emailVerified, ok := payload.Claims["email_verified"].(bool)
	if !ok || !emailVerified {
		return nil, errors.New("google email not verified")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	if email == "" || payload.Subject == "" {
		return nil, errors.New("missing required claims in google token")
	}

	return &GoogleClaims{
		Email: email,
		Name:  name,
		Sub:   payload.Subject,
	}, nil
}
