package koduck

import (
	"encoding/json"
	"fmt"
	"log"
)

const ServerHeartbeat = 1
const ClientHeartbeat = 2

// HandlerFunc 定义消息处理函数类型
type HandlerFunc func(*Message) error

// Router 路由结构体，包含路由表
type Router struct {
	routes map[uint32]HandlerFunc
}

// New 创建一个新的 Router 实例
func NewRouter() *Router {
	return &Router{
		routes: make(map[uint32]HandlerFunc),
	}
}

// Register 注册一个路由，绑定一个命令到处理函数
func (r *Router) Register(method uint32, handler HandlerFunc) {
	log.Printf("注册路由: %d", method)
	r.routes[method] = handler
}

// Handle 处理传入的消息
func (r *Router) Handle(msg *Message) error {
	if handler, ok := r.routes[msg.Method]; ok {
		return handler(msg)
	}
	return fmt.Errorf("未找到命令: %d", msg.Method)
}

func RegisterRoute[T any](router *Router, method uint32, handle func(*Conn, *T) error) {
	router.Register(method, func(msg *Message) error {
		var params T
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			return err
		}
		return handle(msg.Conn, &params)
	})
}
