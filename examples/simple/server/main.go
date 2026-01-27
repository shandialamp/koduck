package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shandialamp/koduck"
)

type SayName struct {
	Name string `json:"name"`
}

type Ok struct {
	Message string `json:"message"`
}

func main() {
	server, err := koduck.NewServerWithConfig(koduck.DefaultServerConfig())
	if err != nil {
		panic(err)
	}

	server.On(koduck.ServerEventClientConnected, func(_payload koduck.EventPayload) error {
		payload := _payload.(*koduck.ServerEventClientConnectedPayload)
		fmt.Println("客户端连接: ", payload.Conn.RemoteAddr())
		return nil
	})

	server.On(koduck.ServerEventClientDisconnected, func(_payload koduck.EventPayload) error {
		payload := _payload.(*koduck.ServerEventClientDisconnectedPayload)
		fmt.Println("客户端断开连接: ", payload.ConnAddr)
		return nil
	})
	server.On(koduck.ServerEventError, func(_payload koduck.EventPayload) error {
		payload := _payload.(*koduck.ServerEventErrorPayload)
		fmt.Println("服务端错误: ", payload.Err)
		return nil
	})

	router := koduck.NewRouter()
	koduck.RegisterRoute(router, 1000, func(c *koduck.Conn, params *SayName) error {
		fmt.Println("客户端说它的名字是：" + params.Name)
		msg, _ := koduck.EncodeMessage(2000, &Ok{
			Message: "知道了",
		})
		c.Send(msg)
		return nil
	})
	server.SetRouter(router)

	go func() {
		if err := server.Start(); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("服务端准备停止")
	if err := server.Stop(); err != nil {
		panic(err)
	}
	fmt.Println("服务端停止")
}
