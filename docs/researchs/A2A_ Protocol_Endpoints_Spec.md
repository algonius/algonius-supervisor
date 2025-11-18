# A2A Protocol Endpoints 规约文档

## 1. 概述

A2A (Agent2Agent) Protocol 定义了一组标准化的端点，用于 AI 代理系统之间的互操作与通信。该协议支持三种主要传输协议：JSON-RPC 2.0、gRPC 和 HTTP+JSON/REST。所有端点必须通过 HTTPS 传输，确保安全通信。

## 2. 传输协议映射

### 2.1 端点映射总览

| 功能描述 | JSON-RPC 方法 | gRPC 方法 | REST 端点 | 说明 |
|---------|--------------|----------|----------|------|
| 发送消息 | `message/send` | `SendMessage` | `POST /v1/message:send` | 向代理发送消息，创建或继续任务 |
| 流式消息 | `message/stream` | `SendStreamingMessage` | `POST /v1/message:stream` | 发送消息并订阅实时更新 |
| 获取任务 | `tasks/get` | `GetTask` | `GET /v1/tasks/{id}` | 获取任务状态和详情 |
| 列出任务 | `tasks/list` | `ListTask` | `GET /v1/tasks` | 列出任务（支持过滤和分页） |
| 取消任务 | `tasks/cancel` | `CancelTask` | `POST /v1/tasks/{id}:cancel` | 取消进行中的任务 |
| 重新订阅 | `tasks/resubscribe` | `TaskSubscription` | `POST /v1/tasks/{id}:subscribe` | 重新连接到任务流 |
| 设置推送通知 | `tasks/pushNotificationConfig/set` | `CreateTaskPushNotification` | `POST /v1/tasks/{id}/pushNotificationConfigs` | 配置任务的推送通知 |
| 获取推送配置 | `tasks/pushNotificationConfig/get` | `GetTaskPushNotification` | `GET /v1/tasks/{id}/pushNotificationConfigs/{configId}` | 获取推送通知配置 |
| 列出推送配置 | `tasks/pushNotificationConfig/list` | `ListTaskPushNotification` | `GET /v1/tasks/{id}/pushNotificationConfigs` | 列出任务的推送配置 |
| 删除推送配置 | `tasks/pushNotificationConfig/delete` | `DeleteTaskPushNotification` | `DELETE /v1/tasks/{id}/pushNotificationConfigs/{configId}` | 删除推送通知配置 |
| 获取扩展卡片 | `agent/getAuthenticatedExtendedCard` | `GetAgentCard` | `GET /v1/card` | 获取认证后的详细代理卡片 |

## 3. 详细端点规范

### 3.1 消息相关端点

#### 3.1.1 `message/send`

**功能**：向代理发送消息，创建新任务或继续现有任务。

**协议映射**：
- JSON-RPC: `message/send`
- gRPC: `SendMessage`
- REST: `POST /v1/message:send`

**请求参数** (`MessageSendParams`):
```json
{
  "message": Message,
  "configuration": {
    "acceptedOutputModes": string[],
    "historyLength": number,
    "pushNotificationConfig": PushNotificationConfig,
    "blocking": boolean
  },
  "metadata": Record<string, any>
}
```

**响应**：
- 成功: `Task` 或 `Message` 对象
- 错误: 标准 JSON-RPC 错误

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

#### 3.1.2 `message/stream`

**功能**：发送消息并订阅实时任务更新（SSE 流）。

**协议映射**：
- JSON-RPC: `message/stream`
- gRPC: `SendStreamingMessage`
- REST: `POST /v1/message:stream`

**请求参数**：同 `message/send`

**响应**：
- 成功: SSE 流（`text/event-stream`），每个事件包含 `SendStreamingMessageResponse`
- 错误: 标准 JSON-RPC 错误

**内容类型**：
- 请求: `application/json`
- 响应: `text/event-stream`

### 3.2 任务管理端点

#### 3.2.1 `tasks/get`

**功能**：获取特定任务的当前状态，包括状态、产物和历史记录。

**协议映射**：
- JSON-RPC: `tasks/get`
- gRPC: `GetTask`
- REST: `GET /v1/tasks/{id}`

**请求参数** (`TaskQueryParams`):
```json
{
  "id": string,
  "historyLength": number
}
```

**响应**：
- 成功: `Task` 对象
- 错误: 标准 JSON-RPC 错误

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

#### 3.2.2 `tasks/list`

**功能**：列出任务，支持过滤、分页和排序。

**协议映射**：
- JSON-RPC: `tasks/list`
- gRPC: `ListTask`
- REST: `GET /v1/tasks`

