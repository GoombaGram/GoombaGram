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

package tl

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
)

type DecodeBuffer struct {
	buffer []byte
	off    int
	size   int
	err    error
}

// Buffer constructor from an input byte slice
func NewDecodeBuffer(input []byte) *DecodeBuffer {
	return &DecodeBuffer{
		buffer: input,
		off:    0,
		size:   len(input),
		err:    nil,
	}
}

// Return error state
func (buf *DecodeBuffer) GetError() error {
	return buf.err
}

// Parse an int64 from DecodeBuffer
func (buf *DecodeBuffer) Long() int64 {
	// Check for errors
	if buf.err != nil {
		return 0
	}

	if buf.off + 8 > buf.size {
		buf.err = errors.New("DecodeLong: too few bytes to decode")
		return 0
	}

	// Decode int64 from buffer
	long := int64(binary.LittleEndian.Uint64(buf.buffer[buf.off : buf.off + 8]))

	// Offset = Offset + 8 bytes. 64 bit -> 8 byte
	buf.off += 8

	// Return result
	return long
}

// Parse a float64 from DecodeBuffer
func (buf *DecodeBuffer) Double() float64 {
	// Check for errors
	if buf.err != nil {
		return 0
	}

	if buf.off + 8 > buf.size {
		buf.err = errors.New("DecodeDouble: too few bytes to decode")
		return 0
	}

	// Decode float64 from buffer
	double := math.Float64frombits(binary.LittleEndian.Uint64(buf.buffer[buf.off : buf.off+8]))

	// Offset = Offset + 8 bytes. 64 bit -> 8 byte
	buf.off += 8

	// Return result
	return double
}

// Parse an int32 from DecodeBuffer
func (buf *DecodeBuffer) Int() int32 {
	// Check for errors
	if buf.err != nil {
		return 0
	}

	if buf.off + 4 > buf.size {
		buf.err = errors.New("DecodeInt: too few bytes to decode")
		return 0
	}

	// Decode int32 from buffer
	intVar := binary.LittleEndian.Uint32(buf.buffer[buf.off : buf.off+4])

	// Offset = Offset + 4 bytes. 32 bit -> 4 byte
	buf.off += 4

	// Return result
	return int32(intVar)
}

// Parse an uint32 from DecodeBuffer
func (buf *DecodeBuffer) UInt() uint32 {
	// Check for errors
	if buf.err != nil {
		return 0
	}

	if buf.off + 4 > buf.size {
		buf.err = errors.New("DecodeUInt: too few bytes to decode")
		return 0
	}

	// Decode uint32 from buffer
	uintVar := binary.LittleEndian.Uint32(buf.buffer[buf.off : buf.off+4])

	// Offset = Offset + 4 bytes. 32 bit -> 4 byte
	buf.off += 4

	// Return result
	return uintVar
}

// Read n bytes from DecodeBuffer
func (buf *DecodeBuffer) Bytes(length int) []byte {
	// Check for errors
	if buf.err != nil {
		return nil
	}

	if buf.off + length > buf.size {
		buf.err = errors.New("DecodeBytes: too few bytes to decode")
		return nil
	}

	// Make an empty slice with length = int input
	bytes := make([]byte, length)

	// Copy bytes to new slice
	copy(bytes, buf.buffer[buf.off : buf.off + length])

	// Offset = Offset + read length
	buf.off += length

	// Return result
	return bytes
}

// Read a string from DecodeBuffer as byte slice
func (buf *DecodeBuffer) StringBytes() []byte {
	// Check for errors
	if buf.err != nil {
		return nil
	}

	if buf.off + 1 > buf.size {
		buf.err = errors.New("DecodeStringBytes: too few bytes to decode")
		return nil
	}

	// Get first available byte as length
	size := int(buf.buffer[buf.off])
	buf.off++

	// Calculate padding length as int (if L <= 253)
	padding := (4 - ((size + 1) % 4)) & 3

	// https://core.telegram.org/mtproto/serialize#base-types
	// If L <= 253, the serialization contains one byte with the value of L, then L bytes of the string followed by 0 to 3 characters containing 0,
	// such that the overall length of the value be divisible by 4, whereupon all of this is interpreted as a sequence of int(L/4)+1 32-bit numbers.
	//
	// If L >= 254, the serialization contains byte 254, followed by 3 bytes with the string length L, followed by L bytes of the string, further followed by 0 to 3 null padding bytes.


	if size == 254 {
		// Check for errors
		if buf.off + 3 > buf.size {
			buf.err = errors.New("DecodeStringBytes: too few bytes to decode")
			return nil
		}

		// Get string length
		size = int(buf.buffer[buf.off]) | int(buf.buffer[buf.off+1])<<8 | int(buf.buffer[buf.off+2])<<16
		buf.off += 3

		padding = (4 - size % 4) & 3
	}

	// Check for errors
	if buf.off + size > buf.size {
		buf.err = errors.New("DecodeStringBytes: Wrong size")
		return nil
	}

	// Create an empty byte slice
	stringVar := make([]byte, size)

	// Copy bytes to new slice and increase DecodeBuffer offset
	copy(stringVar, buf.buffer[buf.off:buf.off + size])
	buf.off += size

	// Check for padding size errors and increase offset
	if buf.off + padding > buf.size {
		buf.err = errors.New("DecodeStringBytes: Wrong padding")
		return nil
	}
	buf.off += padding

	// Return result
	return stringVar
}

