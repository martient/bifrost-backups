package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

func Decipher(key []byte, cipher_text []byte) (*bytes.Buffer, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(cipher_text) < nonceSize {
		return nil, fmt.Errorf("the cipher text is too short")
	}

	nonce, cipher_text := cipher_text[:nonceSize], cipher_text[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, cipher_text, nil) //#nosec
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(plaintext), nil
}
