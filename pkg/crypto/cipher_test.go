package crypto

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestCipher(t *testing.T) {
	plaintext := []byte("hello world")
	cipher_key, err := GenerateCipherKey(32)
	if err != nil {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	byte_cipher_key, err := base64.StdEncoding.DecodeString(cipher_key)
	if err != nil {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	if len(byte_cipher_key) != 32 {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	cipher_text, err := Cipher(byte_cipher_key, plaintext)
	if err != nil {
		t.Errorf("Cipher failed: %v", err)
	}

	if len(cipher_text.String()) == 0 {
		t.Error("Cipher returned empty cipher_text")
	}
}

func TestDecipher(t *testing.T) {
	plaintext := []byte("hello world")
	cipher_key, err := GenerateCipherKey(32)
	if err != nil {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	byte_cipher_key, err := base64.StdEncoding.DecodeString(cipher_key)
	if err != nil {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	if len(byte_cipher_key) != 32 {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	cipher_text, err := Cipher(byte_cipher_key, plaintext)
	if err != nil {
		t.Errorf("Cipher failed: %v", err)
	}

	decipher, err := Decipher(byte_cipher_key, cipher_text.Bytes())
	if err != nil {
		t.Errorf("Decipher failed: %v", err)
	}

	if !bytes.Equal(decipher.Bytes(), plaintext) {
		t.Errorf("Decipher failed: got %s, want %s", decipher.String(), string(plaintext))
	}
}

func TestCipherDecipherWithInvalidKey(t *testing.T) {
	plaintext := []byte("hello world")
	invalid_cipher_key, err := GenerateCipherKey(11)
	if err != nil {
		t.Errorf("Invalid GenerateCipherKey failed: %v", err)
	}
	byte_invalid_cipher_key, err := base64.StdEncoding.DecodeString(invalid_cipher_key)
	if err != nil {
		t.Errorf("Invalid GenerateCipherKey failed: %v", err)
	}
	if len(byte_invalid_cipher_key) != 11 {
		t.Errorf("Invalid GenerateCipherKey failed: %v", err)
	}

	cipher_key, err := GenerateCipherKey(32)
	if err != nil {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	byte_cipher_key, err := base64.StdEncoding.DecodeString(cipher_key)
	if err != nil {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}
	if len(byte_cipher_key) != 32 {
		t.Errorf("GenerateCipherKey failed: %v", err)
	}

	_, err = Cipher(byte_invalid_cipher_key, plaintext)
	if err == nil {
		t.Error("Encrypt should fail with invalid key")
	}

	cipher_text, _ := Cipher(byte_cipher_key, plaintext)
	_, err = Decipher(byte_invalid_cipher_key, cipher_text.Bytes())
	if err == nil {
		t.Error("Decrypt should fail with invalid key")
	}
}
