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
	server, err := koduck.NewServerWithConfig(koduck.DefaultServerConfig())
	if err != nil {
		panic(err)
	}

	server.On(koduck.ServerEventClientConnected, func(_ koduck.EventPayload) error {
		return nil
	})

	server.On(koduck.ServerEventClientDisconnected, func(_ koduck.EventPayload) error {
		return nil
	})
	server.On(koduck.ServerEventError, func(_ koduck.EventPayload) error {
		return nil
	})

	router := koduck.NewRouter()
	koduck.RegisterRoute(router, 1000, func(c *koduck.Conn, _ *SayName) error {
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
	if err := server.Stop(); err != nil {
		panic(err)
	}
}
