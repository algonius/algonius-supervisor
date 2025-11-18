# 如何在一个服务端口上支持多个 Agent 的 A2A Protocol 端点

根据 A2A (Agent2Agent) Protocol 规范文档，实现单端口多 Agent 服务需要考虑路由机制、Agent Card 发现和端点映射。以下是几种可行的实现方案：

## 1. 路径前缀路由方案

### 架构设计
```
https://api.example.com/
├── /agent1/v1/          # Agent 1 的端点
│   ├── /message:send
│   ├── /tasks/{id}
│   └── /.well-known/agent-card.json
├── /agent2/v1/          # Agent 2 的端点
│   ├── /message:send
│   ├── /tasks/{id}
│   └── /.well-known/agent-card.json
└── /agents/             # Agent 注册表/目录
```

### 实现要点
- **Agent Card 配置**：每个 Agent 的 `url` 字段指向其专属路径
  ```json
  {
    "url": "https://api.example.com/agent1/v1",
    "preferredTransport": "HTTP+JSON",
    "additionalInterfaces": [
      {"url": "https://api.example.com/agent1/v1", "transport": "JSONRPC"},
      {"url": "https://api.example.com/agent1/grpc", "transport": "GRPC"}
    ]
  }
  ```

- **路由规则**：HTTP 服务器根据路径前缀路由到对应的 Agent 处理逻辑
- **规范符合性**：完全符合 A2A 规范，每个 Agent 有独立的端点 URL 和 Agent Card

### 优点
- 清晰的 URL 结构，易于理解和调试
- 每个 Agent 完全隔离，互不影响
- 符合 RESTful 设计原则

## 2. 虚拟主机/域名路由方案

### 架构设计
```
单个服务器端口 (443)
├── agent1.example.com/v1/    # Agent 1
├── agent2.example.com/v1/    # Agent 2  
└── agents.example.com/       # Agent 目录
```

### 实现要点
- **Agent Card 配置**：使用不同域名
  ```json
  {
    "url": "https://agent1.example.com/v1",
    "preferredTransport": "HTTP+JSON"
  }
  ```

- **服务器配置**：使用 SNI (Server Name Indication) 和虚拟主机配置
- **DNS 设置**：所有子域名指向同一 IP 地址

### 优点
- 完全符合 A2A 规范要求
- 每个 Agent 有独立的域名，易于品牌化
- 天然支持 TLS 证书隔离

## 3. Agent ID 头路由方案

### 架构设计
```
https://api.example.com/v1/
├── /message:send         # 通过 X-A2A-Agent-ID 头指定目标 Agent
├── /tasks/{id}           # 通过 X-A2A-Agent-ID 头指定目标 Agent
└── /.well-known/agent-cards/  # Agent Card 目录
    ├── agent1.json
    └── agent2.json
```

### 实现要点
- **自定义 HTTP 头**：使用 `X-A2A-Agent-ID` 或 `Agent-Id` 头指定目标 Agent
  ```http
  POST /v1/message:send
  Content-Type: application/json
  X-A2A-Agent-ID: agent1
  
  {"message": {...}}
  ```

- **Agent Card 发现**：
  - 公共注册表：`https://api.example.com/.well-known/agent-cards/agent1.json`
  - 客户端先获取目标 Agent 的 Card，再在请求中指定 Agent ID

- **Agent Card 配置**：
  ```json
  {
    "url": "https://api.example.com/v1",
    "preferredTransport": "HTTP+JSON",
    "additionalInterfaces": [
      {
        "url": "https://api.example.com/v1",
        "transport": "HTTP+JSON",
        "metadata": {"agentId": "agent1"}
      }
    ]
  }
  ```

### 优点
- 共享同一端点路径，简化 URL 结构
- 适合微服务架构，后端可动态路由
- 减少 URL 路径复杂度

## 4. 查询参数路由方案

### 架构设计
```
https://api.example.com/v1/
├── /message:send?agentId=agent1
├── /tasks/{id}?agentId=agent1
└── /agent-card?agentId=agent1
```

### 实现要点
- **查询参数**：使用 `agentId` 参数指定目标 Agent
  ```http
  POST /v1/message:send?agentId=agent1
  Content-Type: application/json
  
  {"message": {...}}
  ```

- **Agent Card 动态生成**：单个端点根据 agentId 参数返回不同的 Agent Card

- **规范调整**：需要扩展规范或使用扩展字段，因为标准 Agent Card 的 `url` 字段通常不包含查询参数

### 优点
- 实现简单，无需复杂的路由配置
- 适合原型开发和小型部署

## 5. 混合方案（推荐）

