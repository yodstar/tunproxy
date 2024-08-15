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
	serverConnMap sync.Map
	serverTCPConn atomic.Value
)

// ServerContext
type ServerContext struct {
	socket.Context
}

// ServerTCPConn
type ServerTCPConn struct {
	*socket.Conn
}

// ServerContext.Handle
func (ctx *ServerContext) Handle(conn *socket.Conn, bufr io.Reader) (err error) {
	var data []byte
	var msg pb.Message

	if data, err = ioutil.ReadAll(bufr); err != nil {
		LOG.Error(err)
		return
	}
	if err = proto.Unmarshal(data, &msg); err != nil {
		LOG.Error(err)
		return
	}

	switch msg.MsgId {
	case -1:
		ServerHeartbeat(conn)
	case 0:
		err = ServerForwardUDP(msg.DstAddr, msg.Content)
	default:
		err = ServerForwardTCP(conn, msg.MsgId)
	}
	return
}

// ServerContext.Destroy
func (ctx *ServerContext) Destroy(conn *socket.Conn) error {
	LOG.Trace("ServerContext.Destroy(%v)", conn)

	serverTCPConn.Store(&ServerTCPConn{})
	serverConnMap.Delete(conn.RemoteAddr())
	serverConnMap.Range(func(k, v interface{}) bool {
		serverTCPConn.Store(&ServerTCPConn{Conn: v.(*socket.Conn)})
		return false
	})

	if v := serverTCPConn.Load(); v != nil {
		LOG.Debug("serverTCPConn: %v", v.(*ServerTCPConn).Conn)
	}
	return nil
}

// ServerForwardUDP
func ServerForwardUDP(taddr string, content []byte) (err error) {
	LOG.Trace("ServerForwardUDP(%s, [%d]byte)", taddr, content)

	if v, ok := forwardMap.Load(taddr); ok {
		_, err = v.(*net.UDPConn).Write(content)
	}
	return
}

// ServerForwardTCP
func ServerForwardTCP(conn net.Conn, msgid int64) error {
	if v, ok := forwardMap.Load(msgid); ok {
		remote := v.(net.Conn)
		forwardMap.Delete(msgid)

		saddr := remote.LocalAddr().String()
		if pos := strings.LastIndex(saddr, ":"); pos > 0 {
			saddr = saddr[pos:]
		}
		LOG.Trace("ServerForwardTCP(%s, %s)", saddr, conn.RemoteAddr())

		abort := make(chan error, 1)
		// conn -> remote
		go func(dst net.Conn, src net.Conn, abort chan error) {
			_, err := io.Copy(dst, src)
			abort <- err
		}(remote, conn, abort)
		// remote -> conn
		go func(dst net.Conn, src net.Conn, abort chan error) {
			_, err := io.Copy(dst, src)
			abort <- err
		}(conn, remote, abort)

		err := <-abort
		if err != nil {
			LOG.Notice(err)
		}

		remote.Close()
		return err
	}
	return nil
}

// ServerHeartbeat
func ServerHeartbeat(conn *socket.Conn) {
	_ = conn.SetDeadline(time.Now().Add(60 * time.Second))
	serverConnMap.Store(conn.RemoteAddr(), conn)
	serverTCPConn.Store(&ServerTCPConn{Conn: conn})
}

// ServerSendMessage
func ServerSendMessage(msgid int64, saddr, taddr string, content []byte) error {
	LOG.Trace("ServerSendMessage(%d, %s, %s, [%d]byte)", msgid, saddr, taddr, len(content))

	var conn *socket.Conn
	if v := serverTCPConn.Load(); v != nil {
		conn = v.(*ServerTCPConn).Conn
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

// ServerSendMessage2
func ServerSendMessage2(msgid int64, taddr string) error {
	LOG.Trace("ServerSendMessage2(%d, %s)", msgid, taddr)

	var conn *socket.Conn
	if v := serverTCPConn.Load(); v != nil {
		conn = v.(*ServerTCPConn).Conn
	}

	if conn == nil {
		return fmt.Errorf("not connected: %s", taddr)
	}

	msg := pb.Message{
		MsgId:   msgid,
		DstAddr: taddr,
		SrcAddr: conn.LocalAddr().String(),
	}

	err := socket.Protobuf.Send(conn, &msg)
	if err != nil {
		LOG.Error(err)
	}
	return err
}

// ServerSendMessage3
func ServerSendMessage3(msgid int64, saddr, taddr string) error {
	LOG.Trace("ServerSendMessage3(%d, %s, %s)", msgid, saddr, taddr)

	var conn *socket.Conn
	if v := serverTCPConn.Load(); v != nil {
		conn = v.(*ServerTCPConn).Conn
	}

	if conn == nil {
		return fmt.Errorf("not connected: %s", taddr)
	}

	msg := pb.Message{
		MsgId:   msgid,
		DstAddr: taddr,
		SrcAddr: saddr,
	}

	err := socket.Protobuf.Send(conn, &msg)
	if err != nil {
		LOG.Error(err)
	}
	return err
}

// ServerStartup
func ServerStartup() {
	LOG.Trace("ServerStartup()")

	if CONF.Listen == "" {
		return
	}
	go func() {
		err := socket.ListenTCP(CONF.Listen, &ServerContext{})
		if err != nil {
			LOG.Error(err)
		}
	}()
}
