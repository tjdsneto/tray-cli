package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// NewCodeVerifier returns a PKCE code_verifier (RFC 7636) and its S256 code_challenge.
func NewCodeVerifier() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	v := base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(v))
	c := base64.RawURLEncoding.EncodeToString(h[:])
	return v, c, nil
}
