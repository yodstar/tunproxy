package service

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	forwardMap sync.Map
	forwardMid int64
)

// ForwardStartup
func ForwardStartup() {
	LOG.Trace("ForwardStartup()")

	forward := make(map[string]string)
	for k, v := range CONF.Forward {
		if len(v) > 6 {
			switch v[:6] {
			case dstaddrUDP:
				go ForwardUDPServer(k, v)
				forward[k] = v
			case dstaddrTCP:
				go ForwardTCPServer(k, v)
			}
		}
	}

	go func() {
		t := time.NewTicker(10 * time.Second)
		for {
			<-t.C
			for k, v := range forward {
				if err := ServerSendMessage3(0, k, v); err != nil {
					LOG.Notice(err)
				}
			}
		}
	}()
}

// ForwardUDPServer
func ForwardUDPServer(saddr, taddr string) {
	LOG.Trace("ForwardUDPServer(%s, %s)", saddr, taddr)

	addr, err := net.ResolveUDPAddr("udp", saddr)
	if err != nil {
		LOG.Error(err)
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		LOG.Error(err)
		return
	}
	defer func() {
		forwardMap.Delete(saddr)
		conn.Close()
	}()
	forwardMap.Store(saddr, conn)

	if err := ServerSendMessage3(0, saddr, taddr); err != nil {
		LOG.Notice(err)
	}

	buf := make([]byte, 65535)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			LOG.Error(err)
		}
		if n > 0 {
			if err := ServerSendMessage(0, saddr, taddr, buf[:n]); err != nil {
				LOG.Error(err)
			}
		}
	}
}

// ForwardTCPServer
func ForwardTCPServer(saddr, taddr string) {
	LOG.Trace("ForwardTCPServer(%s, %s)", saddr, taddr)

	addr, err := net.ResolveTCPAddr("tcp", saddr)
	if err != nil {
		LOG.Error(err)
		return
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		LOG.Error(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			LOG.Warn(err)
			continue
		}
		msgid := atomic.AddInt64(&forwardMid, 1)
		forwardMap.Store(msgid, conn)
		// ServerSendMessage2
		if err := ServerSendMessage2(msgid, taddr); err != nil {
			forwardMap.Delete(msgid)
			if err = conn.Close(); err != nil {
				LOG.Error(err)
			}
		}
	}
}
