package gotcp

import (
	"bufio"
	"errors"
	"io"

	"net"
	"sync"
	"sync/atomic"
	"time"

	"mae_proj/MAE/common/logging"
)

// Error type
var (
	ErrConnClosing   = errors.New("use of closed network connection")
	ErrWriteBlocking = errors.New("write packet was blocking")
	ErrReadBlocking  = errors.New("read packet was blocking")
)

const defaultBufferSize = 16 * 1024
const defaultOutputBufferTimeout = 250 * time.Millisecond

// Conn exposes a set of callbacks for the various events that occur on a connection
type Conn struct {
	Owner                interface{}
	srv                  *ConnWraper
	conn                 *net.TCPConn  // the raw connection
	extraData            string        // to save extra data
	closeOnce            sync.Once     // close the conn, once, per instance
	closeFlag            int32         // close flag
	closeChan            chan struct{} // close chanel
	packetSendChan       chan Packet   // packet send chanel
	packetReceiveChan    chan Packet   // packeet receive chanel
	tickTime             int64         //上次心跳时间
	needHeartBeat        bool
	hbSendInterval       int64 //每隔多久发一次心跳，同时检测是否超时
	hbTimeout            int64
	UnCompleteReadBuffer []byte
	Reader               *bufio.Reader
	Writer               *bufio.Writer
	OutputBufferTimeout  time.Duration
	LenBuf               [4]byte
	LenSlice             []byte
	sync.RWMutex
}

// ConnCallback is an interface of methods that are used as callbacks on a connection
type ConnCallback interface {
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*Conn) bool

	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*Conn, Packet) bool

	// OnClose is called when the connection closed
	OnClose(*Conn)
}

// newConn returns a wrapper of raw conn
func newConn(conn *net.TCPConn, srv *ConnWraper) *Conn {
	c := &Conn{
		srv:                 srv,
		conn:                conn,
		closeChan:           make(chan struct{}),
		packetSendChan:      make(chan Packet, srv.config.PacketSendChanLimit),
		packetReceiveChan:   make(chan Packet, srv.config.PacketReceiveChanLimit),
		tickTime:            time.Now().Unix(),
		needHeartBeat:       srv.needHeartBeat,
		hbSendInterval:      srv.hbSendInterval,
		hbTimeout:           srv.hbTimeout,
		Reader:              bufio.NewReaderSize(conn, defaultBufferSize),
		Writer:              bufio.NewWriterSize(conn, defaultBufferSize),
		OutputBufferTimeout: defaultOutputBufferTimeout,
	}
	c.LenSlice = c.LenBuf[:]
	return c
}

func (c *Conn) ResetTick() {
	c.tickTime = time.Now().Unix()
}

// GetExtraData gets the extra data from the Conn
func (c *Conn) SetOwner(o interface{}) {
	c.Owner = o
}

// GetExtraData gets the extra data from the Conn
func (c *Conn) GetExtraData() string {
	return c.extraData
}

// PutExtraData puts the extra data with the Conn
func (c *Conn) PutExtraData(data string) {
	c.extraData = data
}

// GetRawConn returns the raw net.TCPConn from the Conn
func (c *Conn) GetRawConn() *net.TCPConn {
	return c.conn
}

// Close closes the connection
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		c.conn.Close()
		c.srv.callback.OnClose(c)
	})
}

// IsClosed indicates whether or not the connection is closed
func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

// AsyncReadPacket async reads a packet, this method will never block
func (c *Conn) AsyncReadPacket(timeout time.Duration) (Packet, error) {
	if c.IsClosed() {
		return nil, ErrConnClosing
	}

	if timeout == 0 {
		select {
		case p := <-c.packetReceiveChan:
			return p, nil

		default:
			return nil, ErrReadBlocking
		}

	} else {
		select {
		case p := <-c.packetReceiveChan:
			return p, nil

		case <-c.closeChan:
			return nil, ErrConnClosing

		case <-time.After(timeout):
			return nil, ErrReadBlocking
		}
	}
}

//同步发送，异步发送太慢
func (c *Conn) SyncWritePacket(p Packet) error {
	if c.IsClosed() {
		return ErrConnClosing
	}
	c.conn.SetWriteDeadline(time.Now().Add(time.Second * 20))
	packetstr := p.Serialize()

	c.Lock()
	_, err := c.Writer.Write(packetstr)
	c.Writer.Flush()
	c.Unlock()
	if err != nil {
		logging.Error("con  SyncWritePacket write found a error: %v", err)
		return err
	}
	return nil
}

// AsyncWritePacket async writes a packet, this method will never block
func (c *Conn) AsyncWritePacket(p Packet, timeout time.Duration) error {
	if c.IsClosed() {
		return ErrConnClosing
	}

	c.conn.SetWriteDeadline(time.Now().Add(time.Second * 20))
	packetstr := p.Serialize()

	//写到缓冲区而已
	c.Lock()
	_, err := c.Writer.Write(packetstr)
	c.Unlock()

	if err != nil {
		logging.Error("con  AsyncWritePacket write found a error: %v", err)
		return err
	}

	return nil
	/*
		if timeout == 0 {
			select {
			case c.packetSendChan <- p:
				return nil

				//default:
				//return ErrWriteBlocking
			}

		} else {
			select {
			case c.packetSendChan <- p:
				return nil

			case <-c.closeChan:
				return ErrConnClosing

			case <-time.After(timeout):
				return ErrWriteBlocking
			}
		}
	*/
}

