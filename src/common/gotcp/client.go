package gotcp

import (
	"mae_proj/MAE/common/logging"
	"net"
	"sync"
)

type Client struct {
	cw   *ConnWraper
	Conn *Conn
}

// NewServer creates a server
func NewClient(config *Config, callback ConnCallback, protocol Protocol) *Client {
	return &Client{
		cw: &ConnWraper{
			config:    config,
			callback:  callback,
			protocol:  protocol,
			exitChan:  make(chan struct{}),
			waitGroup: &sync.WaitGroup{},
		},
	}
}

// Start starts service
func (s *Client) Start(hostAndPort string) bool {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", hostAndPort)

	if err != nil {
		logging.Error("ResolveTCPAddr failed,hostAndPort=%s", hostAndPort)
		return false
	}
	conn, err2 := net.DialTCP("tcp", nil, tcpAddr)
	if err2 != nil {
		logging.Error("DialTCP failed,hostAndPort=%s", hostAndPort)
		return false
	}

	s.Conn = newConn(conn, s.cw)
	s.Conn.Do()
	return true
}

// Stop stops service
func (s *Client) Stop() {
	close(s.cw.exitChan)
	s.cw.waitGroup.Wait()
}
