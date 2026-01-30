package koduck

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Conn struct {
	ID        uint64
	Conn      net.Conn
	LastBeat  int64         // 最后心跳时间，UnixNano
	Timeout   time.Duration // 心跳超时时间
	mu        sync.Mutex
	sendChan  chan []byte
	closed    chan struct{}
	closeOnce sync.Once // 防止重复关闭
	property  sync.Map
}

func NewConn(id uint64, conn net.Conn) *Conn {
	c := &Conn{
		ID:       id,
		Conn:     conn,
		LastBeat: time.Now().UnixNano(),
		sendChan: make(chan []byte, 1024),
		closed:   make(chan struct{}),
		Timeout:  30 * time.Second,
	}
	go c.writeLoop()
	return c
}

// Send 向连接发送消息
func (c *Conn) Send(msg *Message) error {
	b, err := msg.Encode()
	if err != nil {
		return err
	}
	packet := make([]byte, 4+len(b))
	binary.BigEndian.PutUint32(packet[:4], uint32(len(b)))
	copy(packet[4:], b)
	select {
	case c.sendChan <- packet:
		return nil
	case <-c.closed:
		return io.ErrClosedPipe
	}
}

// Close 关闭连接
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
		c.mu.Unlock()
		close(c.closed)
	})
}
func (c *Conn) writeLoop() {
	for {
		select {
		case data := <-c.sendChan:
			if err := c.writeFull(data); err != nil {
				// 网络异常，主动关闭连接
				c.Close()
				return
			}
		case <-c.closed:
			return
		}
	}
}

func (c *Conn) writeFull(buf []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn == nil {
		return io.ErrClosedPipe
	}

	total := 0
	for total < len(buf) {
		n, err := c.Conn.Write(buf[total:])
		if err != nil {
			return err
		}
		total += n
	}
	return nil
}

func (c *Conn) RemoteAddr() string {
	return c.Conn.RemoteAddr().String()
}

func (c *Conn) String() string {
	return fmt.Sprintf("Conn[%d] %s", c.ID, c.RemoteAddr())
}

func (c *Conn) CheckHeartbeat(now int64) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Duration(now - c.LastBeat)
	return elapsed <= c.Timeout
}

func (c *Conn) CloseSafely() {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
		c.mu.Unlock()
		close(c.closed)
	})
}

func (c *Conn) UpdateHeartbeat() {
	c.mu.Lock()
	c.LastBeat = time.Now().UnixNano()
	c.mu.Unlock()
}

// 设置属性
func (c *Conn) SetProperty(key string, value any) {
	c.property.Store(key, value)
}

// 获取属性
func (c *Conn) GetProperty(key string) (any, bool) {
	return c.property.Load(key)
}

// 删除属性
func (c *Conn) DeleteProperty(key string) {
	c.property.Delete(key)
}
