# Agent 模块设计文档

Agent 模块是 AGGO 框架的核心组件，提供基于 CloudWeGo Eino ADK (Agent Development Kit) 的智能代理构建能力。

## 概述

Agent 模块采用 **Builder 模式** 封装 Eino ADK，简化 Agent 创建流程，同时通过 **Middleware 机制** 实现可扩展的功能增强。核心设计理念：

- **配置透传**：不重新实现执行逻辑，仅做配置组装和透传
- **链式调用**：Builder 模式提供流畅的 API 体验
- **中间件扩展**：通过 Handler 链实现横切关注点（记忆、指令格式化等）

## 架构设计

```html
┌─────────────────────────────────────────────────────────────────┐
│                        AgentBuilder                              │
│                                                                  │
│  WithName() → WithInstruction() → WithTools() → WithMemory()   │
│       ↓                                                          │
│  Build() ──────────────────────────────────────────────────────►│
│       │                                                          │
│       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                    Middleware Chain                          ││
│  │                                                              ││
│  │  [MemoryMiddleware] → [CustomMiddleware] → [instructionFmt] ││
│  │                                                              ││
│  └─────────────────────────────────────────────────────────────┘│
│       │                                                          │
│       ▼                                                          │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                     adk.Agent                                ││
│  │            (CloudWeGo Eino ADK ReAct Agent)                 ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## 核心组件

### 1. AgentBuilder

Agent 构建器，提供链式 API 配置 Agent 参数。

**文件**: `builder.go`

```go
agent, err := agent.NewAgentBuilder(chatModel).
    WithName("智能助手").
    WithDescription("一个友好的 AI 助手").
    WithInstruction("你是 {name}，一个专业的助手").
    WithTools(tool1, tool2).
    WithMemory(memoryProvider).
    WithMaxStep(10).
    WithSubAgents(agent.SubAgentModeSupervisor, subAgent1, subAgent2).
    Build(ctx)
```

**配置选项**:

| 方法 | 说明 | 默认值 |
| ------ | ------ | -------- |
| `WithName(name string)` | Agent 名称 | `"adk agent"` |
| `WithDescription(desc string)` | Agent 描述 | `"adk agent"` |
| `WithInstruction(instruction string)` | 系统提示词 | 无 |
| `WithTools(tools ...tool.BaseTool)` | 工具列表 | 空 |
| `WithMemoryMiddleware(mm *memory.MemoryMiddleware)` | 记忆中间件 | 无 |
| `WithMemory(provider memory.MemoryProvider)` | 记忆提供者（自动创建中间件） | 无 |
| `WithMiddlewares(mw ...adk.ChatModelAgentMiddleware)` | 自定义中间件 | 无 |
| `WithMaxStep(maxStep int)` | 最大迭代次数 | ADK 默认值 |
| `WithSubAgents(mode string, agents ...adk.Agent)` | 子 Agent 配置 | 无 |

### 2. instructionFormatter

指令格式化中间件，负责：

- 为框架注入的段落添加 XML 标签，提升 LLM 理解能力
- 运行时检测 Skill 工具是否有可用技能，无技能时移除相关段落

**文件**: `instruction_formatter.go`

**处理流程**:

```html
原始 Instruction
       │
       ▼
┌──────────────────────────────────────────────────────────────┐
│  检测 # Skills System 段落                                    │
│  ├─ 存在 → 检查 skill 工具是否有实际技能                       │
│  │   ├─ 有技能 → 保留段落，添加 <skills_system> 标签          │
│  │   └─ 无技能 → 移除段落 + 移除 skill 工具                   │
│  └─ 不存在 → 跳过                                             │
└──────────────────────────────────────────────────────────────┘
       │
       ▼
┌──────────────────────────────────────────────────────────────┐
│  检测 Available other agents 段落                             │
│  └─ 存在 → 添加 <available_agents> 标签                       │
└──────────────────────────────────────────────────────────────┘
       │
       ▼
格式化后的 Instruction
```

**输出示例**:

```html
你是一个智能助手。

## 工作原则
1. 回复简洁

<available_agents>
Available other agents:
- Agent name: cron
  Agent description: 定时任务助手
  Decision rule: ...
</available_agents>

<skills_system>
# Skills System

