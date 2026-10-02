package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type Encryptor interface {
	Encrypt(plainText string) (string, error)
	Decrypt(cipherText string) (string, error)
}

type AESEncryptor struct {
	key []byte
}

var (
	ErrInvalidKeySize = errors.New("encryption key must be exactly 32 bytes for aes-256")
	ErrInvalidData    = errors.New("invalid or corrupted encrypted data")
)

func NewAESEncryptor(keyString string) (*AESEncryptor, error) {
	key := []byte(keyString)
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	return &AESEncryptor{
		key: key,
	}, nil
}

// METHODS

// Encryption
func (a *AESEncryptor) Encrypt(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}

	//create aes cipher
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	//wrap the block in gcm
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//create a slice to hold random nonce
	nonce := make([]byte, aesGCM.NonceSize()) //12 bytes
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	//encrypt and seal
	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func (a *AESEncryptor) Decrypt(cipherTextBase64 string) (string, error) {

	if cipherTextBase64 == "" {
		return "", nil
	}

	//base64 decode
	cipherText, err := base64.StdEncoding.DecodeString(cipherTextBase64)
	if err != nil {
		return "", ErrInvalidData
	}

	//create cipher block
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return "", err
	}

	//wrap the block in gcm
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//check if cipher text is at least 12 bytes (nonceSize)
	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return "", ErrInvalidData
	}

	//split the nonce from the ciphertext payload
	nonce, actualCipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	//open and authenticate
	plainText, err := aesGCM.Open(nil, nonce, actualCipherText, nil)
	if err != nil {
		return "", ErrInvalidData
	}

	return string(plainText), nil

}
