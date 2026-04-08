# pkg/sse 包设计文档

## 概述

`sse` 包提供 Server-Sent Events (SSE) 服务端实现，支持流式数据推送。封装了 SSE 协议标准格式，提供高性能、易扩展的流式响应能力。

---

## 目录结构

```shell
pkg/sse/
├── sse.go      # 基础配置与工具函数
├── event.go    # Event 对象设计与对象池
└── writer.go   # 流式写入器核心实现
```

---

## 模块设计

### 1. sse.go - 基础配置

提供 SSE 协议相关的常量定义和辅助函数。

```go
const (
    MIMEType = "text/event-stream"
)

// 设置 SSE 响应头
func SetupSSEHeaders(w http.ResponseWriter)

// 判断请求是否接受 SSE
func IsSSEAcceptable(r *http.Request) bool
```

**HTTP Headers 说明**：

| Header | 值 | 作用 |
| -------- | ----- | ------ |
| Content-Type | text/event-stream | SSE 协议标识 |
| Cache-Control | no-cache | 禁用缓存 |
| Connection | keep-alive | 保持连接 |
| Access-Control-Allow-Origin | * | 允许跨域 |

---

### 2. event.go - Event 对象

#### 核心结构

```go
type Event struct {
    ID    string        // 事件 ID
    Type  string        // 事件类型
    Data  []byte        // 事件数据
    Retry time.Duration // 重连间隔

    bitset uint8        // 内部标记位
}
```

#### bitset 标记设计

使用位操作标记字段是否被设置，避免写入空字段：

```go
const (
    isSetID    uint8 = 1 << iota  // 0001
    isSetType                      // 0010
    isSetData                      // 0100
    isSetRetry                     // 1000
)
```

**优势**：

- 减少内存分配
- 快速判断字段是否存在
- 节省序列化开销

#### sync.Pool 对象池

```go
var eventPool = sync.Pool{
    New: func() interface{} {
        return &Event{}
    },
}
```

**优势**：

- 减少 GC 压力
- 高频创建/释放 Event 时性能提升
- 自动复用已分配对象

#### 核心方法

| 方法 | 说明 |
| ------ | ------ |
| `NewEvent()` | 从对象池获取 Event |
| `Release()` | 归还对象池 |
| `Clone()` | 深拷贝 Event |
| `SetID/SetEvent/SetData/SetRetry` | 设置字段（同时标记 bitset） |
| `AppendData/AppendDataString` | 追加数据 |
| `IsSetXxx()` | 检查字段是否已设置 |
| `String()` | 序列化为 SSE 格式 |

#### SSE 格式输出

`String()` 方法将 Event 序列化为标准 SSE 格式：

```yaml
id: {event_id}
event: {event_type}
data: {data_line1}
data: {data_line2}

```

**特点**：

- 多行数据自动拆分为多个 `data:` 行
- 支持 `\n` 和 `\r\n` 换行符处理
- 末尾自动追加双换行分隔符

---

### 3. writer.go - 流式写入器

#### 3.1 核心结构

```go
type Writer struct {
    id       string            // 当前会话 ID
    w        http.ResponseWriter // HTTP 响应写入器
    flusher  http.Flusher       // 支持流式刷新的接口
    closed   bool               // 是否已关闭
    Done     func(http.ResponseWriter) error  // 完成时回调
}
```

#### 3.2 工厂方法

```go
func NewWriter(id string, w http.ResponseWriter) *Writer
```

**功能**：

- 检查 `http.ResponseWriter` 是否实现 `http.Flusher` 接口
- 设置 SSE 相关 HTTP 头
- 初始化 Writer 实例

#### 写入方法

| 方法 | 说明 |
| ------ | ------ |
| `WriteEvent(event)` | 写入单个 Event |
| `WriteEventSimple(id, eventType, data)` | 快捷写入 |
| `WriteEventString(id, eventType, data)` | 字符串数据版本 |
| `WriteEventJSON(id, eventType, data)` | JSON 数据版本 |
| `WriteData(data)` | 快速写入数据（使用默认 ID） |
| `WriteDataString(data)` | 字符串数据版本 |
| `WriteJSONData(data)` | JSON 数据写入 |
| `WriteComment(comment)` | 写入注释（用于心跳/保活） |
| `WriteKeepAlive()` | 发送心跳保活 |
| `WriteDone()` | 发送结束信号 `[DONE]` |

#### 核心流式方法

```go
func (w *Writer) Stream(
    ctx context.Context,
    stream *schema.StreamReader[*schema.Message],
    fn func(output *schema.Message, index int) any,
) error
```

**参数说明**：

| 参数 | 类型 | 说明 |
| ------ | ------ | ------ |
| `ctx` | context.Context | 上下文，支持取消/超时 |
| `stream` | *schema.StreamReader | Eino 流式消息读取器 |
| `fn` | transform function | 消息转换函数，可自定义输出格式 |

**执行流程**：

