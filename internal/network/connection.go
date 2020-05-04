package network

import (
	"github.com/TelegramGo/TelegramGo/internal/network/proxyType"
)

type netApplication interface {
	connect(address string, proxyConnect *proxyType.SOCKS5Proxy) error
	send(data []byte) error
	receive(data []byte) error
	close() error
}

var modes = []string {"abridged", "abridgedO", "full", "intermediate", "intermediateO"}

