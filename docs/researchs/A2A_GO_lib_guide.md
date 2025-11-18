基于提供的信息，我来为您详细介绍如何使用 `a2aproject/a2a-go` 库，特别是结合 JSON-RPC 和 gRPC 服务器的实现方式。

## A2A-Go 库使用指南

### 1. 项目概述
A2A-Go 是一个用于构建遵循 Agent2Agent (A2A) 协议的代理应用程序的 Go 库。支持构建 A2A 服务器和客户端，具有协议无关性，可轻松扩展不同的通信协议和数据库后端。

**要求**: Go 1.24.4 或更高版本
**安装**: `go get github.com/a2aproject/a2a-go`

### 2. 服务器端实现

#### 2.1 核心架构
A2A 服务器采用分层架构：
- **传输协议无关层**: 核心业务逻辑
- **传输协议层**: 具体的通信协议实现（gRPC、JSON-RPC等）
- **服务器层**: 实际的服务器实现

#### 2.2 JSON-RPC 服务器实现
```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/a2aproject/a2a-go/a2asrv"
    "github.com/a2aproject/a2a-go/a2asrv/a2ajsonrpc"
)

func main() {
    // 1. 创建自定义选项（可选）
    var options []a2asrv.RequestHandlerOption
    // 例如：设置超时、认证等
    options = append(options, a2asrv.WithRequestTimeout(30*time.Second))
    
    // 2. 创建 AgentExecutor 实现
    // 这是核心业务逻辑，需要实现 a2asrv.AgentExecutor 接口
    agentExecutor := &HelloWorldAgentExecutor{}
    
    // 3. 创建传输协议无关的请求处理器
    requestHandler := a2asrv.NewHandler(agentExecutor, options...)
    
    // 4. 将处理器包装到 JSON-RPC 传输实现中
    jsonrpcHandler := a2ajsonrpc.NewHandler(requestHandler)
    
    // 5. 设置 HTTP 服务器
    http.Handle("/a2a", jsonrpcHandler)
    
    // 6. 启动服务器
    log.Println("Starting JSON-RPC server on :8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

// HelloWorldAgentExecutor 实现 AgentExecutor 接口
type HelloWorldAgentExecutor struct{}

func (e *HelloWorldAgentExecutor) Execute(ctx context.Context, request *a2a.Message) (*a2a.Message, error) {
    // 业务逻辑处理
    response := a2a.NewMessage(
        a2a.MessageRoleAssistant,
        a2a.TextPart{Text: "Hello, World!"},
    )
    return response, nil
}
```

#### 2.3 gRPC 服务器实现
```go
package main

import (
    "context"
    "log"
    "net"
    
    "github.com/a2aproject/a2a-go/a2agrpc"
    "github.com/a2aproject/a2a-go/a2asrv"
    "google.golang.org/grpc"
)

func main() {
    // 1. 创建自定义选项
    var options []a2asrv.RequestHandlerOption
    // 例如：设置最大消息大小
    options = append(options, a2asrv.WithMaxMessageSize(1024*1024))
    
    // 2. 创建 AgentExecutor 实现
    agentExecutor := &HelloWorldAgentExecutor{}
    
    // 3. 创建传输协议无关的请求处理器
    requestHandler := a2asrv.NewHandler(agentExecutor, options...)
    
    // 4. 将处理器包装到 gRPC 传输实现中
    grpcHandler := a2agrpc.NewHandler(requestHandler)
    
    // 5. 创建 gRPC 服务器
    server := grpc.NewServer()
    
    // 6. 注册处理器
    grpcHandler.RegisterWith(server)
    
    // 7. 启动监听
    listener, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    
    log.Println("Starting gRPC server on :50051")
    if err := server.Serve(listener); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}
```

### 3. 客户端实现

#### 3.1 基本客户端使用
```go
package main

import (
    "context"
    "log"
    
    "github.com/a2aproject/a2a-go/a2a"
    "github.com/a2aproject/a2a-go/a2aclient"
    "github.com/a2aproject/a2a-go/agentcard"
)

func main() {
    ctx := context.Background()
    
    // 1. 解析 AgentCard 获取代理信息
    // AgentCard 包含了代理的元数据和连接信息
    card, err := agentcard.DefaultResolver.Resolve(ctx, "hello-world-agent")
    if err != nil {
        log.Fatalf("Failed to resolve agent card: %v", err)
    }
    
    // 2. 创建客户端选项（可选）
    var options a2aclient.FactoryOption
    // 例如：设置超时
    options = a2aclient.WithTimeout(10 * time.Second)
    
    // 3. 从 AgentCard 创建客户端
    client, err := a2aclient.NewFromCard(ctx, card, options)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }
    defer client.Close()
    
    // 4. 发送消息到服务器
    msg := a2a.NewMessage(
        a2a.MessageRoleUser,
        a2a.TextPart{Text: "Hello, A2A Server!"},
    )
    
    resp, err := client.SendMessage(ctx, &a2a.MessageSendParams{
        Message: msg,
    })
    if err != nil {
        log.Fatalf("Failed to send message: %v", err)
    }
    
    // 5. 处理响应
    log.Printf("Response received: %s", resp.Content)
}
```

