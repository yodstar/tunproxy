package service

import (
	"github.com/yodstar/goutil/logger"
	"github.com/yodstar/goutil/socket"

	"tunproxy/config"
)

const (
	dstaddrUDP = "udp://"
	dstaddrTCP = "tcp://"
)

var (
	CONF = config.CONF
	LOG  = logger.LOG
)

// Startup
func Startup() {
	LOG.Info("Tunnel Proxy Startup")

	// FramePrefix
	socket.FramePrefix = []byte{'F', 'I', 'N', 'D', 'G', 'O', '-', 'P', 'R', 'O', 'X', 'Y', '0', '1'}

	ServerStartup()
	ClientWatcher()
	ForwardStartup()
	select {}
}
