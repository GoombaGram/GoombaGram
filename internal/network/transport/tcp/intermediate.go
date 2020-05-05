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
	"github.com/TelegramGo/TelegramGo/internal/crypto/aes"
)

// Intermediate
//
// In case 4-byte data alignment is needed, an intermediate version of the original protocol may be used.
//
// Overhead: small
// Minimum envelope length: 4 bytes
// Maximum envelope length: 4 bytes
type Intermediate struct{
	*tcpConnection                  // TCP connection
	encrypt, decrypt *aes.AES256CTR // AES-256 CTR encrypt/decrypt (only with obfuscation true)
}

func (inter *Intermediate) Connect(address string, obfuscation bool) error {
	inter.tcpConnection = tcpNew()

	err := inter.tcpConnection.connect(address)
	if err != nil {
		return err
	}

	if obfuscation {
		nonce, reversedNonce := obfuscationCTRGenerator(0xEE)

		inter.encrypt = aes.AES256CTRNew(nonce[8:40], nonce[40:56])
		inter.decrypt = aes.AES256CTRNew(reversedNonce[8:40], reversedNonce[40:56])

		// Add aes encrypted to nonce
		aesNonce := make([]byte, 64); copy(aesNonce, nonce)
		inter.encrypt.EncryptDecrypt(aesNonce)

		// Send encrypted nonce to server (when connect)
		err = inter.tcpConnection.sendAll(append(nonce[:56], aesNonce[56:64]...))
	} else {
		// Telegram docs:
		// Before sending anything into the underlying socket (see transports), the client must first send 0xeeeeeeee as the first int (four bytes, the server will not send 0xeeeeeeee as the first int in the first reply).	err = inter.tcpConnection.sendAll([]byte{0xEF})
		err = inter.tcpConnection.sendAll([]byte{0xEE, 0xEE, 0xEE, 0xEE})
	}

	if err != nil {
		return err
	}

	return nil
}

// Send data to Telegram server using Intermediate TCP
//
// +----+----...----+
// +len.+  payload  +
// +----+----...----+
//
// Length: payload length encoded as 4 length bytes (little endian)
// Payload: the MTProto payload
func (inter *Intermediate) Send(data []byte) error {
	// Parse length to 4 bytes slice
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(data)))
	data = append(length, data...)

	if inter.encrypt != nil {
		inter.encrypt.EncryptDecrypt(data)
	}

	err := inter.tcpConnection.sendAll(data)
	if err != nil {
		return err
	}

	return nil
}

// Receive n data to Telegram server using Intermediate TCP
//
// +----+----...----+
// +len.+  payload  +
// +----+----...----+
func (inter *Intermediate) Receive(data []byte) error {
	length, err := inter.tcpConnection.receiveAll(4)
	if err != nil {
		return err
	}

	// Decrypt length
	if inter.decrypt != nil {
		inter.decrypt.EncryptDecrypt(length)
	}

	// Get length of data as int
	lenInt := int(binary.LittleEndian.Uint32(length))

	// Get n bytes
	data, err = inter.tcpConnection.receiveAll(lenInt)
	if err != nil {
		return err
	}

	// Decrypt received data
	if inter.decrypt != nil {
		inter.decrypt.EncryptDecrypt(data)
	}

	return nil
}