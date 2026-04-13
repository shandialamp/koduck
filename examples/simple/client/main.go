package main

import (
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
	client := koduck.NewClientWithConfig(koduck.DefaultClientConfig())

	router := koduck.NewRouter()
	koduck.RegisterRoute(router, 2000, func(c *koduck.Conn, _ *Ok) error {
		return nil
	})
	client.SetRouter(router)

	client.On(koduck.ClientEventConnected, func(_ koduck.EventPayload) error {
		msg, _ := koduck.EncodeMessage(1000, &SayName{
			Name: "lkg",
		})
		if err := client.GetConn().Send(msg); err != nil {
			return err
		}
		return nil
	})
	client.On(koduck.ClientEventDisconnected, func(_ koduck.EventPayload) error {
		return nil
	})

	if err := client.Start(); err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	client.Stop()
}
