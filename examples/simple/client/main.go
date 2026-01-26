package main

import (
	"fmt"
	"time"

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

	client.On(koduck.ClientEventConnected, func(payload koduck.EventPayload) error {
		msg, _ := koduck.EncodeMessage(1000, &SayName{
			Name: "lkg",
		})
		if err := client.GetConn().Send(msg); err != nil {
			return err
		}
		return nil
	})

	if err := client.Start(); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Hour)
}
