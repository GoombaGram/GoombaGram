package aes

import (
	"errors"
)


// Function that make a byte xorSlice between two byte slices
func xorSlice(dst, src []byte) {
	for i := range dst {
		dst[i] = dst[i] ^ src[i]
	}
}

// Check the AES IGE inputs
func inputAESCheck(data, key, iv []byte) error {
	if len(key) == 0 || len(iv) == 0 || len(data) == 0 {
		return errors.New("some parameters are empty")
	}

	if len(key) != 32 {
		return errors.New("key length must be 32 bytes")
	}

	return nil
}