package setup

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
)

const (
	saltLength   = 16
	iterations   = 100000
	keyLength    = 32
	maxPasswordLength = 1024 // Maximum password length in bytes
)

// SetMasterPassword sets the master password for config export
func SetMasterPassword(config *Config, password string) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if password == "" {
		return fmt.Errorf("master password cannot be empty")
	}

	if len(password) > maxPasswordLength {
		return fmt.Errorf("password exceeds maximum length of %d bytes", maxPasswordLength)
	}

	// Generate a random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate hash using PBKDF2
	hash := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)

	// Combine salt and hash, and encode as base64
	combined := append(salt, hash...)
	config.MasterHash = base64.StdEncoding.EncodeToString(combined)

	return nil
}

// ValidateMasterPassword checks if the provided password matches the stored hash
func ValidateMasterPassword(config *Config, password string) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if config.MasterHash == "" {
		return fmt.Errorf("master password not set")
	}

	if len(password) > maxPasswordLength {
		return fmt.Errorf("password exceeds maximum length of %d bytes", maxPasswordLength)
	}

	// Decode the stored hash
	combined, err := base64.StdEncoding.DecodeString(config.MasterHash)
	if err != nil {
		return fmt.Errorf("invalid master hash format")
	}

	if len(combined) < saltLength+1 {
		return fmt.Errorf("invalid master hash length")
	}

	// Split salt and hash
	salt := combined[:saltLength]
	storedHash := combined[saltLength:]

	// Generate hash from provided password
	hash := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha256.New)

	// Compare hashes
	if !compareHashes(hash, storedHash) {
		return fmt.Errorf("invalid master password")
	}

	return nil
}

// compareHashes compares two hashes in constant time
func compareHashes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
