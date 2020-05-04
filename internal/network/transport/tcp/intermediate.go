package tcp

import (
	"encoding/binary"
	"github.com/TelegramGo/TelegramGo/internal/network/proxyType"
)

// Intermediate
//
// In case 4-byte data alignment is needed, an intermediate version of the original protocol may be used.
//
// Overhead: small
// Minimum envelope length: 4 bytes
// Maximum envelope length: 4 bytes
type Intermediate struct{
	*tcp
}

func (inter *Intermediate) Connect(address string, proxyConnect *proxyType.SOCKS5Proxy) error {
	var err error
	inter.tcp, err = tcpNew(proxyConnect)
	if err != nil {
		return err
	}

	err = inter.tcp.connect(address)
	if err != nil {
		return err
	}

	// Telegram docs:
	// Before sending anything into the underlying socket (see transports), the client must first send 0xeeeeeeee as the first int (four bytes, the server will not send 0xeeeeeeee as the first int in the first reply).	err = inter.tcp.sendAll([]byte{0xEF})
	err = inter.tcp.sendAll([]byte{0xEE, 0xEE, 0xEE, 0xEE})
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
func (inter *Intermediate) Send(data []byte) error {
	// Parse length to 4 bytes slice
	length := make([]byte, 4)
	binary.LittleEndian.PutUint32(length, uint32(len(data)))

	// Length: payload length encoded as 4 length bytes (little endian)
	// Payload: the MTProto payload
	err := inter.tcp.sendAll(append(length, data...))
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
	length, err := inter.tcp.receiveAll(4)
	if err != nil {
		return err
	}

	// Get length of data as int
	lenInt := int(binary.LittleEndian.Uint32(length))

	// Get n bytes
	data, err = inter.tcp.receiveAll(lenInt)
	if err != nil {
		return err
	}

	return nil
}