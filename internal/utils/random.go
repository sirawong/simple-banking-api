package utils

import (
	cryptoRand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

// GenerateAccountNumber returns a random 10-digit numeric account number.
func GenerateAccountNumber() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%010d", r.Int63n(9000000000)+1000000000)
}

// GenerateOpaqueToken generates a cryptographically random token for use as a refresh token.
func GenerateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := cryptoRand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
