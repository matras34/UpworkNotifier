package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

type KMS struct {
	masterKey []byte
}

func NewKMS(masterKeyStr string) (*KMS, error) {
	if len(masterKeyStr) < 32 {
		return nil, fmt.Errorf("master key must be at least 32 characters")
	}

	// Derive a proper key from the master key string
	key := pbkdf2.Key([]byte(masterKeyStr), []byte("webssh-salt"), 4096, 32, sha256.New)

	return &KMS{
		masterKey: key,
	}, nil
}

// Encrypt encrypts data and returns encrypted data, salt, and nonce
func (k *KMS) Encrypt(plaintext string) (encrypted, salt, nonce string, err error) {
	if plaintext == "" {
		return "", "", "", nil
	}

	// Generate a random salt for this encryption
	saltBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, saltBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive encryption key from master key and salt
	encKey := pbkdf2.Key(k.masterKey, saltBytes, 4096, 32, sha256.New)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceBytes := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonceBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonceBytes, []byte(plaintext), nil)

	return base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(saltBytes),
		base64.StdEncoding.EncodeToString(nonceBytes),
		nil
}

// Decrypt decrypts data using the provided salt and nonce
func (k *KMS) Decrypt(encryptedStr, saltStr, nonceStr string) (string, error) {
	if encryptedStr == "" {
		return "", nil
	}

	encrypted, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %w", err)
	}

	saltBytes, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode salt: %w", err)
	}

	nonceBytes, err := base64.StdEncoding.DecodeString(nonceStr)
	if err != nil {
		return "", fmt.Errorf("failed to decode nonce: %w", err)
	}

	// Derive the same encryption key
	encKey := pbkdf2.Key(k.masterKey, saltBytes, 4096, 32, sha256.New)

	block, err := aes.NewCipher(encKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonceBytes, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
