package service

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yodstar/goutil/socket"
	"google.golang.org/protobuf/proto"

	"tunproxy/service/pb"
)

var (
	clientUDPConn sync.Map
	clientTCPConn atomic.Value
)

// ClientContext
type ClientContext struct {
	socket.Context
	MsgID   int64
	DstAddr string
}

// ClientTCPConn
type ClientTCPConn struct {
	*socket.Conn
}

// ClientContext.Handshake
func (ctx *ClientContext) Handshake(conn *socket.Conn) (err error) {
	msg := pb.Message{
		MsgId:   ctx.MsgID,
		DstAddr: ctx.DstAddr,
		SrcAddr: conn.LocalAddr().String(),
	}
	if err = socket.Protobuf.Send(conn, &msg); err != nil {
		LOG.Error(err)
		return
	}
	if ctx.MsgID == -1 {
		clientTCPConn.Store(&ClientTCPConn{Conn: conn})
	} else if ctx.MsgID > 0 {
		err = ClientForwardTCP(conn, ctx.DstAddr)
	}
	return
}

// ClientContext.Handle
func (ctx *ClientContext) Handle(conn *socket.Conn, bufr io.Reader) (err error) {
	var data []byte
	var msg pb.Message

	if data, err = ioutil.ReadAll(bufr); err != nil {
		LOG.Error(err)
		return err
	}
	if err = proto.Unmarshal(data, &msg); err != nil {
		LOG.Error(err)
		return err
	}

	if ctx.MsgID == -1 && len(msg.DstAddr) > 6 {
		taddr := msg.DstAddr[6:]
		switch msg.DstAddr[:6] {
		case dstaddrUDP:
			if len(msg.Content) > 0 {
				ClientForwardUDP(msg.SrcAddr, taddr, msg.Content)
			} else if _, ok := clientUDPConn.Load(taddr); !ok {
				ClientDialUDP(msg.SrcAddr, taddr)
			}
		case dstaddrTCP:
			go ClientDialTCP(msg.MsgId, taddr)
		}
	}
	return nil
}

// ClientContext.Destroy
func (ctx *ClientContext) Destroy(conn *socket.Conn) error {
	LOG.Trace("ClientContext.Destroy(%v)", conn)

	clientTCPConn.Store(&ClientTCPConn{})
	return nil
}

// ClientForwardUDP
func ClientForwardUDP(saddr, taddr string, content []byte) {
	LOG.Trace("ClientForwardUDP(%s, %s, [%d]byte)", saddr, taddr, len(content))

	var conn *net.UDPConn
	if v, ok := clientUDPConn.Load(taddr); !ok {
		conn = ClientDialUDP(saddr, taddr)
	} else {
		conn = v.(*net.UDPConn)
	}
	if conn != nil {
		if _, err := conn.Write(content); err != nil {
			LOG.Error(err)
		}
	}
}

// ClientForward
func ClientForwardTCP(conn net.Conn, taddr string) error {
	saddr := conn.RemoteAddr().String()
	if pos := strings.LastIndex(saddr, ":"); pos > 0 {
		saddr = saddr[pos:]
	}
	LOG.Trace("ClientForwardTCP(%s, %s)", saddr, taddr)

	addr, err := net.ResolveTCPAddr("tcp", taddr)
	if err != nil {
		return err
	}
	local, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		LOG.Warn(err)
		return err
	}

	abort := make(chan error, 1)
	// conn -> local
	go func(dst net.Conn, src net.Conn, abort chan error) {
		_, err = io.Copy(dst, src)
		abort <- err
	}(local, conn, abort)
	// local -> conn
	go func(dst net.Conn, src net.Conn, abort chan error) {
		_, err = io.Copy(dst, src)
		abort <- err
	}(conn, local, abort)

	err = <-abort
	if err != nil {
		LOG.Notice(err)
	}

	local.Close()
	return err
}

// ClientDialUDP
func ClientDialUDP(saddr, taddr string) *net.UDPConn {
	LOG.Trace("ClientDialUDP(%s, %s)", saddr, taddr)

	addr, err := net.ResolveUDPAddr("udp", taddr)
	if err != nil {
		LOG.Warn(err)
		return nil
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		LOG.Warn(err)
		return nil
	}
	clientUDPConn.Store(taddr, conn)

	go func() {
		defer func() {
			clientUDPConn.Delete(taddr)
			conn.Close()
		}()

		buf := make([]byte, 65535)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				LOG.Error(err)
			}
			if n > 0 {
				if err := ClientSendMessage(0, saddr, buf[:n]); err != nil {
					LOG.Error(err)
				}
			}
		}
	}()
	return conn
}

// ClientDialTCP
func ClientDialTCP(msgid int64, taddr string) {
	LOG.Trace("ClientDialTCP(%d, %s)", msgid, taddr)

	go func() {
		err := socket.DialTCP(CONF.Server, &ClientContext{MsgID: msgid, DstAddr: taddr})
		if err != nil && err != io.EOF {
			LOG.Warn(err)
		}
	}()
}

// ClientSendMessage
func ClientSendMessage(msgid int64, taddr string, content []byte) error {
	LOG.Trace("ClientSendMessage(%d, %s, [%d]byte)", msgid, taddr, len(content))

	var conn *socket.Conn
	if v := clientTCPConn.Load(); v != nil {
		conn = v.(*ClientTCPConn).Conn
	}

	if conn == nil {
		return fmt.Errorf("not connected: %s", taddr)
	}

	msg := pb.Message{
		MsgId:   msgid,
		DstAddr: taddr,
		SrcAddr: conn.LocalAddr().String(),
		Content: content,
	}

	err := socket.Protobuf.Send(conn, &msg)
	if err != nil {
		LOG.Error(err)
	}
	return err
}

// ClientWatcher
func ClientWatcher() {
	LOG.Trace("ClientWatcher()")

	if CONF.Server == "" {
		return
	}

	ClientDialTCP(-1, "")

	go func() {
		var conn *socket.Conn

		t := time.NewTicker(10 * time.Second)
		for {
			<-t.C
			if v := clientTCPConn.Load(); v != nil {
				conn = v.(*ClientTCPConn).Conn
			}
			if conn != nil {
				// Heartbeat
				msg := pb.Message{
					MsgId: -1,
				}
				if err := socket.Protobuf.Send(conn, &msg); err == nil {
					_ = conn.SetDeadline(time.Now().Add(60 * time.Second))
				}
			} else {
				// Reconnect
				ClientDialTCP(-1, "")
			}
		}
	}()
}