#### 3.2 高级客户端配置
```go
func createAdvancedClient() {
    // 自定义解析器
    resolver := agentcard.NewResolver(agentcard.WithCache(true))
    
    // 自定义连接选项
    connOptions := a2aclient.WithConnectionParams(
        a2aclient.ConnectionParams{
            MaxRetries: 3,
            RetryDelay: 100 * time.Millisecond,
            TLSConfig:  &tls.Config{InsecureSkipVerify: true},
        },
    )
    
    // 自定义认证
    authOptions := a2aclient.WithAuthenticator(
        a2aclient.NewBearerTokenAuthenticator("your-token-here"),
    )
    
    // 创建客户端
    client, err := a2aclient.NewFromCard(ctx, card, connOptions, authOptions)
    // ...
}
```

### 4. AgentCard 管理

AgentCard 是 A2A 协议的核心概念，包含代理的元数据和连接信息：

```go
// 手动创建 AgentCard
card := &agentcard.AgentCard{
    ID:   "hello-world-agent",
    Name: "Hello World Agent",
    Description: "A simple agent that says hello",
    Protocols: []agentcard.Protocol{
        {
            Type: "jsonrpc",
            Endpoint: "http://localhost:8080/a2a",
        },
        {
            Type: "grpc",
            Endpoint: "localhost:50051",
        },
    },
    Capabilities: []string{"text-generation", "simple-responses"},
}

// 将 AgentCard 保存到文件
err := agentcard.SaveToFile(card, "agent-card.json")

// 从文件加载 AgentCard
loadedCard, err := agentcard.LoadFromFile("agent-card.json")
```

### 5. 最佳实践

#### 5.1 错误处理
```go
func handleA2AError(err error) {
    if err == nil {
        return
    }
    
    switch {
    case a2a.IsTimeoutError(err):
        log.Println("Request timed out")
    case a2a.IsAuthenticationError(err):
        log.Println("Authentication failed")
    case a2a.IsProtocolError(err):
        log.Printf("Protocol error: %v", err)
    default:
        log.Printf("Unknown error: %v", err)
    }
}
```

#### 5.2 性能优化
```go
// 服务器端优化
options := []a2asrv.RequestHandlerOption{
    a2asrv.WithConcurrencyLimit(100),          // 并发限制
    a2asrv.WithRequestTimeout(30*time.Second), // 请求超时
    a2asrv.WithResponseCaching(true),          // 响应缓存
    a2asrv.WithMetricsCollection(true),        // 指标收集
}

// 客户端优化
clientOptions := a2aclient.FactoryOption(
    a2aclient.WithConnectionPool(10),           // 连接池
    a2aclient.WithKeepAlive(30*time.Second),    // 保活
    a2aclient.WithRetryPolicy(3, 100*time.Millisecond), // 重试策略
)
```

#### 5.3 安全配置
```go
// 服务器端 TLS 配置
tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    MinVersion:   tls.VersionTLS12,
}

// gRPC 服务器 TLS
server := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))

// JSON-RPC 服务器 TLS
server := &http.Server{
    Addr:      ":8443",
    Handler:   jsonrpcHandler,
    TLSConfig: tlsConfig,
}
```

### 6. 调试和监控

```go
// 启用详细日志
a2asrv.EnableDebugLogging(true)
a2aclient.EnableDebugLogging(true)

// 自定义日志记录器
logger := logrus.New()
a2asrv.SetLogger(logger)
a2aclient.SetLogger(logger)

// 指标收集
metricsHandler := a2asrv.NewMetricsHandler(requestHandler)
// 注册到 Prometheus 等监控系统
```

### 7. 完整示例结构

```
project/
├── server/
│   ├── grpc/
│   │   └── main.go
│   ├── jsonrpc/
│   │   └── main.go
│   └── common/
│       └── executor.go
├── client/
│   └── main.go
├── agentcard/
│   └── hello-world.json
└── go.mod
```

### 8. 学习资源

- **完整文档**: [pkg.go.dev/a2asrv](https://pkg.go.dev/a2asrv) 和 [pkg.go.dev/a2aclient](https://pkg.go.dev/a2aclient)
- **示例仓库**: [a2a-samples](https://github.com/a2aproject/a2a-samples)
- **协议规范**: [A2A Protocol Specification](https://github.com/a2aproject/A2A)

这个指南提供了从基础到高级的完整使用说明，涵盖了 JSON-RPC 和 gRPC 两种传输协议的服务器实现，以及客户端的使用方法。实际开发时，建议从 helloworld 示例开始，逐步扩展到更复杂的应用场景。