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
func (aesCtr *AES256CTR) EncryptDecrypt(in []byte) {
	// Check the inputs
	if len(in) == 0 {
		return
	}

	// CTR stream encrypt/decrypt
	aesCtr.stream.XORKeyStream(in, in)
}