**How to Use Skills**
...
</skills_system>
```

### 3. 常量定义

**文件**: `types.go`

```go
const (
    SubAgentModeDefault    = ""           // 默认模式
    SubAgentModeSupervisor = "supervisor" // 监督者模式
)
```

**子 Agent 模式说明**:

- `SubAgentModeDefault`: 子 Agent 可自由转移
- `SubAgentModeSupervisor`: 子 Agent 只能转移回父 Agent，形成层级控制

## 执行流程

### 单 Agent 执行流程

```html
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   用户输入    │────►│  adk.Runner  │────►│  Middleware  │
└──────────────┘     └──────────────┘     │   Chain      │
                                          └──────┬───────┘
                                                 │
                     ┌───────────────────────────┘
                     ▼
          ┌─────────────────────┐
          │ instructionFormatter│
          │ • 格式化指令         │
          │ • 清理无效技能       │
          └──────────┬──────────┘
                     │
                     ▼
          ┌─────────────────────┐
          │    MemoryMiddleware │
          │ • 检索历史记忆       │
          │ • 注入上下文         │
          └──────────┬──────────┘
                     │
                     ▼
          ┌─────────────────────┐
          │     Chat Model      │
          │   (ReAct 循环)      │
          │ • 推理 (Reasoning)  │
          │ • 执行工具 (Acting) │
          └──────────┬──────────┘
                     │
                     ▼
          ┌─────────────────────┐
          │    AfterModel       │
          │ • 存储对话记忆       │
          └─────────────────────┘
```

### 多 Agent 协作流程

```html
                    ┌─────────────────┐
                    │   Router Agent  │
                    │   (主路由器)     │
                    └────────┬────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
          ▼                  ▼                  ▼
   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
   │  Math Agent │   │Weather Agent│   │  Time Agent │
   │  数学助手   │   │  天气助手   │   │  时间助手   │
   └─────────────┘   └─────────────┘   └─────────────┘
```

**Supervisor 模式**:

```go
// 子 Agent 只能返回父 Agent，不能互相转移
agent, err := agent.NewAgentBuilder(chatModel).
    WithInstruction("你是路由助手").
    WithSubAgents(agent.SubAgentModeSupervisor, subAgent1, subAgent2).
    Build(ctx)
```

## 使用示例

### 基础用法

```go
package main

import (
    "context"
    "log"

    "github.com/CoolBanHub/aggo/agent"
    "github.com/CoolBanHub/aggo/model"
    "github.com/cloudwego/eino/adk"
    "github.com/cloudwego/eino/schema"
)