// Read an Unicode string from DecodeBuffer
func (buf *DecodeBuffer) String() string {
	stringBytes := buf.StringBytes()

	// Check for errors
	if buf.err != nil {
		return ""
	}

	// Convert bytes to string
	stringVar := string(stringBytes)

	// Return result
	return stringVar
}

// Read a BigInt from DecodeBuffer
func (buf *DecodeBuffer) BigInt() *big.Int {
	stringBytes := buf.StringBytes()

	// Check for errors
	if buf.err != nil {
		return nil
	}

	// Create a new byte slice with index 0 = 0 followed by string bytes
	bytes := make([]byte, len(stringBytes)+1)
	bytes[0] = 0
	copy(bytes[1:], stringBytes)

	// Store all into a big int
	bigVar := new(big.Int).SetBytes(bytes)

	// Return result
	return bigVar
}

// Read an int32 vector from DecodeBuffer
func (buf *DecodeBuffer) VectorInt() []int32 {
	// Get constructor CRC
	constructor := buf.UInt()

	// Check for errors
	if buf.err != nil {
		return nil
	}
	if constructor != crcVector {
		buf.err = fmt.Errorf("DecodeVectorInt: Wrong constructor (0x%08x)", constructor)
		return nil
	}

	// Read Vector size from buffer
	size := buf.Int()
	if buf.err != nil {
		return nil
	}

	if size < 0 {
		buf.err = errors.New("DecodeVectorInt: Wrong size")
		return nil
	}

	// Make an empty slice
	intResult := make([]int32, size)

	// Fill the slice
	for i := int32(0); i < size; i++ {
		// Read an int32 from buffer
		intResult[i] = buf.Int()
		if buf.err != nil {
			return nil
		}
	}

	// Return result
	return intResult
}

// Read an int64 vector from DecodeBuffer
func (buf *DecodeBuffer) VectorLong() []int64 {
	// Get constructor CRC
	constructor := buf.UInt()

	// Check for errors
	if buf.err != nil {
		return nil
	}
	if constructor != crcVector {
		buf.err = fmt.Errorf("DecodeVectorLong: Wrong constructor (0x%08x)", constructor)
		return nil
	}

	// Read Vector size from buffer
	size := buf.Int()
	if buf.err != nil {
		return nil
	}

	if size < 0 {
		buf.err = errors.New("DecodeVectorLong: Wrong size")
		return nil
	}

	// Make an empty slice
	longResult := make([]int64, size)

	// Fill the slice
	for i := int32(0); i < size; i++ {
		// Read an int64 from buffer
		longResult[i] = buf.Long()
		if buf.err != nil {
			return nil
		}
	}

	// Return result
	return longResult
}

// Read a string vector from DecodeBuffer
func (buf *DecodeBuffer) VectorString() []string {
	// Get constructor CRC
	constructor := buf.UInt()

	// Check for errors
	if buf.err != nil {
		return nil
	}
	if constructor != crcVector {
		buf.err = fmt.Errorf("DecodeVectorString: Wrong constructor (0x%08x)", constructor)
		return nil
	}

	// Read Vector size from buffer
	size := buf.Int()
	if buf.err != nil {
		return nil
	}

	if size < 0 {
		buf.err = errors.New("DecodeVectorString: Wrong size")
		return nil
	}

	// Make an empty slice
	stringResult := make([]string, size)

	// Fill the slice
	for i := int32(0); i < size; i++ {
		// Read a string from buffer
		stringResult[i] = buf.String()
		if buf.err != nil {
			return nil
		}
	}

	// Return result
	return stringResult
}

