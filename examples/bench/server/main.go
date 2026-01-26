package main

import (
	"flag"
	"fmt"
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
	addr := flag.String("addr", ":10001", "server listen address")
	echo := flag.Bool("echo", true, "whether to echo ack(3001)")
	flag.Parse()

	server, err := koduck.NewServerWithConfig(koduck.ServerConfig{
		Addr:         *addr,
		MsgQueueSize: 10000,
		PoolSize:     5000,
		BufSize:      8192,
	})
	if err != nil {
		panic(err)
	}

	var recvCount int64
	var lastCount int64
	var lastTime = time.Now()

	router := koduck.NewRouter()
	koduck.RegisterRoute(router, 3000, func(c *koduck.Conn, p *Ping) error {
		atomic.AddInt64(&recvCount, 1)
		if *echo {
			msg, _ := koduck.EncodeMessage(3001, &Ack{Seq: p.Seq})
			c.Send(msg)
		}
		return nil
	})
	server.SetRouter(router)

	server.Every(1*time.Second, func() {
		now := time.Now()
		total := atomic.LoadInt64(&recvCount)
		delta := total - lastCount
		elapsed := now.Sub(lastTime).Seconds()
		qps := float64(delta) / elapsed
		fmt.Printf("[server] total=%d qps=%.0f\n", total, qps)
		lastCount = total
		lastTime = now
	})

	if err := server.Start(); err != nil {
		panic(err)
	}
}
