package tcp

import (
	"encoding/binary"
	"errors"
	"github.com/TelegramGo/TelegramGo/internal/crypto/aes"
	"github.com/TelegramGo/TelegramGo/internal/network/proxyType"
	"math/rand"
)

// Obfuscated Abridged TCP (to prevent ISP block)
//
// https://core.telegram.org/mtproto/mtproto-transports#transport-obfuscation
type AbridgedO struct{
	*tcp
	encrypt, decrypt *aes.AES256CTR
}

// Connect with KEY exchange
func (abrO *AbridgedO) Connect(address string, proxyConnect *proxyType.SOCKS5Proxy) error {
	var err error
	abrO.tcp, err = tcpNew(proxyConnect)
	if err != nil {
		return err
	}

	err = abrO.tcp.connect(address)
	if err != nil {
		return err
	}

	nonce := make([]byte, 64)
	for {
		// 64 random bytes
		rand.Read(nonce)

		// first byte different from EF, bytes 4-8 different from "HEAD", "POST", "GET", "OPTI", 0xEEEEEEEE, 0xDDDDDD
		firstFourInt := binary.LittleEndian.Uint32(nonce[:4])
		if nonce[0] != 0xEF && binary.LittleEndian.Uint32(nonce [4:8]) == 0x00000000 && firstFourInt != 0x44414548 && firstFourInt != 0x54534F50 && firstFourInt != 0x20544547 && firstFourInt != 0x4954504F && firstFourInt != 0xDDDDDDDD && firstFourInt != 0xEEEEEEEE {
			nonce[56] = 0xEF
			nonce[57] = 0xEF
			nonce[58] = 0xEF
			nonce[59] = 0xEF
			break
		}
	}

	reversedNonce := make([]byte, 64)
	// Reverse nonce
	for i := 63; i >= 0; i-- {
		reversedNonce[63 - i] = nonce[i]
	}

	abrO.encrypt = aes.AES256CTRNew(nonce[8:40], nonce[40:56])
	abrO.decrypt = aes.AES256CTRNew(reversedNonce[8:40], reversedNonce[40:56])

	// Add aes encrypted to nonce
	var aesNonce []byte
	copy(aesNonce, nonce)
	abrO.encrypt.EncryptDecrypt(aesNonce)
	aesNonce = aesNonce[56:64]
	for i := 0; i < len(aesNonce); i++ {
		nonce[56 + i] = aesNonce[i]
	}

	// Send encrypted nonce to server (when connect)
	err = abrO.tcp.sendAll(nonce)
	if err != nil {
		return err
	}

	return nil
}

// Send data to Telegram server using Abridged TCP (Obfuscated, using AES256-CTR)
//
// +-+----...----+
// |l|  payload  |
// +-+----...----+
// OR
// +-+---+----...----+
// |h|len|  payload  +
// +-+---+----...----+
func (abrO *AbridgedO) Send(data []byte) error {
	length := uint32(len(data)/4)

	var err error = nil
	if length >= 127 {
		// Parse length to 3 bytes slice
		lenB := make([]byte, 4)
		binary.LittleEndian.PutUint32(lenB, length)
		lenB = lenB[:len(lenB)-1]

		// If the packet length divided by four is bigger than or equal to 127 (>= 0x7f), the following envelope must be used, instead:
		// Header: A single byte of value 0x7f
		// Length: payload length, divided by four, and encoded as 3 length bytes (little endian)
		// Payload: the MTProto payload
		lenB = append([]byte{0x7F}, lenB...)
		data = append(lenB, data...)
	} else {
		// Length: payload length, divided by four, and encoded as a single byte, only if the resulting packet length is a value between 0x01..0x7e.
		// Payload: the MTProto payload
		data = append([]byte{byte(length)}, data...)
	}

	// Encrypt data
	abrO.encrypt.EncryptDecrypt(data)
	if data == nil {
		return errors.New("unable to calculate AES data")
	}

	err = abrO.tcp.sendAll(data)
	if err != nil {
		return err
	}

	return nil
}

// Receive data from Telegram server using Abridged TCP (obfuscated)
//
// +-+----...----+
// |l|  payload  |
// +-+----...----+
// OR
// +-+---+----...----+
// |h|len|  payload  +
// +-+---+----...----+
func (abrO *AbridgedO) Receive(data []byte) error {
	length, err := abrO.tcp.receiveAll(1)
	if err != nil {
		return err
	}

	abrO.decrypt.EncryptDecrypt(length)

	if length[0] == 0x7F {
		length, err := abrO.tcp.receiveAll(3)
		if err != nil {
			return err
		}
		abrO.decrypt.EncryptDecrypt(length)
	}

	// Set slice length to 4
	if len(length) < 4 {
		l := make([]byte, 4 - len(length))
		length = append(l, length...)
	}

	// Get length of data as int (x4)
	lenInt := int(binary.LittleEndian.Uint32(length))*4

	// Get n bytes
	data, err = abrO.tcp.receiveAll(lenInt)
	if err != nil {
		return err
	}
	abrO.decrypt.EncryptDecrypt(data)

	return nil
}