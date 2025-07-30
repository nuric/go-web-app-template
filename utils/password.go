package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/argon2"
)

// https://thecopenhagenbook.com/random-values
var humanFriendlyEncoding = base32.NewEncoding("0123456789ABCDEFGHJKMNPQRSTVWXYZ").WithPadding(base32.NoPadding)

func HumanFriendlyToken() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return humanFriendlyEncoding.EncodeToString(bytes)
}

// https://thecopenhagenbook.com/password-authentication

func GenerateRandomString() string {
	bytes := make([]byte, 12)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

func HashPassword(password string) string {
	salt := GenerateRandomString()
	hash := argon2.IDKey([]byte(password), []byte(salt), 2, 19*1024, 1, 32)
	return salt + "$" + base64.StdEncoding.EncodeToString(hash)

}

func VerifyPassword(storedHash, password string) bool {
	parts := strings.Split(storedHash, "$")
	if len(parts) != 2 {
		return false
	}
	salt := parts[0]
	expectedHash := parts[1]

	hash := argon2.IDKey([]byte(password), []byte(salt), 2, 19*1024, 1, 32)
	baseHash := base64.StdEncoding.EncodeToString(hash)
	return subtle.ConstantTimeCompare([]byte(expectedHash), []byte(baseHash)) == 1
}
