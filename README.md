# Koduck

一个轻量级的 TCP 消息路由框架，专注于简单稳定的客户端/服务端通信。它提供：消息编码/解码、路由注册、连接管理、事件总线、心跳检测、定时任务调度与高性能工作池（基于 `ants`）。

> 模块名：`github.com/shandialamp/koduck`

## 特性
- 简洁的 JSON 消息协议（`method` + `params`）
- 基于方法号的路由注册与泛型参数反序列化
- 长连接管理：长度前缀包、线程安全发送队列
- 事件系统：服务端/客户端生命周期与错误事件
- 心跳机制与超时自动断开
- 简易定时任务调度器（毫秒级 tick）
- 高性能协程池处理消息（`github.com/panjf2000/ants/v2`）

## 安装
```bash
go get github.com/shandialamp/koduck
```

## 快速上手
项目已包含可运行示例：
- 服务端示例：`examples/simple/server/main.go`
- 客户端示例：`examples/simple/client/main.go`


## 消息协议与路由
- 消息结构：
  ```json
  { "method": 1000, "params": { ... } }
  ```
- 传输层：发送时采用长度前缀（`uint32`） + JSON 字节；接收端按长度完整读取。
- 路由注册：
  ```go
  // 将方法号与处理函数绑定，并自动反序列化 params
  koduck.RegisterRoute[T](router, method, func(c *koduck.Conn, params *T) error { ... })
  ```
- 手动编码/解码：
  ```go
  msg, _ := koduck.EncodeMessage(1234, payload)
  raw, _ := msg.Encode()
  ```

## 事件系统
可订阅以下事件：
- 服务端：`ServerEventStarted`、`ServerEventClientConnected`、`ServerEventClientDisconnected`、`ServerEventError`
- 客户端：`ClientEventConnected`、`ClientEventDisconnected`、`ClientEventError`

示例：
```go
server.On(koduck.ServerEventClientConnected, func(p koduck.EventPayload) error {
    payload := p.(*koduck.ServerEventClientConnectedPayload)
    fmt.Println("新连接：", payload.Conn.String())
    return nil
})
```

## 心跳与调度
- 客户端每隔 5s 发送心跳（方法号 `ClientHeartbeat`）
- 服务端每隔 30s 检查连接心跳，超时自动关闭
- 自带简易 `Scheduler`，支持 `Every(interval, handler)` 周期任务

## 配置
默认配置见 `config.go`：
```go
// 服务端默认
koduck.DefaultServerConfig() // Addr=:10001, MsgQueueSize=10000, PoolSize=5000, BufSize=4096

// 客户端默认（占位，需自行设置可用地址）
koduck.DefaultClientLog()   // Addr=ser:10001
```
如需自定义：
```go
server, _ := koduck.NewServerWithConfig(koduck.ServerConfig{
  Addr: ":10001",
  MsgQueueSize: 10000,
  PoolSize: 2000,
  BufSize: 8192,
})

client := koduck.NewClientWithConfig(koduck.ClientConfig{ Addr: "localhost:10001" })
```

## 开发建议
- 使用路由方法号进行语义分组（如 1xxx=用户相关，2xxx=通知）
- `Conn.Send` 为非阻塞队列发送，避免直接在业务中进行重 IO 操作
- 处理函数中注意错误上报（事件总线会统一投递错误事件）
- 长任务可提交到工作池或自行开协程，保持处理函数快捷

## 依赖
- `github.com/panjf2000/ants/v2` 协程池

## 目录结构
```
client.go       // 客户端实现
server.go       // 服务端实现
router.go       // 路由注册与分发
message.go      // 消息结构与编解码
conn.go         // 连接封装与发送队列
event.go        // 事件系统
scheduler.go    // 简易调度器
config.go       // 默认配置
workpool.go     // 备用工作池实现
examples/       // 客户端/服务端示例
```