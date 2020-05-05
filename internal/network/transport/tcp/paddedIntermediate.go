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

package tcp

import (
	"encoding/binary"
	"github.com/GoombaGram/GoombaGram/internal/crypto/aes"
	"math/rand"
)

// Padded intermediate
//
// Padded version of the intermediate protocol, to use with obfuscation enabled to bypass ISP blocks.
//
// Overhead: small-medium
// Minimum envelope length: random
// Maximum envelope length: random
type PaddedIntermediate struct{
	*tcpConnection                  // TCP connection
	encrypt, decrypt *aes.AES256CTR // AES-256 CTR encrypt/decrypt (only with obfuscation true)
}

func (pad *PaddedIntermediate) Connect(address string, obfuscation bool) error {
	pad.tcpConnection = tcpNew()

	err := pad.tcpConnection.connect(address)
	if err != nil {
		return err
	}

	if obfuscation {
		nonce, reversedNonce := obfuscationCTRGenerator(0xDD)

		pad.encrypt = aes.AES256CTRNew(nonce[8:40], nonce[40:56])
		pad.decrypt = aes.AES256CTRNew(reversedNonce[8:40], reversedNonce[40:56])

		// Add aes encrypted to nonce
		aesNonce := make([]byte, 64); copy(aesNonce, nonce)
		pad.encrypt.EncryptDecrypt(aesNonce)

		// Send encrypted nonce to server (when connect)
		err = pad.tcpConnection.sendAll(append(nonce[:56], aesNonce[56:64]...))
	} else {
		// Telegram docs:
		// Before sending anything into the underlying socket (see transports), the client must first send 0xdddddddd as the first int (four bytes, the server will not send 0xdddddddd as the first int in the first reply).
		err = pad.tcpConnection.sendAll([]byte{0xDD, 0xDD, 0xDD, 0xDD})
	}

	if err != nil {
		return err
	}

	return nil
}

// Send data to Telegram server using Intermediate TCP
//
// +----+----...----+----...----+
// |tLen|  payload  |  padding  |
// +----+----...----+----...----+
//
// Total length: payload+padding length encoded as 4 length bytes (little endian)
// Payload: the MTProto payload
// Padding: A random padding string of length 0-15
func (pad *PaddedIntermediate) Send(data []byte) error {
	// Generate a random number between 0 and 15
	padding := make([]byte, rand.Intn(16))
	if len(padding) != 0 {
		rand.Read(padding)
	}

	// Parse length to 4 bytes slice
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(data) + len(padding)))

	// Add data and padding
	data = append(append(length, data...), padding...)

	if pad.encrypt != nil {
		pad.encrypt.EncryptDecrypt(data)
	}

	err := pad.tcpConnection.sendAll(data)
	if err != nil {
		return err
	}

	return nil
}

// Receive n data to Telegram server using Intermediate TCP
//
// +----+----...----+----...----+
// |tLen|  payload  |  padding  |
// +----+----...----+----...----+
func (pad *PaddedIntermediate) Receive(data []byte) error {
	length, err := pad.tcpConnection.receiveAll(4)
	if err != nil {
		return err
	}

	// Decrypt length
	if pad.decrypt != nil {
		pad.decrypt.EncryptDecrypt(length)
	}

	// Get length of data as int
	lenInt := int(binary.LittleEndian.Uint32(length))

	// Get n bytes
	data, err = pad.tcpConnection.receiveAll(lenInt)
	if err != nil {
		return err
	}

	// Decrypt received data
	if pad.decrypt != nil {
		pad.decrypt.EncryptDecrypt(data)
	}

	return nil
}