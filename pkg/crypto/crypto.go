package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

// Encrypt encrypts a plaintext string using AES-GCM and the provided 32-byte key.
// It returns a base64 encoded string containing the nonce and ciphertext.
func Encrypt(plaintext, key string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("crypto: key must be 32 bytes")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("crypto: read nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64 encoded string using AES-GCM and the provided 32-byte key.
func Decrypt(ciphertext64, key string) (string, error) {
	if len(key) != 32 {
		return "", errors.New("crypto: key must be 32 bytes")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertext64)
	if err != nil {
		return "", fmt.Errorf("crypto: decode base64: %w", err)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("crypto: new gcm: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("crypto: ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("crypto: open gcm: %w", err)
	}

	return string(plaintext), nil
}
