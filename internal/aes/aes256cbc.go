package aes

import(
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

func encryptDecryptCBC(in, key, iv []byte, encrypt bool) ([]byte, error) {
	// Check the inputs
	err := inputAESCheck(in, key, iv)
	if err != nil {
		return nil, err
	}
	if len(in) % aes.BlockSize != 0 {
		return nil, errors.New("input data length isn't a multiple of the block size")
	}
	if len(iv) != 16 {
		return nil, errors.New("IV length must be 16 bytes")
	}

	// Create a new AES cipher with the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Result variable
	result := make([]byte, len(in))

	var mode cipher.BlockMode
	if encrypt {
		// Encrypt input
		mode = cipher.NewCBCEncrypter(block, iv)
	} else {
		// Decrypt input
		mode = cipher.NewCBCDecrypter(block, iv)
	}
	mode.CryptBlocks(result, in)

	// Return encrypted/decrypted result
	return result, nil
}

// Encrypt data with AES CBC
func EncryptCBC(in, key, iv []byte) ([]byte, error) {
	return encryptDecryptCBC(in, key, iv, true)
}

// Decrypt data with AES CBC
func DecryptCBC(in, key, iv []byte) ([]byte, error){
	return encryptDecryptCBC(in, key, iv, false)
}