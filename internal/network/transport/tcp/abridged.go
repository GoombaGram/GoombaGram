package tcp

import (
	"encoding/binary"
	"github.com/TelegramGo/TelegramGo/internal/network/proxyType"
)

// Abridged
//
// The lightest protocol available.
//
// Overhead: Very small
// Minimum envelope length: 1 byte
// Maximum envelope length: 4 bytes

type Abridged struct{
	*tcp
}

func (abr *Abridged) Connect(address string, proxyConnect *proxyType.SOCKS5Proxy) error {
	var err error
	abr.tcp, err = tcpNew(proxyConnect)
	if err != nil {
		return err
	}

	err = abr.tcp.connect(address)
	if err != nil {
		return err
	}

	// Telegram docs:
	// Before sending anything into the underlying socket (see transports), the client must first send 0xef as the first byte (the server will not send 0xef as the first byte in the first reply).
	err = abr.tcp.sendAll([]byte{0xEF})
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
		err = abr.tcp.sendAll(append(lenB, data...))
	} else {
		// Length: payload length, divided by four, and encoded as a single byte, only if the resulting packet length is a value between 0x01..0x7e.
		// Payload: the MTProto payload
		err = abr.tcp.sendAll(append([]byte{byte(length)}, data...))
	}

	if err != nil {
		return err
	}

	return nil
}

func (abr *Abridged) Receive(data []byte) error {
	length, err := abr.tcp.receiveAll(1)
	if err != nil || length == nil{
		return err
	}

	if length[0] == 0x7F {
		length, err := abr.tcp.receiveAll(3)
		if err != nil || length == nil{
			return err
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
	data, err = abr.tcp.receiveAll(lenInt)
	if err != nil {
		return err
	}

	return nil
}

// Close TCP connection
func (abr *Abridged) Close() error {
	return abr.tcp.close()
}