package koduck

type ServerConfig struct {
	Addr string

	MsgQueueSize int
	PoolSize     int
	BufSize      int
}

type ClientConfig struct {
	Addr string
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:         ":10001",
		MsgQueueSize: 10000,
		PoolSize:     5000,
		BufSize:      4096,
	}
}

func DefaultClientLog() ClientConfig {
	return ClientConfig{
		Addr: "127.0.0.1:10001",
	}
}
