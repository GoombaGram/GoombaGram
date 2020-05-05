/*
 * Copyright (c) 2020 ErikPelli <https://github.com/ErikPelli>
 * This file is part of GoombaGram.
 *
 * GoombaGram is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * GoombaGram is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 * You should have received a copy of the GNU Affero General Public License
 * along with GoombaGram.  If not, see <http://www.gnu.org/licenses/>.
 */

package aes

import(
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type AES256CBC struct {
	aesKey, aesIV []byte
}

func AES256CBCNew(aesKey, aesIV []byte) *AES256CBC {
	if len(aesIV) != 16 {
		return nil
	}

	if !(len(aesKey) == 16 || len(aesKey) == 24 || len(aesKey) == 32) {
		return nil
	}

	return &AES256CBC{
		aesKey: aesKey,
		aesIV:  aesIV,
	}
}

func encryptDecryptCBC(in, key, iv []byte, encrypt bool) ([]byte, error) {
	// Check the inputs
	if len(in) % aes.BlockSize != 0 && len(in) != 0{
		return nil, errors.New("input data length isn't a multiple of the block size")
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
func (aes *AES256CBC) Encrypt(in []byte) ([]byte, error) {
	return encryptDecryptCBC(in, aes.aesKey, aes.aesIV, true)
}

// Decrypt data with AES CBC
func (aes *AES256CBC) Decrypt(in []byte) ([]byte, error){
	return encryptDecryptCBC(in, aes.aesKey, aes.aesIV, false)
}