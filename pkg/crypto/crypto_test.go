package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	key := "12345678901234567890123456789012" // 32 bytes
	plaintext := "secret message"

	// 1. Encrypt
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if ciphertext == "" {
		t.Fatal("Ciphertext is empty")
	}

	// 2. Decrypt
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncrypt_InvalidKey(t *testing.T) {
	key := "short"
	_, err := Encrypt("msg", key)
	if err == nil {
		t.Error("Expected error for short key, got nil")
	}
}

func TestDecrypt_InvalidKey(t *testing.T) {
	key := "short"
	_, err := Decrypt("msg", key)
	if err == nil {
		t.Error("Expected error for short key, got nil")
	}
}

func TestDecrypt_Corrupted(t *testing.T) {
	key := "12345678901234567890123456789012"

	// Create valid ciphertext
	valid, _ := Encrypt("msg", key)

	// Corrupt it (change last char)
	corrupted := valid[:len(valid)-1] + "A"

	// If base64 is still valid, GCM Open should fail authentication
	// If base64 is invalid, DecodeString should fail
	_, err := Decrypt(corrupted, key)
	if err == nil {
		// It's possible we just happened to hit a valid tag, but unlikely.
		// However, modifying base64 might make it invalid length too.
		// Let's try to decode, flip a bit in ciphertext, and re-encode.
		data, _ := base64.StdEncoding.DecodeString(valid)
		data[len(data)-1] ^= 0x01 // Flip last bit
		corrupted = base64.StdEncoding.EncodeToString(data)

		_, err = Decrypt(corrupted, key)
		if err == nil {
			t.Error("Expected error for corrupted ciphertext, got nil")
		} else {
			if !strings.Contains(err.Error(), "open gcm") {
				t.Logf("Got expected error but message differs: %v", err)
			}
		}
	}
}
