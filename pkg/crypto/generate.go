package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateCipherKey(length int) (string, error) {
	if length < 0 {
		return "", fmt.Errorf("length must be greater than zero")
	}
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
