// https://core.telegram.org/mtproto/mtproto-transports
package tcp

import (
	"errors"
	"github.com/TelegramGo/TelegramGo/internal/network/proxyType"
	"golang.org/x/net/proxy"
	"net"
)

type tcp struct{
	net.Conn
	proxy.Dialer
}

func tcpNew(proxyConnect *proxyType.SOCKS5Proxy) (*tcp, error) {
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

func (tcp *tcp) receiveAll(data []byte) error {
	if tcp.Conn == nil {
		return errors.New("tcp hasn't been connected")
	}

	_, err := tcp.Conn.Read(data)

	if err != nil {
		return err
	}

	return nil
}

func (tcp *tcp) close() error {
	if tcp.Conn == nil {
		return errors.New("tcp hasn't been connected")
	}

	return tcp.Conn.Close()
}