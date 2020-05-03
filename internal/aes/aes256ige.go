package aes

import (
	"crypto/aes"
	"errors"
)

// Encrypt or Decrypt an input slice using AES IGE
func encryptDecryptIGE(in, key, iv []byte, encrypt bool) ([]byte, error) {
	// Check the inputs
	err := inputAESCheck(in, key, iv)
	if err != nil {
		return nil, err
	}
	if len(in) % aes.BlockSize != 0 {
		return nil, errors.New("input data length isn't a multiple of the block size")
	}
	if len(iv) != 32 {
		return nil, errors.New("IV length must be 32 bytes")
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
		xorSlice(iv1, chunk)

		if encrypt {
			// AES256 encrypt (aes package)
			block.Encrypt(expandedKey, iv1)
		} else {
			// AES256 decrypt (aes package)
			block.Decrypt(expandedKey, iv1)
		}

		// AddRoundKey passage
		// expandedKey = sessionKey ^ previous input (or iv)
		xorSlice(expandedKey, iv2)

		// Reassign iv1 and iv2 for next iteration
		// iv1: output just obtained
		// iv2: input just used
		iv1, iv2 = expandedKey, chunk

		// Copy iteration result to result variable
		copy(result[i:], expandedKey)
	}

	return result, nil
}

// Encrypt data with AES IGE
func EncryptIGE(in, key, iv []byte) ([]byte, error) {
	return encryptDecryptIGE(in, key, iv, true)
}

// Decrypt data with AES IGE
func DecryptIGE(in, key, iv []byte) ([]byte, error) {
	return encryptDecryptIGE(in, key, iv, false)
}