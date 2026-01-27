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
	client := koduck.NewClientWithConfig(koduck.DefaultClientLog())

	router := koduck.NewRouter()
	koduck.RegisterRoute(router, 2000, func(c *koduck.Conn, params *Ok) error {
		fmt.Println("服务端说：" + params.Message)
		return nil
	})
	client.SetRouter(router)

	client.On(koduck.ClientEventConnected, func(_payload koduck.EventPayload) error {
		payload := _payload.(*koduck.ClientEventConnectedPayload)
		fmt.Println("连接成功: ", payload.Conn.RemoteAddr())
		msg, _ := koduck.EncodeMessage(1000, &SayName{
			Name: "lkg",
		})
		if err := client.GetConn().Send(msg); err != nil {
			return err
		}
		return nil
	})
	client.On(koduck.ClientEventDisconnected, func(_payload koduck.EventPayload) error {
		payload := _payload.(*koduck.ClientEventDisconnectedPayload)
		fmt.Println("断开连接: ", payload.ConnAddr)
		return nil
	})

	if err := client.Start(); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("客户端准备停止")
	client.Stop()
	fmt.Println("客户端停止")
}
