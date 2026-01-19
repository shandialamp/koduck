package koduck

import (
	"encoding/binary"
	"io"
	"net"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	config    ClientConfig
	conn      *Conn
	router    *Router
	eventBus  *EventBus
	scheduler *Scheduler
	Log       *Log
}

func NewClientWithConfig(cnf ClientConfig) *Client {
	c := &Client{
		config:    cnf,
		conn:      nil,
		eventBus:  NewEventBus(),
		scheduler: NewScheduler(),
		Log:       NewLog(cnf.Log),
	}
	return c
}

func (c *Client) Start() error {
	conn, err := net.Dial("tcp", c.config.Addr)
	if err != nil {
		return err
	}
	c.conn = NewConn(0, conn)
	c.Log.Access("连接成功")
	c.eventBus.Publish(ClientEventConnected, &ClientEventConnectedPayload{})
	c.startHeartbeat(5 * time.Second)
	go c.scheduler.Start()
	go c.handleConn(c.conn)
	return nil
}

func (c *Client) handleConn(conn *Conn) {
	defer func() {
		c.Log.Access("断开连接")
		c.conn.Close()
	}()

	for {
		lenBuf := make([]byte, 4)
		if _, err := io.ReadFull(conn.Conn, lenBuf); err != nil {
			c.Log.Error(err, "读取错误", zap.String("conn", conn.String()))
			return
		}

		msgLen := binary.BigEndian.Uint32(lenBuf)
		data := make([]byte, msgLen)
		if _, err := io.ReadFull(conn.Conn, data); err != nil {
			c.Log.Error(err, "读取消息长度错误", zap.String("conn", conn.String()))
			return
		}

		msg, err := DecodeMessage(data)
		if err != nil {
			c.Log.Error(err, "解码失败", zap.String("conn", conn.String()))
			continue
		}
		msg.Conn = conn
		c.router.Handle(msg)
	}
}

func (c *Client) SetRouter(r *Router) {
	c.router = r
}

func (c *Client) On(eventName string, handler func(payload EventPayload) error) {
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
	c.conn.Close()
}
