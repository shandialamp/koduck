package main

import (
	"fmt"

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

	if err := server.Start(); err != nil {
		panic(err)
	}
}
