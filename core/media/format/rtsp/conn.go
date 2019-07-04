package rtsp

import (
	"net"
	"time"
)

type connWithTimeout struct {
	Timeout time.Duration
	net.Conn
}

func (conn connWithTimeout) Read(p []byte) (n int, err error) {
	if conn.Timeout > 0 {
		conn.Conn.SetReadDeadline(time.Now().Add(conn.Timeout))
	}
	return conn.Conn.Read(p)
}

func (conn connWithTimeout) Write(p []byte) (n int, err error) {
	if conn.Timeout > 0 {
		conn.Conn.SetWriteDeadline(time.Now().Add(conn.Timeout))
	}
	return conn.Conn.Write(p)
}