func main() {
    ctx := context.Background()

    // 1. 创建聊天模型
    chatModel, err := model.NewChatModel(
        model.WithBaseUrl("https://api.example.com"),
        model.WithAPIKey("your-api-key"),
        model.WithModel("gpt-4"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 2. 构建 Agent
    ag, err := agent.NewAgentBuilder(chatModel).
        WithName("智能助手").
        WithInstruction("你是 {name}，一个友好的助手").
        Build(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // 3. 运行 Agent
    runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: ag})
    iter := runner.Run(ctx, []*schema.Message{
        schema.UserMessage("你好！"),
    }, adk.WithSessionValues(map[string]any{
        "name": "小智",
    }))

    // 4. 获取响应
    for {
        event, ok := iter.Next()
        if !ok {
            break
        }
        if event.Err != nil {
            log.Fatal(event.Err)
        }
        if event.Output != nil && event.Output.MessageOutput != nil {
            msg, _ := event.Output.MessageOutput.GetMessage()
            if msg != nil {
                println(msg.Content)
            }
        }
    }
}
```

### 带工具的 Agent

```go
// 定义工具
calculatorTool, _ := toolUtils.InferTool("calculator", "执行数学运算",
    func(ctx context.Context, p struct {
        Operation string  `json:"operation"`
        A         float64 `json:"a"`
        B         float64 `json:"b"`
    }) (interface{}, error) {
        return map[string]interface{}{"result": p.A + p.B}, nil
    })

// 构建带工具的 Agent
ag, err := agent.NewAgentBuilder(chatModel).
    WithName("数学助手").
    WithInstruction("你是一个数学助手，使用 calculator 工具进行计算").
    WithTools(calculatorTool).
    WithMaxStep(5).
    Build(ctx)
```

### 带记忆的 Agent

```go
// 创建记忆提供者
memoryProvider := memory.NewMemoryProvider(/* config */)

// 构建带记忆的 Agent
ag, err := agent.NewAgentBuilder(chatModel).
    WithName("记忆助手").
    WithInstruction("你是一个有记忆的助手").
    WithMemory(memoryProvider).  // 自动创建 MemoryMiddleware
    Build(ctx)

// 运行时需要提供 sessionID 和 userID
runner.Run(ctx, messages, adk.WithSessionValues(map[string]any{
    "sessionID": "session-123",
    "userID":    "user-456",
}))
```

### 多 Agent 协作

```go
// 创建子 Agent
mathAgent, _ := agent.NewAgentBuilder(chatModel).
    WithName("数学助手").
    WithDescription("专业的数学计算助手").
    WithInstruction("使用工具进行精确计算").
    WithTools(calculatorTool).
    Build(ctx)

weatherAgent, _ := agent.NewAgentBuilder(chatModel).
    WithName("天气助手").
    WithDescription("专业的天气查询助手").
    WithInstruction("使用工具查询天气").
    WithTools(weatherTool).
    Build(ctx)

// 创建路由 Agent
routerAgent, err := agent.NewAgentBuilder(chatModel).
    WithName("智能路由").
    WithDescription("根据问题类型选择合适的助手").
    WithInstruction("你是路由助手，选择合适的专业助手处理问题").
    WithSubAgents(agent.SubAgentModeSupervisor, mathAgent, weatherAgent).
    Build(ctx)
```

## 中间件机制

### 内置中间件

| 中间件 | 功能 | 文件位置 |
|--------|------|----------|
| `instructionFormatter` | 指令格式化、清理无效技能 | `instruction_formatter.go` |
| `MemoryMiddleware` | 记忆检索和存储 | `memory/middleware.go` |

### 自定义中间件

```go
// 实现中间件接口
type CustomMiddleware struct {
    *adk.BaseChatModelAgentMiddleware
}

func (m *CustomMiddleware) BeforeAgent(ctx context.Context, runCtx *adk.ChatModelAgentContext) (
    context.Context, *adk.ChatModelAgentContext, error) {
    // 在 Agent 执行前处理
    return ctx, runCtx, nil
}

func (m *CustomMiddleware) AfterModelRewriteState(ctx context.Context, state *adk.ChatModelAgentState, mc *adk.ModelContext) (
    context.Context, *adk.ChatModelAgentState, error) {
    // 在模型调用后处理
    return ctx, state, nil
}

// 使用自定义中间件
ag, err := agent.NewAgentBuilder(chatModel).
    WithMiddlewares(&CustomMiddleware{}).
    Build(ctx)
```

### 中间件执行顺序

```html
BeforeAgent (逆序执行)
    │
    ├── instructionFormatter.BeforeAgent
    │
    ├── MemoryMiddleware.BeforeAgent (如果有)
    │
    └── CustomMiddleware.BeforeAgent (如果有)
           │
           ▼
    BeforeModelRewriteState
           │
           ▼
       Chat Model
           │
           ▼
    AfterModelRewriteState
           │
           ├── MemoryMiddleware.AfterModelRewriteState
           │
           └── CustomMiddleware.AfterModelRewriteState
```

## 设计决策

### 为什么使用 Builder 模式？

1. **配置复杂**：Agent 有多个可选配置项，Builder 模式提供清晰的链式 API
2. **类型安全**：编译时检查参数类型
3. **可扩展**：新增配置项只需添加 WithXxx 方法

### 为什么封装 ADK 而不是重新实现？

1. **复用成熟逻辑**：ADK 的 ReAct 循环、工具调用、状态管理已经非常成熟
2. **简化使用**：Builder 模式降低 ADK 复杂的配置门槛
3. **保持兼容**：返回 `adk.Agent` 接口，与 Eino 生态完全兼容

### 为什么需要 instructionFormatter？

1. **优化 Token 使用**：无可用技能时移除 Skill 段落，减少无效 Token
2. **提升 LLM 理解**：XML 标签帮助 LLM 更好理解指令结构
3. **运行时决策**：根据实际状态动态调整指令内容

## 文件结构

```shell
agent/
├── builder.go              # AgentBuilder 构建器
├── instruction_formatter.go # 指令格式化中间件
├── instruction_formatter_test.go # 单元测试
├── types.go                # 常量定义
└── README.md               # 本文档
```

## 相关模块

- [memory](../memory/) - 记忆管理系统
- [model](../model/) - 聊天模型封装
- [tools](../tools/) - 工具集成

## 参考资料

- [CloudWeGo Eino](https://github.com/cloudwego/eino) - Eino 框架源码
- [Eino ADK 文档](https://github.com/cloudwego/eino/tree/main/adk) - Agent Development Kit 文档
