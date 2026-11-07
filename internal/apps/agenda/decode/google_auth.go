package decode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

type GoogleClaims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"`
}

func DecodeGoogleToken(credential string) (*GoogleClaims, error) {
	parts := strings.Split(credential, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid token")
	}

	payloadPart := parts[1]
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadPart)
	if err != nil {
		return nil, errors.New("error decoding token: " + err.Error())
	}

	var claims GoogleClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, errors.New("error parsing token claims")
	}

	return &claims, nil
}
