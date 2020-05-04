package tcp

import "github.com/TelegramGo/TelegramGo/internal/network/proxyType"

type Abridged struct{
	*tcp
}

func (abr *Abridged) connect(address string, proxyConnect *proxyType.SOCKS5Proxy) error {
	var err error
	abr.tcp, err = tcpNew(proxyConnect)
	if err != nil {
		return err
	}


	err := abr.tcp.connect(address)
	if err != nil {
		return err
	}

	return nil
}