**请求参数** (`ListTasksParams`):
```json
{
  "contextId": string,
  "status": TaskState,
  "pageSize": number (1-100, 默认 50),
  "pageToken": string,
  "historyLength": number (默认 0),
  "lastUpdatedAfter": number (毫秒时间戳),
  "includeArtifacts": boolean (默认 false),
  "metadata": Record<string, any>
}
```

**响应** (`ListTasksResult`):
```json
{
  "tasks": Task[],
  "totalSize": number,
  "pageSize": number,
  "nextPageToken": string (无更多结果时为空字符串)
}
```

**错误情况**：
- `pageSize` 不在 1-100 范围内
- `pageToken` 格式无效或过期
- `historyLength` 为负数
- `status` 值无效
- `lastUpdatedAfter` 时间戳格式无效或为未来时间

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

#### 3.2.3 `tasks/cancel`

**功能**：请求取消进行中的任务。

**协议映射**：
- JSON-RPC: `tasks/cancel`
- gRPC: `CancelTask`
- REST: `POST /v1/tasks/{id}:cancel`

**请求参数** (`TaskIdParams`):
```json
{
  "id": string
}
```

**响应**：
- 成功: 更新后的 `Task` 对象
- 错误: 标准 JSON-RPC 错误（如 `TaskNotCancelableError`）

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

#### 3.2.4 `tasks/resubscribe`

**功能**：重新连接到任务的 SSE 流（在连接中断后恢复）。

**协议映射**：
- JSON-RPC: `tasks/resubscribe`
- gRPC: `TaskSubscription`
- REST: `POST /v1/tasks/{id}:subscribe`

**请求参数** (`TaskIdParams`):
```json
{
  "id": string
}
```

**响应**：
- 成功: SSE 流，包含 `SendStreamingMessageResponse` 事件
- 错误: 标准 JSON-RPC 错误

**内容类型**：
- 请求: `application/json`
- 响应: `text/event-stream`

### 3.3 推送通知端点

> **注意**：这些端点需要代理在其 AgentCard 中声明 `capabilities.pushNotifications: true`

#### 3.3.1 `tasks/pushNotificationConfig/set`

**功能**：设置任务的推送通知配置。

**协议映射**：
- JSON-RPC: `tasks/pushNotificationConfig/set`
- gRPC: `CreateTaskPushNotification`
- REST: `POST /v1/tasks/{id}/pushNotificationConfigs`

**请求参数** (`TaskPushNotificationConfig`):
```json
{
  "taskId": string,
  "pushNotificationConfig": {
    "id": string,
    "url": string,
    "token": string,
    "authentication": {
      "schemes": string[],
      "credentials": string
    }
  }
}
```

**响应**：
- 成功: `TaskPushNotificationConfig` 对象
- 错误: 标准 JSON-RPC 错误（如 `PushNotificationNotSupportedError`）

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

#### 3.3.2 `tasks/pushNotificationConfig/get`

**功能**：获取特定任务的推送通知配置。

**协议映射**：
- JSON-RPC: `tasks/pushNotificationConfig/get`
- gRPC: `GetTaskPushNotification`
- REST: `GET /v1/tasks/{id}/pushNotificationConfigs/{configId}`

**请求参数** (`GetTaskPushNotificationConfigParams`):
```json
{
  "taskId": string,
  "configId": string
}
```

**响应**：
- 成功: `TaskPushNotificationConfig` 对象
- 错误: 标准 JSON-RPC 错误

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

#### 3.3.3 `tasks/pushNotificationConfig/list`

**功能**：列出特定任务的所有推送通知配置。

**协议映射**：
- JSON-RPC: `tasks/pushNotificationConfig/list`
- gRPC: `ListTaskPushNotification`
- REST: `GET /v1/tasks/{id}/pushNotificationConfigs`

**请求参数** (`ListTaskPushNotificationConfigParams`):
```json
{
  "taskId": string
}
```

**响应**：
- 成功: `TaskPushNotificationConfig[]` 数组
- 错误: 标准 JSON-RPC 错误

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

#### 3.3.4 `tasks/pushNotificationConfig/delete`

**功能**：删除特定任务的推送通知配置。

**协议映射**：
- JSON-RPC: `tasks/pushNotificationConfig/delete`
- gRPC: `DeleteTaskPushNotification`
- REST: `DELETE /v1/tasks/{id}/pushNotificationConfigs/{configId}`

**请求参数** (`DeleteTaskPushNotificationConfigParams`):
```json
{
  "taskId": string,
  "configId": string
}
```