```shell
┌─────────────────────────┐
│  Stream(ctx, stream, fn) │
└───────────┬─────────────┘
            │
            ▼
    ┌───────────────┐
    │ ctx 已取消?    │──Yes──▶ 返回 ctx.Err()
    └───────┬───────┘
            │ No
            ▼
    ┌───────────────┐
    │ writer 已关闭?│──Yes──▶ 返回 error
    └───────┬───────┘
            │ No
            ▼
    ┌───────────────┐
    │  stream.Recv()│
    └───────┬───────┘
            │
      ┌─────┴─────┐
      │ err ==    │──Yes──▶ WriteDone() → 返回 nil
      │ io.EOF    │
      └─────┬─────┘
            │ No
            ▼
    ┌───────────────┐
    │ chunk 为空?   │──Yes──▶ continue
    └───────┬───────┘
            │ No
            ▼
    ┌───────────────┐
    │ fn(chunk)     │ 转换消息格式
    └───────┬───────┘
            │
            ▼
    ┌───────────────┐
    │ Marshal → JSON│
    └───────┬───────┘
            │
            ▼
    ┌───────────────┐
    │ WriteData(b)  │ 写入 SSE 流
    └───────┬───────┘
            │
            └──▶ loop
```

**转换函数示例**：

```go
// 使用默认格式
err := writer.Stream(ctx, stream, nil)

// 自定义转换（OpenAI 格式）
err := writer.Stream(ctx, stream, func(output *schema.Message, index int) any {
    return model.OutStreamMessageEinoToOpenai(output, index)
})
```

---

## 使用示例

### 基础用法

```go
func chatHandler(w http.ResponseWriter, r *http.Request) {
    // 创建 SSE Writer
    writer := sse.NewWriter(sessionID, w)
    if writer == nil {
        http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
        return
    }
    defer writer.Close()

    // 模拟流式数据
    stream := model.Stream(...)
    
    // 启动流
    err := writer.Stream(r.Context(), stream, nil)
    if err != nil {
        log.Printf("Stream error: %v", err)
    }
}
```

### 完整聊天示例

```go
func chatHandler(w http.ResponseWriter, r *http.Request) {
    var req ChatRequest
    json.NewDecoder(r.Body).Decode(&req)

    // 调用模型流式接口
    ctx := r.Context()
    input := []*schema.Message{
        schema.UserMessage(req.Message),
    }

    stream, err := cm.Stream(ctx, input)
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }
    defer stream.Close()

    // 创建 SSE Writer
    writer := sse.NewWriter(sessionID, w)
    defer writer.Close()

    // 自定义完成回调
    writer.SetDone(func(w http.ResponseWriter) error {
        // 可在此处执行清理或发送额外数据
        return nil
    })

    // 流式推送（转换为 OpenAI 兼容格式）
    err = writer.Stream(ctx, stream, func(output *schema.Message, index int) any {
        return model.OutStreamMessageEinoToOpenai(output, index)
    })
}
```

### 前端接收

```javascript
const response = await fetch('/api/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ message: 'Hello' })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    const chunk = decoder.decode(value);
    const lines = chunk.split('\n');

    for (const line of lines) {
        if (line.startsWith('data: ')) {
            const data = line.slice(6);
            if (data === '[DONE]') break;

            const parsed = JSON.parse(data);
            const content = parsed.choices?.[0]?.delta?.content;
            if (content) {
                console.log('Received:', content);
            }
        }
    }
}
```

---

## 设计亮点

### 1. 高性能

- **sync.Pool 对象池**：Event 对象复用，减少 GC
- **bitset 标记位**：轻量级字段状态追踪
- **字节切片拼接**：避免字符串频繁分配

### 2. 零外部依赖

仅依赖标准库 `net/http`，易于集成和移植。

### 3. 灵活的转换机制

`Stream()` 方法支持自定义转换函数，可适配各种输出格式（OpenAI、Claude、自定义格式）。

### 4. 完善的错误处理

- 上下文取消/超时感知
- Writer 关闭状态检测
- 流结束（EOF）自动处理

### 5. 保活机制

支持心跳注释（Keep-Alive），防止连接被代理/负载均衡器断开。

---

## 扩展建议

### 1. 支持更多事件类型

可扩展 Event 支持：

```go
type Event struct {
    // 现有字段...
    Comment string  // 注释行
}
```

### 2. 批量写入优化

```go
func (w *Writer) WriteEvents(events []*Event) error {
    for _, e := range events {
        if err := w.WriteEvent(e); err != nil {
            return err
        }
    }
    return nil
}
```

### 3. 限流/背压控制

在高并发场景下，可添加：

```go
type Writer struct {
    // 现有字段...
    rateLimiter *time.Ticker  // 限流器
    bufferSize  int           // 缓冲区大小
}
```

---

## 依赖关系

```
pkg/sse
└── net/http (标准库)
    └── github.com/bytedance/sonic (JSON 序列化，仅 writer.go)
    └── github.com/cloudwego/eino/schema (流式消息类型，仅 writer.go)
```

---

## 适用场景

| 场景 | 适用性 |
|------|--------|
| Web 聊天应用 | ✅ 最佳实践 |
| AI 流式响应 | ✅ 完美支持 |
| 实时通知推送 | ✅ 适用 |
| 进度条/任务状态 | ✅ 适用 |
| 双向通信（需要 WebSocket） | ❌ 不适用 |
| 高频低延迟交易 | ⚠️ 需优化 |
