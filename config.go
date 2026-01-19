package koduck

type ServerConfig struct {
	Addr string

	MsgQueueSize int
	PoolSize     int
	BufSize      int

	Log LogConfig
}

type ClientConfig struct {
	Addr string
	Log  LogConfig
}

type LogConfig struct {
	AccessFile string
	ErrorFile  string
	AppFile    string
	MaxSize    int // MB
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:         ":10101",
		MsgQueueSize: 10000,
		PoolSize:     5000,
		BufSize:      4096,
		Log: LogConfig{
			AccessFile: "./logs/access.log",
			ErrorFile:  "./logs/error.log",
			AppFile:    "./logs/framework.log",
			MaxSize:    50, // MB
			MaxBackups: 30,
			MaxAge:     7, // days
			Compress:   true,
		},
	}
}

func DefaultClientLog() ClientConfig {
	return ClientConfig{
		Addr: "127.0.0.1:10101",
		Log: LogConfig{
			AccessFile: "./logs/access.log",
			ErrorFile:  "./logs/error.log",
			AppFile:    "./logs/framework.log",
			MaxSize:    50, // MB
			MaxBackups: 30,
			MaxAge:     7, // days
			Compress:   true,
		},
	}
}
