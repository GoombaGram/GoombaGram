package aes

import (
	"crypto/aes"
	"crypto/cipher"
)

type AES256CTR struct {
	stream cipher.Stream
}

func AES256CTRNew(aesKey, aesIV []byte) *AES256CTR {
	if len(aesIV) != 16 || len(aesKey) == 0 {
		return nil
	}

	// Create a new AES cipher with the key
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil
	}

	// Create a new CTR stream
	ctr := new(AES256CTR)
	stream := cipher.NewCTR(block, aesIV)
	ctr.stream = stream

	return ctr
}

// Function to decrypt or encrypt using AES CTR
func (aesCtr *AES256CTR) EncryptDecrypt(in []byte) []byte {
	// Check the inputs
	if len(in) == 0 {
		return nil
	}

	// XOR CTR stream encrypt/decrypt
	result := make([]byte, len(in))
	aesCtr.stream.XORKeyStream(result, in)

	// Return result (encrypted/decrypted)
	return result
}