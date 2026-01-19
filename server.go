package koduck

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
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
	Log       *Log
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
		Log:      NewLog(config.Log),
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
	s.Log.Access("server started", zap.String("addr", s.config.Addr))

	s.startHeartbeat(30 * time.Second)
	go s.scheduler.Start()
	go func() {
		for msg := range s.msgQueue {
			m := msg
			s.pool.Submit(func() {
				err := s.router.Handle(m)
				if err != nil {
					s.Log.Error(err, "router handle error")
				}
			})
		}
	}()

	// 接收连接
	for {
		connSock, err := ln.Accept()
		s.Log.Access("client connected",
			zap.String("addr", connSock.RemoteAddr().String()),
		)
		if err != nil {
			select {
			case <-s.shutdownC:
				return nil
			default:
				s.Log.Error(err, "接收错误")
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
			ConnKey: c.String(),
			Time:    time.Now(),
		})
		s.conns.Delete(c.ID)
		c.Close()
	}()

	lenBuf := make([]byte, 4)
	for {
		if _, err := io.ReadFull(c.Conn, lenBuf); err != nil {
			s.Log.Error(err, "读取错误")
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
			s.Log.Error(err, "读取长度错误")
			s.bufPool.Put(buf)
			return
		}

		msg, err := DecodeMessage(data)
		s.bufPool.Put(buf)
		if err != nil {
			s.Log.Error(err, "message decode", zap.String("conn", c.String()), zap.String("message", string(data)))
			continue
		}
		msg.Conn = c
		s.msgQueue <- msg
	}
}

func (s *Server) Broadcast(msg *Message) {
	s.conns.Range(func(_, value any) bool {
		if c, ok := value.(*Conn); ok {
			c.Send(msg)
		}
		return true
	})
}

func (s *Server) Every(interval time.Duration, handler func()) {
	s.scheduler.Every(interval, handler)
}

func (s *Server) On(eventName string, handler func(payload EventPayload) error) {
	s.eventBus.Subscribe(eventName, handler)
}

func (s *Server) startHeartbeat(interval time.Duration) {
	s.scheduler.Every(interval, func() {
		now := time.Now().UnixNano()

		s.conns.Range(func(key, value any) bool {
			c := value.(*Conn)

			if !c.CheckHeartbeat(now) {
				s.Log.Error(nil, "心跳超时，关闭连接", zap.String("conn", c.String()))
				c.CloseSafely()
				s.conns.Delete(c.ID)
			}
			return true
		})
	})
}

func (s *Server) registerDefaultRoute() {
	RegisterRoute(s.router, ClientHeartbeat, func(c *Conn, t *string) error {
		s.Log.Access("客户端心跳", zap.String("conn", c.String()))
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