// Read a float64 vector from DecodeBuffer
func (buf *DecodeBuffer) VectorDouble() []float64 {
	// Get constructor CRC
	constructor := buf.UInt()

	// Check for errors
	if buf.err != nil {
		return nil
	}
	if constructor != crcVector {
		buf.err = fmt.Errorf("DecodeVectorDouble: Wrong constructor (0x%08x)", constructor)
		return nil
	}

	// Read Vector size from buffer
	size := buf.Int()
	if buf.err != nil {
		return nil
	}

	if size < 0 {
		buf.err = errors.New("DecodeVectorString: Wrong size")
		return nil
	}

	// Make an empty slice
	doubleResult := make([]float64, size)

	// Fill the slice
	for i := int32(0); i < size; i++ {
		// Read a float64 from buffer
		doubleResult[i] = buf.Double()
		if buf.err != nil {
			return nil
		}
	}

	// Return result
	return doubleResult
}

// Read a boolean value from DecodeBuffer
func (buf *DecodeBuffer) Bool() bool {
	// Get constructor CRC
	constructor := buf.UInt()

	// Check for errors
	if buf.err != nil {
		return false
	}

	// Return true if constructor is equals to crcBoolTrue
	return constructor == crcBoolTrue
}

// Read a TLObject vector from DecodeBuffer
func (buf *DecodeBuffer) Vector() []TL {
	// Get constructor CRC
	constructor := buf.UInt()

	// Check for errors
	if buf.err != nil {
		return nil
	}
	if constructor != crcVector {
		buf.err = fmt.Errorf("DecodeVector: Wrong constructor (0x%08x)", constructor)
		return nil
	}

	// Read Vector size from buffer
	size := buf.Int()
	if buf.err != nil {
		return nil
	}

	if size < 0 {
		buf.err = errors.New("DecodeVector: Wrong size")
		return nil
	}

	// Make an empty slice
	tlResult := make([]TL, size)

	// Fill the slice
	for i := int32(0); i < size; i++ {
		// Read a TL Object from buffer
		tlResult[i] = buf.Object()
		if buf.err != nil {
			return nil
		}
	}

	// Return result
	return tlResult
}

/*
func toBool(x TL) bool {
	_, ok := x.(TLBoolTrue)
	return ok
}
 */

/*
func (buf *DecodeBuffer) Object() (r TL) {
	// Get constructor CRC
	constructor := buf.UInt()

	// Check for errors
	if buf.err != nil {
		return nil
	}

	switch constructor {

	case crc_resPQ:
		r = TL_resPQ{buf.Bytes(16), buf.Bytes(16), buf.BigInt(), buf.VectorLong()}

	case crc_server_DH_params_ok:
		r = TL_server_DH_params_ok{buf.Bytes(16), buf.Bytes(16), buf.StringBytes()}

	case crc_server_DH_inner_data:
		r = TL_server_DH_inner_data{
			buf.Bytes(16), buf.Bytes(16), buf.Int(),
			buf.BigInt(), buf.BigInt(), buf.Int(),
		}

	case crc_dh_gen_ok:
		r = TL_dh_gen_ok{buf.Bytes(16), buf.Bytes(16), buf.Bytes(16)}

	case crc_ping:
		r = TL_ping{buf.Long()}

	case crc_pong:
		r = TL_pong{buf.Long(), buf.Long()}

	case crc_msg_container:
		size := buf.Int()
		arr := make([]TL_MT_message, size)
		for i := int32(0); i < size; i++ {
			arr[i] = TL_MT_message{buf.Long(), buf.Int(), buf.Int(), buf.Object()}
			if buf.err != nil {
				return nil
			}
		}
		r = TL_msg_container{arr}

	case crc_rpc_result:
		r = TL_rpc_result{buf.Long(), buf.Object()}

	case crc_rpc_error:
		r = TL_rpc_error{buf.Int(), buf.String()}

	case crc_new_session_created:
		r = TL_new_session_created{buf.Long(), buf.Long(), buf.Bytes(8)}

	case crc_bad_server_salt:
		r = TL_bad_server_salt{buf.Long(), buf.Int(), buf.Int(), buf.Bytes(8)}

	case crc_bad_msg_notification:
		r = TL_crc_bad_msg_notification{buf.Long(), buf.Int(), buf.Int()}

	case crc_msgs_ack:
		r = TL_msgs_ack{buf.VectorLong()}

	case crc_gzip_packed:
		obj := make([]byte, 0, 4096)

		var buf bytes.Buffer
		_, _ = buf.Write(buf.StringBytes())
		gz, _ := gzip.NewReader(&buf)

		b := make([]byte, 4096)
		for true {
			n, _ := gz.Read(b)
			obj = append(obj, b[0:n]...)
			if n <= 0 {
				break
			}
		}
		d := NewDecodeBuf(obj)
		r = d.Object()

	default:
		r = buf.ObjectGenerated(constructor)

	}

	if buf.err != nil {
		return nil
	}

	return
}
 */