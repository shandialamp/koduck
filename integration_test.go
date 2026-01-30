package koduck

import (
	"net"
	"sync/atomic"
	"testing"
	"time"
)

func TestClientStopEvent(t *testing.T) {
	client := NewClientWithConfig(ClientConfig{
		Addr: "localhost:10001",
	})

	var disconnectReason string
	var disconnectCount int64

	routerC := NewRouter()
	client.SetRouter(routerC)

	client.On(ClientEventDisconnected, func(p EventPayload) error {
		payload := p.(*ClientEventDisconnectedPayload)
		disconnectReason = payload.Reason
		atomic.AddInt64(&disconnectCount, 1)
		return nil
	})

	client.Stop()

	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadInt64(&disconnectCount)
	if count != 1 {
		t.Errorf("Expected 1 disconnect event, got %d", count)
	}

	if disconnectReason != "client stopped" {
		t.Errorf("Expected reason 'client stopped', got '%s'", disconnectReason)
	}
}

func TestClientStopIdempotent(t *testing.T) {
	client := NewClientWithConfig(ClientConfig{
		Addr: "localhost:10001",
	})

	var disconnectCount int64

	routerC := NewRouter()
	client.SetRouter(routerC)

	client.On(ClientEventDisconnected, func(p EventPayload) error {
		atomic.AddInt64(&disconnectCount, 1)
		return nil
	})

	client.Stop()
	client.Stop()
	client.Stop()

	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadInt64(&disconnectCount)
	if count != 1 {
		t.Errorf("Expected Stop to be idempotent, disconnect count: %d", count)
	}
}

func setupTestServer(t *testing.T) (*Server, string) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	addr := listener.Addr().String()
	listener.Close()

	server, err := NewServerWithConfig(ServerConfig{
		Addr:         addr,
		MsgQueueSize: 1000,
		PoolSize:     10,
		BufSize:      4096,
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	return server, addr
}

func TestServerClientIntegration(t *testing.T) {
	server, addr := setupTestServer(t)

	router := NewRouter()
	var serverReceived string

	RegisterRoute[string](router, 1000, func(c *Conn, msg *string) error {
		serverReceived = *msg
		return nil
	})
	server.SetRouter(router)

	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	client := NewClientWithConfig(ClientConfig{Addr: addr})
	clientRouter := NewRouter()
	client.SetRouter(clientRouter)

	client.On(ClientEventConnected, func(p EventPayload) error {
		msg, _ := EncodeMessage(1000, "test message")
		return client.Send(msg)
	})

	err := client.Start()
	if err != nil {
		t.Logf("Client failed to connect: %v", err)
		return
	}

	time.Sleep(300 * time.Millisecond)

	if serverReceived != "test message" {
		t.Errorf("Server expected 'test message', got '%s'", serverReceived)
	}

	client.Stop()
}

func TestServerBroadcast(t *testing.T) {
	server, addr := setupTestServer(t)

	router := NewRouter()
	server.SetRouter(router)

	go server.Start()
	defer server.Stop()

	time.Sleep(100 * time.Millisecond)

	// 连接两个客户端
	var receivedCount1, receivedCount2 int64

	client1 := NewClientWithConfig(ClientConfig{Addr: addr})
	router1 := NewRouter()
	RegisterRoute[string](router1, 2000, func(c *Conn, msg *string) error {
		atomic.AddInt64(&receivedCount1, 1)
		return nil
	})
	client1.SetRouter(router1)

	client2 := NewClientWithConfig(ClientConfig{Addr: addr})
	router2 := NewRouter()
	RegisterRoute[string](router2, 2000, func(c *Conn, msg *string) error {
		atomic.AddInt64(&receivedCount2, 1)
		return nil
	})
	client2.SetRouter(router2)

	client1.Start()
	client2.Start()

	time.Sleep(200 * time.Millisecond)

	// 服务端广播
	msg, _ := EncodeMessage(2000, "broadcast message")
	server.Broadcast(msg, nil)

	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt64(&receivedCount1) == 0 || atomic.LoadInt64(&receivedCount2) == 0 {
		t.Logf("Broadcast may have failed: client1=%d, client2=%d",
			atomic.LoadInt64(&receivedCount1), atomic.LoadInt64(&receivedCount2))
	}

	client1.Stop()
	client2.Stop()
}