**响应**：
- 成功: `null`
- 错误: 标准 JSON-RPC 错误

**内容类型**：
- 请求: `application/json`
- 响应: `application/json`

### 3.4 代理发现端点

#### 3.4.1 `agent/getAuthenticatedExtendedCard`

**功能**：获取认证后的详细代理卡片（包含可能的额外信息）。

**协议映射**：
- JSON-RPC: `agent/getAuthenticatedExtendedCard`
- gRPC: `GetAgentCard`
- REST: `GET /v1/card`

**请求参数**：无

**响应**：
- 成功: 完整的 `AgentCard` 对象
- 错误: 标准 HTTP 错误（如 401 Unauthorized）

**要求**：
- 代理必须在公共 AgentCard 中设置 `supportsAuthenticatedExtendedCard: true`
- 客户端必须使用声明的认证方案进行认证

**内容类型**：
- 请求: 无特殊要求（标准认证头）
- 响应: `application/json`

## 4. 通用规范

### 4.1 身份验证

- 所有端点必须通过 HTTPS 传输
- 身份验证通过 HTTP 头进行（如 `Authorization: Bearer <token>`）
- 代理必须在 `AgentCard` 中声明所需的认证方案
- 客户端必须在每次请求中提供有效的认证凭证

### 4.2 错误处理

#### 4.2.1 标准 JSON-RPC 错误

| 代码 | 含义 | 说明 |
|------|------|------|
| -32700 | Parse error | JSON 格式无效 |
| -32600 | Invalid Request | 无效的 JSON-RPC 请求 |
| -32601 | Method not found | 方法不存在或不支持 |
| -32602 | Invalid params | 参数无效 |
| -32603 | Internal error | 服务器内部错误 |

#### 4.2.2 A2A 特定错误

| 代码 | 错误名称 | 说明 |
|------|----------|------|
| -32001 | TaskNotFoundError | 任务不存在或已过期 |
| -32002 | TaskNotCancelableError | 任务无法取消（已处于终止状态） |
| -32003 | PushNotificationNotSupportedError | 代理不支持推送通知 |
| -32004 | UnsupportedOperationError | 操作不被支持 |
| -32005 | ContentTypeNotSupportedError | 不支持的内容类型 |

### 4.3 传输协议要求

- **JSON-RPC 2.0**：
  - 使用 `application/json` 内容类型
  - 方法名称格式：`{category}/{action}`
  - 严格遵循 JSON-RPC 2.0 规范

- **gRPC**：
  - 使用 Protocol Buffers 版本 3
  - 支持 TLS 加密
  - 实现 `A2AService` 服务

- **HTTP+JSON/REST**：
  - 遵循 RESTful 设计原则
  - 使用标准 HTTP 状态码
  - 资源 URL 格式：`/v1/{resource}[/{id}][:{action}]`

### 4.4 端点兼容性要求

- 代理必须至少支持一种传输协议
- 支持多种传输协议的代理必须保证功能等价性
- 所有传输协议必须使用相同的数据结构和错误代码
- 代理必须在其 AgentCard 中准确声明支持的传输协议

## 5. 典型工作流程

### 5.1 基本任务执行

1. 获取代理卡片：从 well-known URI 或注册表获取 `AgentCard`
2. 发送消息：使用 `message/send` 创建任务
3. 轮询状态：使用 `tasks/get` 检查任务状态
4. 获取结果：任务完成后获取产物

### 5.2 流式任务执行

1. 获取代理卡片
2. 发送流式消息：使用 `message/stream` 启动任务并订阅更新
3. 处理流事件：接收 `TaskStatusUpdateEvent` 和 `TaskArtifactUpdateEvent`
4. 处理完成：当收到 `final: true` 事件时，流结束

### 5.3 异步任务执行（推送通知）

1. 获取代理卡片
2. 配置推送通知：使用 `tasks/pushNotificationConfig/set`
3. 发送消息：使用 `message/send`（包含推送配置）
4. 释放连接：客户端可断开连接
5. 接收通知：代理完成任务后 POST 到配置的 webhook
6. 获取结果：收到通知后使用 `tasks/get` 获取最终结果

## 6. 安全考虑

- 所有生产部署必须使用 HTTPS
- 代理必须验证客户端的 TLS 证书
- 推送通知 webhook 必须进行认证和授权
- 客户端必须验证推送通知的来源
- 实现输入验证以防止注入攻击
- 实施速率限制和资源控制
- 遵循数据隐私法规，最小化敏感数据传输

> 文档基于 A2A Protocol Specification v0.3.0 (2025年4月)