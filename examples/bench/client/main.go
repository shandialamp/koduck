package main

import (
	"flag"
	"fmt"
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
	var sendCount int64

	var wg sync.WaitGroup
	wg.Add(*clients)

	start := time.Now()

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
					atomic.AddInt64(&sendCount, 1)
				}
				return nil
			})

			if err := client.Start(); err != nil {
				fmt.Println("client start error:", err)
				return
			}

			// 等待足够时间收包
			time.Sleep(3 * time.Second)
			client.Stop()
		}(i)
	}

	// 客户端侧 TPS 打点
	go func() {
		prevAck := int64(0)
		prevTime := time.Now()
		for {
			time.Sleep(1 * time.Second)
			now := time.Now()
			cur := atomic.LoadInt64(&ackCount)
			delta := cur - prevAck
			elapsed := now.Sub(prevTime).Seconds()
			fmt.Printf("[client] ack_total=%d ack_qps=%.0f\n", cur, float64(delta)/elapsed)
			prevAck = cur
			prevTime = now
		}
	}()

	wg.Wait()
	elapsed := time.Since(start).Seconds()
	fmt.Printf("[client] sent=%d acks=%d elapsed=%.2fs avg_qps=%.0f\n", sendCount, ackCount, elapsed, float64(ackCount)/elapsed)
}
