package koduck

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"
)

type Client struct {
	config    ClientConfig
	conn      *Conn
	router    *Router
	eventBus  *EventBus
	scheduler *Scheduler
	mu        sync.Mutex
	stopped   bool
}

func NewClientWithConfig(cnf ClientConfig) *Client {
	c := &Client{
		config:    cnf,
		conn:      nil,
		eventBus:  NewEventBus(),
		scheduler: NewScheduler(),
	}
	return c
}

func (c *Client) Start() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.config.Addr)
	if err != nil {
		c.eventBus.Publish(ClientEventError, &ClientEventErrorPayload{
			Err: err,
		})
		return err
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		return err
	}
	c.conn = NewConn(0, conn)
	c.eventBus.Publish(ClientEventConnected, &ClientEventConnectedPayload{
		Conn:       c.conn,
		ServerAddr: c.config.Addr,
		LocalAddr:  conn.LocalAddr().String(),
		Time:       time.Now(),
	})
	c.startHeartbeat(5 * time.Second)
	go c.scheduler.Start()
	go c.handleConn(c.conn)
	return nil
}

func (c *Client) handleConn(conn *Conn) {
	defer func() {
		c.mu.Lock()
		alreadyStopped := c.stopped
		c.mu.Unlock()

		if !alreadyStopped {
			c.eventBus.Publish(ClientEventDisconnected, &ClientEventDisconnectedPayload{
				ServerAddr: c.config.Addr,
				Reason:     "connection closed unexpectedly",
				Time:       time.Now(),
			})
		}
		c.conn.Close()
	}()

	for {
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn.Conn, lenBuf); err != nil {
			c.eventBus.Publish(ClientEventError, &ClientEventErrorPayload{
				Err: err,
			})
			return
		}

		msgLen := binary.BigEndian.Uint32(lenBuf)
		data := make([]byte, msgLen)
		if _, err := io.ReadFull(conn.Conn, data); err != nil {
			c.eventBus.Publish(ClientEventError, &ClientEventErrorPayload{
				Err: err,
			})
			return
		}

		msg, err := DecodeMessage(data)
		if err != nil {
			c.eventBus.Publish(ClientEventError, &ClientEventErrorPayload{
				Err: err,
			})
			continue
		}
		msg.Conn = conn
		c.router.Handle(msg)
	}
}

func (c *Client) SetRouter(r *Router) {
	c.router = r
}

func (c *Client) On(eventName int, handler func(payload EventPayload) error) {
	c.eventBus.Subscribe(eventName, handler)
}

func (c *Client) Send(msg *Message) error {
	return c.conn.Send(msg)
}

func (c *Client) GetConn() *Conn {
	return c.conn
}

func (c *Client) startHeartbeat(interval time.Duration) {
	c.scheduler.Every(interval, func() {
		c.conn.Send(&Message{
			Method: ClientHeartbeat,
		})
	})
}

func (c *Client) Stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	c.mu.Unlock()

	// 发布客户端主动断开事件
	c.eventBus.Publish(ClientEventDisconnected, &ClientEventDisconnectedPayload{
		ServerAddr: c.config.Addr,
		Reason:     "client stopped",
		Time:       time.Now(),
	})

	// 关闭连接
	if c.conn != nil {
		c.conn.Close()
	}

	// 停止调度器
	c.scheduler.Stop()
}
