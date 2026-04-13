package main

import (
	"flag"
	"sync/atomic"

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

	if err := server.Start(); err != nil {
		panic(err)
	}
}
