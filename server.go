package koduck

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type Server struct {
	config    ServerConfig
	conns     sync.Map
	nextID    uint64
	router    *Router
	scheduler *Scheduler
	eventBus  *EventBus
	msgQueue  chan *Message
	pool      *ants.Pool
	bufPool   sync.Pool
	shutdownC chan struct{}
	listener  net.Listener
	mu        sync.Mutex
}

type WorkItem struct {
	Conn *Conn
	Msg  *Message
}

func NewServerWithConfig(config ServerConfig) (*Server, error) {
	pool, err := ants.NewPool(config.PoolSize, ants.WithPreAlloc(true))
	if err != nil {
		return nil, err
	}
	scheduler := NewScheduler()
	return &Server{
		config:   config,
		msgQueue: make(chan *Message, config.MsgQueueSize),
		pool:     pool,
		bufPool: sync.Pool{
			New: func() any {
				return make([]byte, config.BufSize)
			},
		},
		shutdownC: make(chan struct{}),
		scheduler: scheduler,
		eventBus:  NewEventBus(),
	}, nil
}

func (s *Server) SetRouter(r *Router) {
	s.router = r
	s.registerDefaultRoute()
}

func (s *Server) allocateConnID() uint64 {
	s.nextID++
	return s.nextID
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.listener = ln
	s.mu.Unlock()
	s.eventBus.Publish(ServerEventStarted, nil)

	s.startHeartbeat(30 * time.Second)
	go s.scheduler.Start()
	go func() {
		for msg := range s.msgQueue {
			m := msg
			s.pool.Submit(func() {
				err := s.router.Handle(m)
				if err != nil {
					s.eventBus.Publish(ServerEventError, &ServerEventErrorPayload{
						Err: err,
					})
				}
			})
		}
	}()

	// 接收连接
	for {
		connSock, err := ln.Accept()
		if err != nil {
			select {
			case <-s.shutdownC:
				return nil
			default:
				s.eventBus.Publish(ServerEventError, &ServerEventErrorPayload{
					Err: err,
				})
				continue
			}
		}

		id := s.allocateConnID()
		c := NewConn(id, connSock)
		s.conns.Store(id, c)
		s.eventBus.Publish(ServerEventClientConnected, &ServerEventClientConnectedPayload{
			Conn: c,
			Time: time.Now(),
		})
		go s.handleConn(c)
	}
}

func (s *Server) handleConn(c *Conn) {
	defer func() {
		s.eventBus.Publish(ServerEventClientDisconnected, &ServerEventClientDisconnectedPayload{
			ConnAddr: c.RemoteAddr(),
			Time:     time.Now(),
		})
		s.conns.Delete(c.ID)
		c.Close()
	}()

	lenBuf := make([]byte, 4)
	for {
		if _, err := io.ReadFull(c.Conn, lenBuf); err != nil {
			s.eventBus.Publish(ServerEventError, &ServerEventErrorPayload{
				Err: err,
			})
			return
		}

		msgLen := binary.BigEndian.Uint32(lenBuf)
		buf := s.bufPool.Get().([]byte)
		if int(msgLen) > cap(buf) {
			// 如果不够大，重新分配
			buf = make([]byte, msgLen)
		}
		data := buf[:msgLen]
		if _, err := io.ReadFull(c.Conn, data); err != nil {
			s.eventBus.Publish(ServerEventError, &ServerEventErrorPayload{
				Err: err,
			})
			s.bufPool.Put(buf)
			return
		}

		msg, err := DecodeMessage(data)
		s.bufPool.Put(buf)
		if err != nil {
			s.eventBus.Publish(ServerEventError, &ServerEventErrorPayload{
				Err: err,
			})
			continue
		}
		msg.Conn = c
		s.msgQueue <- msg
	}
}

func (s *Server) Broadcast(msg *Message, filter func(*Conn) bool) {
	s.conns.Range(func(_, value any) bool {
		if c, ok := value.(*Conn); ok {
			if filter != nil && filter(c) {
				if err := c.Send(msg); err != nil {
					s.eventBus.Publish(ServerEventError, &ServerEventErrorPayload{
						Err: err,
					})
				}
			}
		}
		return true
	})
}

func (s *Server) Every(interval time.Duration, handler func()) {
	s.scheduler.Every(interval, handler)
}

func (s *Server) On(eventName int, handler func(payload EventPayload) error) {
	s.eventBus.Subscribe(eventName, handler)
}

func (s *Server) startHeartbeat(interval time.Duration) {
	s.scheduler.Every(interval, func() {
		now := time.Now().UnixNano()

		s.conns.Range(func(key, value any) bool {
			c := value.(*Conn)

			if !c.CheckHeartbeat(now) {
				s.eventBus.Publish(ServerEventError, &ServerEventErrorPayload{
					Err: errors.New("心跳超时，关闭连接"),
				})
				c.CloseSafely()
				s.conns.Delete(c.ID)
			}
			return true
		})
	})
}

func (s *Server) registerDefaultRoute() {
	RegisterRoute(s.router, ClientHeartbeat, func(c *Conn, t *string) error {
		c.UpdateHeartbeat()
		return nil
	})
}

func (s *Server) FindConn(handle func(c *Conn) bool) *Conn {
	var result *Conn
	s.conns.Range(func(key, value any) bool {
		c := value.(*Conn)
		if handle(c) {
			result = c
			return false
		}
		return true
	})
	return result
}

// Stop 优雅关闭服务器
func (s *Server) Stop() error {
	// 关闭 shutdown 信号
	select {
	case <-s.shutdownC:
		return nil // 已经关闭
	default:
		close(s.shutdownC)
	}

	// 关闭 listener 停止接受新连接
	s.mu.Lock()
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Unlock()

	// 关闭所有现有连接
	s.conns.Range(func(key, value any) bool {
		if c, ok := value.(*Conn); ok {
			c.Close()
		}
		return true
	})

	// 停止调度器
	s.scheduler.Stop()

	// 关闭消息队列
	close(s.msgQueue)

	// 释放工作池
	s.pool.Release()

	return nil
}
