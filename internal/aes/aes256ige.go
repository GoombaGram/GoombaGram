package aes

import (
	"crypto/aes"
	"errors"
)

// Function that make a byte xor between two byte slices
func xor(dst, src []byte) {
	for i := range dst {
		dst[i] = dst[i] ^ src[i]
	}
}

// Check the AES IGE inputs
func inputIGE(data, key, iv []byte) error {
	if len(data) % aes.BlockSize != 0 || len(data) == 0 {
		return errors.New("AES256 IGE: data isn't divisible by block size (16 bytes)")
	}

	if len(key) == 0 {
		return errors.New("key is empty")
	}

	if len(iv) == 0 {
		return errors.New("IV is empty")
	}

	return nil
}

// Encrypt or Decrypt an input slice using AES IGE
func encryptDecryptIGE(in, key, iv []byte, encrypt bool) ([]byte, error) {
	// Check the inputs
	err := inputIGE(in, key, iv)
	if err != nil {
		return nil, err
	}

	// Create a new AES cipher with the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create the result variable
	result := make([]byte, len(in))

	// First part of IV (its length is 1 block)
	iv1 := make([]byte, aes.BlockSize)
	// Second part of IV (its length is 1 block)
	iv2 := make([]byte, aes.BlockSize)

	// Encrypt and decrypt (inverting iv1 and iv2)
	if encrypt {
		copy(iv1, iv[:aes.BlockSize])
		copy(iv2, iv[aes.BlockSize:])
	} else {
		copy(iv1, iv[aes.BlockSize:])
		copy(iv2, iv[:aes.BlockSize])
	}

	// AES expandedKey (256 bit)
	expandedKey := make([]byte, aes.BlockSize)

	// Iterate over input slice
	for i := 0; i < len(in); i += aes.BlockSize {
		// current byte chunk (16 bytes length)
		chunk := in[i:i+aes.BlockSize]

		// iv1 = previous output (or iv) ^ current input
		xor(iv1, chunk)

		if encrypt {
			// AES256 encrypt (aes package)
			block.Encrypt(expandedKey, iv1)
		} else {
			// AES256 decrypt (aes package)
			block.Decrypt(expandedKey, iv1)
		}

		// AddRoundKey passage
		// expandedKey = sessionKey ^ previous input (or iv)
		xor(expandedKey, iv2)

		// Reassign iv1 and iv2 for next iteration
		// iv1: output just obtained
		// iv2: input just used
		iv1, iv2 = expandedKey, chunk

		// Copy iteration result to result variable
		copy(result[i:], expandedKey)
	}

	return result, nil
}

// Encrypt data with AES256 IGE
func EncryptIGE(in, key, iv []byte) ([]byte, error) {
	return encryptDecryptIGE(in, key, iv, true)
}

// Decrypt data with AES256 IGE
func DecryptIGE(in, key, iv []byte) ([]byte, error) {
	return encryptDecryptIGE(in, key, iv, false)
}