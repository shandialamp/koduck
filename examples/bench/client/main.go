package main

import (
	"flag"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shandialamp/koduck"
)

type Ping struct {
	Seq int `json:"seq"`
}

type Ack struct {
	Seq int `json:"seq"`
}

func main() {
	addr := flag.String("addr", "localhost:10001", "server address")
	clients := flag.Int("clients", 100, "number of clients")
	messages := flag.Int("messages", 1000, "messages per client")
	flag.Parse()

	var ackCount int64

	var wg sync.WaitGroup
	wg.Add(*clients)

	for i := 0; i < *clients; i++ {
		go func(id int) {
			defer wg.Done()

			client := koduck.NewClientWithConfig(koduck.ClientConfig{Addr: *addr})

			r := koduck.NewRouter()
			koduck.RegisterRoute(r, 3001, func(c *koduck.Conn, a *Ack) error {
				atomic.AddInt64(&ackCount, 1)
				return nil
			})
			client.SetRouter(r)

			client.On(koduck.ClientEventConnected, func(_ koduck.EventPayload) error {
				for m := 0; m < *messages; m++ {
					msg, _ := koduck.EncodeMessage(3000, &Ping{Seq: m})
					if err := client.GetConn().Send(msg); err != nil {
						return err
					}
				}
				return nil
			})

			if err := client.Start(); err != nil {
				panic(err)
			}

			// 等待足够时间收包
			time.Sleep(3 * time.Second)
			client.Stop()
		}(i)
	}

	wg.Wait()
}
