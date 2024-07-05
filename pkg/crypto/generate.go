package crypto

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateCipherKey(length int) (string, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
