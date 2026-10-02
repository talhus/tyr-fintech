package encryption_test

import (
	"encoding/base64"
	"testing"

	"github.com/iamtbay/tyr-fintech/pkg/encryption"
)

const testKey = "12345678901234567890123456789012" // 32-byte AES-256 key

// 1. Test NewAESEncryptor Constructor
func TestNewAESEncryptor(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		expectedError error
	}{
		{
			name:          "valid 32-byte key",
			key:           testKey,
			expectedError: nil,
		},
		{
			name:          "short key (10 bytes)",
			key:           "short_key!",
			expectedError: encryption.ErrInvalidKeySize,
		},
		{
			name:          "long key (40 bytes)",
			key:           "1234567890123456789012345678901234567890",
			expectedError: encryption.ErrInvalidKeySize,
		},
		{
			name:          "empty key",
			key:           "",
			expectedError: encryption.ErrInvalidKeySize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := encryption.NewAESEncryptor(tt.key)
			if err != tt.expectedError {
				t.Fatalf("expected error %v, got %v", tt.expectedError, err)
			}
			if tt.expectedError == nil && enc == nil {
				t.Fatalf("expected non-nil encryptor instance")
			}
		})
	}
}

// 2. Test Encrypt Method
func TestAESEncryptor_Encrypt(t *testing.T) {
	enc, err := encryption.NewAESEncryptor(testKey)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	t.Run("empty string returns empty string", func(t *testing.T) {
		res, err := enc.Encrypt("")
		if err != nil {
			t.Fatalf("expected no error for empty input, got %v", err)
		}
		if res != "" {
			t.Fatalf("expected empty result, got '%s'", res)
		}
	})

	t.Run("valid card number encrypts successfully", func(t *testing.T) {
		pan := "4111111111111111"
		cipherText, err := enc.Encrypt(pan)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cipherText == pan {
			t.Fatalf("ciphertext should not match plaintext")
		}
		if len(cipherText) == 0 {
			t.Fatalf("ciphertext should not be empty")
		}
	})

	t.Run("nonce randomness produces unique ciphertexts", func(t *testing.T) {
		cvv := "123"
		c1, err1 := enc.Encrypt(cvv)
		c2, err2 := enc.Encrypt(cvv)

		if err1 != nil || err2 != nil {
			t.Fatalf("expected no errors encrypting CVV")
		}
		if c1 == c2 {
			t.Fatalf("encryptions of same text must yield different ciphertexts due to random nonces")
		}
	})
}

// 3. Test Decrypt Method
func TestAESEncryptor_Decrypt(t *testing.T) {
	enc, err := encryption.NewAESEncryptor(testKey)
	if err != nil {
		t.Fatalf("failed to create encryptor: %v", err)
	}

	t.Run("empty ciphertext returns empty string", func(t *testing.T) {
		res, err := enc.Decrypt("")
		if err != nil {
			t.Fatalf("expected no error for empty ciphertext, got %v", err)
		}
		if res != "" {
			t.Fatalf("expected empty string result, got '%s'", res)
		}
	})

	t.Run("encrypt and decrypt roundtrip success", func(t *testing.T) {
		original := "5500000000000004"
		encrypted, err := enc.Encrypt(original)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		decrypted, err := enc.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}
		if decrypted != original {
			t.Fatalf("expected decrypted string '%s', got '%s'", original, decrypted)
		}
	})

	t.Run("invalid base64 input fails", func(t *testing.T) {
		_, err := enc.Decrypt("!!!InvalidBase64!!!")
		if err != encryption.ErrInvalidData {
			t.Fatalf("expected ErrInvalidData for bad Base64, got %v", err)
		}
	})

	t.Run("short payload under nonce size fails", func(t *testing.T) {
		// base64 encoding of 5 bytes (less than 12-byte nonce)
		shortBase64 := base64.StdEncoding.EncodeToString([]byte("12345"))
		_, err := enc.Decrypt(shortBase64)
		if err != encryption.ErrInvalidData {
			t.Fatalf("expected ErrInvalidData for payload smaller than nonce, got %v", err)
		}
	})

	t.Run("tampered payload fails authentication", func(t *testing.T) {
		// Valid base64 of 20 random bytes (valid size, but invalid auth tag)
		tamperedBase64 := base64.StdEncoding.EncodeToString([]byte("12345678901234567890"))
		_, err := enc.Decrypt(tamperedBase64)
		if err != encryption.ErrInvalidData {
			t.Fatalf("expected ErrInvalidData for tampered payload, got %v", err)
		}
	})
}
