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

// https://core.telegram.org/mtproto/mtproto-transports
package tcp

import (
	"encoding/binary"
	"errors"
	"github.com/TelegramGo/TelegramGo/GoombaGram/Proxy"
	"golang.org/x/net/proxy"
	"math/rand"
	"net"
)

type tcp struct{
	net.Conn
	proxy.Dialer
}

func tcpNew(proxyConnect *Proxy.SOCKS5Proxy) (*tcp, error) {
	tcpNew := new(tcp)

	if proxyConnect != nil {
		var authentication *proxy.Auth = nil
		if proxyConnect.Username != "" && proxyConnect.Password != "" {
			authentication.User = proxyConnect.Username
			authentication.Password = proxyConnect.Password
		}

		socksProxy, err := proxy.SOCKS5("tcp", proxyConnect.ProxyIP, authentication, proxy.Direct)
		if err != nil {
			return nil, err
		}

		tcpNew.Dialer = socksProxy
	}

	return tcpNew, nil
}

func (tcp *tcp) connect(address string) error {
	var err error
	if tcp.Dialer != nil {
		tcp.Conn, err = tcp.Dialer.Dial("tcp", address)
	} else {
		tcp.Conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return err
	}

	return nil
}

func (tcp *tcp) sendAll(data []byte) error {
	if tcp.Conn == nil {
		return errors.New("tcp hasn't been connected")
	}

	_, err := tcp.Conn.Write(data)

	if err != nil {
		return err
	}

	return nil
}

func (tcp *tcp) receiveAll(length int) ([]byte, error) {
	if tcp.Conn == nil {
		return nil, errors.New("tcp hasn't been connected")
	}

	data := make([]byte, length)
	num, err := tcp.Conn.Read(data)

	if err != nil {
		return nil, err
	}

	if length > num {
		return nil, errors.New("some bytes are missing")
	}

	return data, nil
}

func (tcp *tcp) close() error {
	if tcp.Conn == nil {
		return errors.New("tcp hasn't been connected")
	}

	return tcp.Conn.Close()
}

func obfuscationCTRGenerator (protocol byte) ([]byte, []byte) {
	nonce := make([]byte, 64)
	for {
		// 64 random bytes
		rand.Read(nonce)

		// first byte different from 0xEF, bytes 0-4 different from "HEAD", "POST", "GET", "PVrG", 0xEEEEEEEE, 0xDDDDDD, bytes 4-8 different from 0x00000000
		firstFourInt := binary.LittleEndian.Uint32(nonce[:4])
		if nonce[0] != 0xEF && firstFourInt != 0x44414548 && firstFourInt != 0x54534F50 && firstFourInt != 0x20544547 && firstFourInt != 0x4954504F && firstFourInt != 0xDDDDDDDD && firstFourInt != 0xEEEEEEEE && binary.LittleEndian.Uint32(nonce[4:8]) == 0x00000000 {
			nonce[56] = protocol; nonce[57] = protocol; nonce[58] = protocol; nonce[59] = protocol
			break
		}
	}

	reversedNonce := make([]byte, 64)
	// Reverse nonce
	for i := 63; i >= 0; i-- {
		reversedNonce[63-i] = nonce[i]
	}

	return nonce, reversedNonce
}