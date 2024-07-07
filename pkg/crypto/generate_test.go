package crypto

import (
	"encoding/base64"
	"testing"
)

func TestGenerateCipherKey(t *testing.T) {
	t.Run("valid length", func(t *testing.T) {
		length := 32
		key, err := GenerateCipherKey(length)
		if err != nil {
			t.Errorf("GenerateCipherKey(%d) returned unexpected error: %v", length, err)
		}
		byte_key, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			t.Errorf("GenerateCipherKey(%d) returned invalid base64 encoded key: %v", length, err)
		}
		if len(byte_key) != 32 {
			t.Errorf("GenerateCipherKey(%d) returned key with invalid length: %d", length, len(byte_key))
		}
	})

	t.Run("zero length", func(t *testing.T) {
		length := 0
		key, err := GenerateCipherKey(length)
		if err != nil {
			t.Errorf("GenerateCipherKey(%d) returned unexpected error: %v", length, err)
		}
		byte_key, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			t.Errorf("GenerateCipherKey(%d) returned invalid base64 encoded key: %v", length, err)
		}
		if len(byte_key) != 0 {
			t.Errorf("GenerateCipherKey(%d) returned non-empty key: %s", length, key)
		}
	})

	t.Run("negative length", func(t *testing.T) {
		length := -10
		key, err := GenerateCipherKey(length)
		if err == nil {
			t.Errorf("GenerateCipherKey(%d) should return error for negative length", length)
		}
		if key != "" {
			t.Errorf("GenerateCipherKey(%d) returned non-empty key: %s", length, key)
		}
	})
}
