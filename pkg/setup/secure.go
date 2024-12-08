package setup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"

	"github.com/martient/bifrost-backups/pkg/crypto"
	"github.com/zalando/go-keyring"
)

const (
	keyringService = "bifrost-backups"
)

type SecureManager struct {
	masterKey    []byte
	noEncryption bool
}

// NewSecureManager creates a new secure manager instance
func NewSecureManager(noEncryption bool) (*SecureManager, error) {
	if noEncryption {
		return &SecureManager{noEncryption: true}, nil
	}

	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Try to get the master key from keyring
	keyStr, err := keyring.Get(keyringService, currentUser.Username)
	if err != nil {
		// Generate a new master key if none exists
		keyStr, err = crypto.GenerateCipherKey(32)
		if err != nil {
			return nil, fmt.Errorf("failed to generate master key: %w", err)
		}

		// Store the new key in keyring
		err = keyring.Set(keyringService, currentUser.Username, keyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to store master key: %w", err)
		}
	}

	masterKey, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode master key: %w", err)
	}

	return &SecureManager{masterKey: masterKey, noEncryption: false}, nil
}

// deriveHostKey derives a host-specific key using the master key and host information
func (sm *SecureManager) deriveHostKey() ([]byte, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	// Create a unique host identifier by combining hostname with master key
	h := sha256.New()
	h.Write([]byte(hostname))
	h.Write(sm.masterKey)

	return h.Sum(nil), nil
}

// encrypt encrypts a string using AES-256-GCM with host-specific key
func (sm *SecureManager) encrypt(plaintext string) (string, error) {
	if plaintext == "" || sm.noEncryption {
		return plaintext, nil
	}

	hostKey, err := sm.deriveHostKey()
	if err != nil {
		return "", fmt.Errorf("failed to derive host key: %w", err)
	}

	block, err := aes.NewCipher(hostKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt decrypts a string using AES-256-GCM with host-specific key
func (sm *SecureManager) decrypt(ciphertext string) (string, error) {
	if ciphertext == "" || sm.noEncryption {
		return ciphertext, nil
	}

	// If we're trying to decrypt a non-encrypted string, return it as is
	if !strings.HasPrefix(ciphertext, "ENC[AES256,") {
		return ciphertext, nil
	}

	// Extract the actual encrypted value
	encValue := strings.TrimPrefix(ciphertext, "ENC[AES256,")
	encValue = strings.TrimSuffix(encValue, "]")

	data, err := base64.StdEncoding.DecodeString(encValue)
	if err != nil {
		return ciphertext, nil // Return original if not base64
	}

	hostKey, err := sm.deriveHostKey()
	if err != nil {
		return "", fmt.Errorf("failed to derive host key: %w", err)
	}

	block, err := aes.NewCipher(hostKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return ciphertext, nil // Return original if too short
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil) //#nosec
	if err != nil {
		return ciphertext, nil // Return original if decryption fails
	}

	return string(plaintext), nil
}

// SecureConfig encrypts sensitive fields in the configuration
func (sm *SecureManager) SecureConfig(config *Config) error {
	if sm.noEncryption {
		return nil
	}

	// Encrypt database credentials
	for i := range config.Databases {
		switch config.Databases[i].Type {
		case Postgresql:
			if config.Databases[i].Postgresql.Password != "" && !strings.HasPrefix(config.Databases[i].Postgresql.Password, "ENC[AES256,") {
				encrypted, err := sm.encrypt(config.Databases[i].Postgresql.Password)
				if err != nil {
					return fmt.Errorf("failed to encrypt PostgreSQL password: %w", err)
				}
				config.Databases[i].Postgresql.Password = fmt.Sprintf("ENC[AES256,%s]", encrypted)
			}
		}
	}

	// Encrypt storage credentials
	for i := range config.Storages {
		switch config.Storages[i].Type {
		case S3:
			if config.Storages[i].S3.AccessKeySecret != "" && !strings.HasPrefix(config.Storages[i].S3.AccessKeySecret, "ENC[AES256,") {
				encrypted, err := sm.encrypt(config.Storages[i].S3.AccessKeySecret)
				if err != nil {
					return fmt.Errorf("failed to encrypt S3 access key secret: %w", err)
				}
				config.Storages[i].S3.AccessKeySecret = fmt.Sprintf("ENC[AES256,%s]", encrypted)
			}
		}
		// Encrypt CipherKey if not already encrypted
		if config.Storages[i].CipherKey != "" && !strings.HasPrefix(config.Storages[i].CipherKey, "ENC[AES256,") {
			encrypted, err := sm.encrypt(config.Storages[i].CipherKey)
			if err != nil {
				return fmt.Errorf("failed to encrypt storage cipher key: %w", err)
			}
			config.Storages[i].CipherKey = fmt.Sprintf("ENC[AES256,%s]", encrypted)
		}
	}

	return nil
}

// DecryptConfig decrypts sensitive fields in the configuration
func (sm *SecureManager) DecryptConfig(config *Config) error {
	// Decrypt database credentials
	for i := range config.Databases {
		switch config.Databases[i].Type {
		case Postgresql:
			if config.Databases[i].Postgresql.Password != "" {
				decrypted, err := sm.decrypt(config.Databases[i].Postgresql.Password)
				if err != nil {
					return fmt.Errorf("failed to decrypt PostgreSQL password: %w", err)
				}
				config.Databases[i].Postgresql.Password = decrypted
			}
		}
	}

	// Decrypt storage credentials
	for i := range config.Storages {
		switch config.Storages[i].Type {
		case S3:
			if config.Storages[i].S3.AccessKeySecret != "" {
				decrypted, err := sm.decrypt(config.Storages[i].S3.AccessKeySecret)
				if err != nil {
					return fmt.Errorf("failed to decrypt S3 access key secret: %w", err)
				}
				config.Storages[i].S3.AccessKeySecret = decrypted
			}
		}
		// Decrypt CipherKey if encrypted
		if strings.HasPrefix(config.Storages[i].CipherKey, "ENC[AES256,") {
			decrypted, err := sm.decrypt(config.Storages[i].CipherKey)
			if err != nil {
				return fmt.Errorf("failed to decrypt storage cipher key: %w", err)
			}
			config.Storages[i].CipherKey = decrypted
		}
	}

	return nil
}