结合上述方案，创建灵活且符合规范的架构：

### 推荐架构
```
https://a2a.example.com/
├── /agents/                          # Agent 目录
│   ├── list.json                     # 所有 Agent 列表
│   ├── agent1/                       # Agent 1 专属空间
│   │   ├── .well-known/agent-card.json
│   │   └── v1/                       # Agent 1 的端点
│   │       ├── message:send
│   │       └── tasks/{id}
│   └── agent2/                       # Agent 2 专属空间
│       ├── .well-known/agent-card.json
│       └── v1/                       # Agent 2 的端点
│           ├── message:send
│           └── tasks/{id}
└── /.well-known/                     # 全局服务
    └── agent-cards/                  # 全局 Agent Card 注册表
        ├── agent1.json → /agents/agent1/.well-known/agent-card.json
        └── agent2.json → /agents/agent2/.well-known/agent-card.json
```

### 实现代码示例（Node.js/Express）

```javascript
const express = require('express');
const app = express();

// Agent 注册表
const agents = {
  'agent1': require('./agents/agent1'),
  'agent2': require('./agents/agent2')
};

// 全局 Agent Card 目录
app.get('/.well-known/agent-cards/:agentId.json', (req, res) => {
  const agentId = req.params.agentId;
  const agent = agents[agentId];
  
  if (!agent) {
    return res.status(404).json({ error: 'Agent not found' });
  }
  
  // 返回该 Agent 的 Card
  res.json(agent.getAgentCard());
});

// Agent 专属路由
Object.entries(agents).forEach(([agentId, agent]) => {
  const agentRouter = express.Router();
  
  // Agent Card 端点
  agentRouter.get('/.well-known/agent-card.json', (req, res) => {
    res.json(agent.getAgentCard());
  });
  
  // A2A 核心端点
  agentRouter.post('/v1/message:send', agent.handleMessageSend);
  agentRouter.get('/v1/tasks/:id', agent.handleGetTask);
  agentRouter.post('/v1/tasks/:id:cancel', agent.handleCancelTask);
  
  // 挂载到主应用
  app.use(`/agents/${agentId}`, agentRouter);
});

// Agent 目录
app.get('/agents/list.json', (req, res) => {
  const agentList = Object.keys(agents).map(agentId => ({
    id: agentId,
    cardUrl: `/.well-known/agent-cards/${agentId}.json`,
    endpoint: `/agents/${agentId}/v1`
  }));
  
  res.json({ agents: agentList });
});

app.listen(443, () => {
  console.log('A2A Multi-Agent Server running on port 443');
});
```

## 关键技术考虑

### 1. 身份验证和授权
- **多租户隔离**：确保每个 Agent 只能访问自己的数据和任务
- **统一认证**：可使用 JWT 令牌，包含 Agent ID 和权限信息
- **Agent Card 安全**：敏感 Agent Card 需要认证访问

### 2. 任务 ID 唯一性
- **全局唯一 ID**：使用 UUID 或包含 Agent 前缀的 ID（如 `agent1:task-123`）
- **上下文隔离**：`contextId` 也应包含 Agent 标识

### 3. 性能和扩展性
- **负载均衡**：单个端口可配合负载均衡器实现水平扩展
- **Agent 动态加载**：支持运行时注册/注销 Agent
- **资源隔离**：为不同 Agent 设置资源配额

## 规范符合性检查

| 方案 | 符合性 | 说明 |
|------|--------|------|
| 路径前缀路由 | ★★★★★ | 完全符合规范，每个 Agent 有独立 URL |
| 虚拟主机路由 | ★★★★★ | 完全符合规范，标准做法 |
| Agent ID 头路由 | ★★★☆☆ | 需要扩展规范，但可通过 `metadata` 字段支持 |
| 查询参数路由 | ★★☆☆☆ | 需要自定义实现，规范未明确支持 |
| 混合方案 | ★★★★★ | 最佳实践，兼顾规范和灵活性 |

## 最佳实践建议

1. **首选路径前缀路由**：为每个 Agent 分配独立的路径命名空间
2. **实现全局 Agent 目录**：提供 `/agents/list.json` 端点
3. **支持 well-known URI**：遵循 RFC 8615，使用 `/.well-known/agent-cards/` 目录
4. **统一认证层**：在路由层之前处理身份验证
5. **动态 Agent 注册**：支持运行时注册新 Agent
6. **监控和日志隔离**：确保不同 Agent 的监控数据隔离

通过上述方案，可以在单个服务端口上高效支持多个 A2A Agent，同时保持规范符合性和系统可扩展性。