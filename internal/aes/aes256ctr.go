package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// Function to decode or encode using AES CTR
func encryptDecryptCTR(in, key, iv []byte, encrypt bool) ([]byte, error) {
	// Check the inputs
	err := inputAESCheck(in, key, iv)
	if err != nil {
		return nil, err
	}
	if len(iv) != 16 {
		return nil, errors.New("IV length must be 16 bytes")
	}

	// Create a new AES cipher with the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new CTR stream
	stream := cipher.NewCTR(block, iv)

	// XOR CTR stream encrypt/decrypt
	result := make([]byte, len(in))
	stream.XORKeyStream(result, in)

	// Return result (encrypted/decrypted)
	return result, nil
}

// Encrypt data with AES CTR
func EncryptCTR(in, key, iv []byte) ([]byte, error) {
	return encryptDecryptCTR(in, key, iv, true)
}

// Decrypt data with AES CTR
func DecryptCTR(in, key, iv []byte) ([]byte, error) {
	return encryptDecryptCTR(in, key, iv, false)
}