package network

import (
	"github.com/TelegramGo/TelegramGo/internal/network/proxyType"
)

type netApplication interface {
	Connect(address string, proxyConnect *proxyType.SOCKS5Proxy) error
	Send(data []byte) error
	Receive(data []byte) error
	Close() error
}

var modes = []string {"abridged", "abridgedO", "full", "intermediate", "intermediateO"}