// Do it
func (c *Conn) Do() {
	if !c.srv.callback.OnConnect(c) {
		return
	}
	//c.conn.SetDeadline(time.Now().Add(time.Second * 30))
	go c.handleLoop()
	go c.readStickPackLoop()
	//go c.readLoop()
	go c.writeStickPacketLoop()
	go c.heartbeatLoop()
}

func (c *Conn) readStickPackLoop() {
	c.srv.waitGroup.Add(1)
	defer func() {
		//recover()
		c.Close()
		c.srv.waitGroup.Done()
	}()

	//reader := bufio.NewReader(c.conn)

	//buffer := make([]byte, 1024)
	for {

		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		default:
		}
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 180))

		err := c.srv.protocol.Unpack(c, c.packetReceiveChan)

		if e, ok := err.(net.Error); ok && e.Timeout() {
			//l4g.Info("con read found a timeout error, i can do")

			continue
			// This was a timeout
		}
		if err != nil {
			if err == io.EOF {
				logging.Info("close by peer")
				return
			}
			logging.Info("con read found a error: %v", err)
			return
		}

		/*
			n, err := reader.Read(buffer)
			if e, ok := err.(net.Error); ok && e.Timeout() {
				//l4g.Info("con read found a timeout error, i can do")

				continue
				// This was a timeout
			}
			if err != nil {
				if err == io.EOF {
					logging.Info("close by peer")
					return
				}
				logging.Info("con read found a error: %v", err)
				return
			}

			if n > 0 {
				//fmt.Println("n is ========================================", n)
				//unCompleteBuffer = Unpack(append(unCompleteBuffer, buffer[:n]...), c.packetReceiveChan)

				c.srv.protocol.Unpack2(c, buffer[:n], c.packetReceiveChan)
			}
		*/
	}
}

func (c *Conn) readLoop() {
	c.srv.waitGroup.Add(1)
	defer func() {
		//recover()
		c.Close()
		c.srv.waitGroup.Done()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		default:
		}

		p, err := c.srv.protocol.ReadPacket(c.conn)
		if err != nil {
			logging.Info("con ReadPacket found a error: %v", err)
			return
		}

		c.packetReceiveChan <- p
	}
}

func (c *Conn) writeLoop() {
	c.srv.waitGroup.Add(1)
	defer func() {
		//recover()
		c.Close()
		c.srv.waitGroup.Done()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetSendChan:
			if _, err := c.conn.Write(DoPacket(p.Serialize())); err != nil {
				logging.Info("con write found a error: %v", err)
				return
			}
		}
	}
}

func (c *Conn) writeStickPacketLoop() {
	c.srv.waitGroup.Add(1)
	defer func() {
		//recover()
		c.Close()
		c.srv.waitGroup.Done()
	}()

	outputBufferTicker := time.NewTicker(c.OutputBufferTimeout)
	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return
		case <-outputBufferTicker.C: //每隔一段时间写入对端
			c.Lock()
			err := c.Writer.Flush()
			c.Unlock()
			if err != nil {
				logging.Error("conn writeStickPacketLoop fFlush failed,err=%s", err.Error())
				return
			}

			/*
				case p := <-c.packetSendChan:
					c.conn.SetWriteDeadline(time.Now().Add(time.Second * 180))
					packetstr := p.Serialize()

					//写到缓冲区而已
					c.Lock()
					_, err := c.Writer.Write(packetstr)
					c.Unlock()
					if e, ok := err.(net.Error); ok && e.Timeout() {
						//l4g.Info("con read found a timeout error, i can do")
						c.packetSendChan <- p //写回去
						continue
					}
					// This was a timeout

					if err != nil {
						logging.Info("con write found a error: %v", err)
						return
					}
			*/
		}
	}
}

func (c *Conn) heartbeatLoop() {
	c.srv.waitGroup.Add(1)
	defer func() {
		//recover()
		c.srv.waitGroup.Done()
	}()

	if !c.needHeartBeat {
		return
	}

	timercheck := time.NewTicker(time.Duration(c.hbSendInterval) * time.Second)
	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return
		case <-timercheck.C:
			curTime := time.Now().Unix()
			//logging.Debug("conn %s timeout,curTime=%d,tickTime=%d,hbTimeout=%d", c.GetExtraData(), int(curTime), int(c.tickTime), c.hbTimeout)
			if curTime >= c.tickTime+c.hbTimeout {
				logging.Error("conn %s timeout,curTime=%d,tickTime=%d,hbTimeout=%d,", c.GetExtraData(), int(curTime), int(c.tickTime), c.hbTimeout)
				c.Close()
				return
			}
			c.SyncWritePacket(c.srv.protocol.GetHeatBeatData())
		}
	}
}

func (c *Conn) handleLoop() {
	c.srv.waitGroup.Add(1)
	defer func() {
		//recover()
		c.Close()
		c.srv.waitGroup.Done()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetReceiveChan:
			//logging.Debug("receive msg:%s", string(p.Serialize()))
			if !c.srv.callback.OnMessage(c, p) {
				return
			}
		}
	}
}
