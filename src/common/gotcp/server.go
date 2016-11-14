package gotcp

//根据https://github.com/whiskerman/gotcp 修改

import (
	"mae_proj/MAE/common/logging"
	"net"
	"sync"
	"time"
)

type Config struct {
	PacketSendChanLimit    uint32 // the limit of packet send channel
	PacketReceiveChanLimit uint32 // the limit of packet receive channel
	ReadTimeOut            uint32
	WriteTimeOut           uint32
}

type ConnWraper struct {
	config         *Config         // server configuration
	callback       ConnCallback    // message callbacks in connection
	protocol       Protocol        // customize packet protocol
	exitChan       chan struct{}   // notify all goroutines to shutdown
	waitGroup      *sync.WaitGroup // wait for all goroutines
	needHeartBeat  bool
	hbSendInterval int64 //每隔多少秒发一次心跳，同时检测是否超时
	hbTimeout      int64 //超时时间，单位秒
}

type Server struct {
	cw *ConnWraper
}

// NewServer creates a server
func NewServer(config *Config, callback ConnCallback, protocol Protocol, hbSendInterval, hbTimeout int64) *Server {
	return &Server{
		cw: &ConnWraper{
			config:         config,
			callback:       callback,
			protocol:       protocol,
			exitChan:       make(chan struct{}),
			waitGroup:      &sync.WaitGroup{},
			needHeartBeat:  true,
			hbSendInterval: hbSendInterval,
			hbTimeout:      hbTimeout,
			
		},
	}
}

// Start starts service
func (s *Server) Start(listener *net.TCPListener, acceptTimeout time.Duration) {
	s.cw.waitGroup.Add(1)
	defer func() {
		listener.Close()
		s.cw.waitGroup.Done()
	}()

	for {
		select {
		case <-s.cw.exitChan:
			return

		default:
		}

		listener.SetDeadline(time.Now().Add(acceptTimeout))

		conn, err := listener.AcceptTCP()

		if e, ok := err.(net.Error); ok && e.Timeout() {
			continue
			// This was a timeout
		} else if err != nil {
			logging.Info("listener accepttcp continue and found a error: %v", err)
			return
			// This was an error, but not a timeout
		}

		go newConn(conn, s.cw).Do()
	}
}

// Stop stops service
func (s *Server) Stop() {
	close(s.cw.exitChan)
	s.cw.waitGroup.Wait()
}
