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

// Abridged
//
// The lightest protocol available.
//
// Overhead: Very small
// Minimum envelope length: 1 byte
// Maximum envelope length: 4 bytes
type Abridged struct{
	*tcpConnection                  // TCP connection (even with a proxy)
	encrypt, decrypt *aes.AES256CTR // AES-256 CTR encrypt/decrypt (only with obfuscation true)
}

// Abridged TCP transport (or obfuscated)
func (abr *Abridged) Connect(address string, obfuscation bool) error {
	abr.tcpConnection = tcpNew()

	err := abr.tcpConnection.connect(address)
	if err != nil {
		return err
	}

	if obfuscation {
		nonce, reversedNonce := obfuscationCTRGenerator(0xEF)

		abr.encrypt = aes.AES256CTRNew(nonce[8:40], nonce[40:56])
		abr.decrypt = aes.AES256CTRNew(reversedNonce[8:40], reversedNonce[40:56])

		// Add aes encrypted to nonce
		aesNonce := make([]byte, 64); copy(aesNonce, nonce)
		abr.encrypt.EncryptDecrypt(aesNonce)

		// Send encrypted nonce to server (when connect)
		err = abr.tcpConnection.sendAll(append(nonce[:56], aesNonce[56:64]...))
	} else {
		// Telegram docs:
		// Before sending anything into the underlying socket (see transports), the client must first send 0xef as the first byte (the server will not send 0xef as the first byte in the first reply).
		err = abr.tcpConnection.sendAll([]byte{0xEF})
	}

	// Check for errors
	if err != nil {
		return err
	}

	return nil
}

// Send data to Telegram server using Abridged TCP
//
// +-+----...----+
// |l|  payload  |
// +-+----...----+
// OR
// +-+---+----...----+
// |h|len|  payload  +
// +-+---+----...----+
func (abr *Abridged) Send(data []byte) error {
	length := uint32(len(data)/4)
	if length >= 127 {
		// Parse length to 3 bytes slice
		lenB := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenB, length)
		lenB = lenB[:len(lenB)-1]

		// If the packet length divided by four is bigger than or equal to 127 (>= 0x7f), the following envelope must be used, instead:
		// Header: A single byte of value 0x7f
		// Length: payload length, divided by four, and encoded as 3 length bytes (little endian)
		// Payload: the MTProto payload
		lenB = append([]byte{0xEF}, lenB...)
		data = append(lenB, data...)
	} else {
		// Length: payload length, divided by four, and encoded as a single byte, only if the resulting packet length is a value between 0x01..0x7e.
		// Payload: the MTProto payload
		data = append([]byte{byte(length)}, data...)
	}

	// If is obfuscated, encrypt the data
	if abr.encrypt != nil {
		abr.encrypt.EncryptDecrypt(data)
	}

	// Send all bytes
	err := abr.tcpConnection.sendAll(data)
	if err != nil {
		return err
	}

	return nil
}

// Receive n data to Telegram server using Abridged TCP
//
// +-+----...----+
// |l|  payload  |
// +-+----...----+
// OR
// +-+---+----...----+
// |h|len|  payload  +
// +-+---+----...----+
func (abr *Abridged) Receive(data []byte) error {
	length, err := abr.tcpConnection.receiveAll(1)
	if err != nil {
		return err
	}

	if abr.decrypt != nil {
		abr.decrypt.EncryptDecrypt(length)
	}

	if length[0] == 0x7F {
		length, err = abr.tcpConnection.receiveAll(3)
		if err != nil {
			return err
		}

		if abr.decrypt != nil {
			abr.decrypt.EncryptDecrypt(length)
		}
	}

	// Set slice length to 4
	if len(length) < 4 {
		l := make([]byte, 4 - len(length))
		length = append(l, length...)
	}

	// Get length of data as int
	lenInt := int(binary.LittleEndian.Uint32(length) * 4)

	// Get n bytes
	data, err = abr.tcpConnection.receiveAll(lenInt)
	if err != nil {
		return err
	}

	// Decrypt data if obfuscation is enabled
	if abr.decrypt != nil {
		abr.decrypt.EncryptDecrypt(data)
	}

	return nil
}