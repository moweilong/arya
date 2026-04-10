package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	schemaGrom "gorm.io/gorm/schema"

	aryaAgent "github.com/moweilong/arya/agent"
	"github.com/moweilong/arya/memory"
	"github.com/moweilong/arya/memory/builtin"
	"github.com/moweilong/arya/memory/builtin/storage"
	aryaModel "github.com/moweilong/arya/model"
	"github.com/moweilong/arya/pkg/feishu"
	"github.com/moweilong/arya/pkg/feishu/card"
)

// 流程：
// 1. 用户发送消息
// 2. 飞书接收到消息，创建卡片实体，发送卡片消息（消息内容正在处理中）
// 3. 同步调用 AI Agent，接收 Agent 流式输出，更新卡片实体，推送更新

// Config 应用配置
type Config struct {
	// 飞书配置
	FeishuAppID     string
	FeishuAppSecret string

	// AI 模型配置
	ModelPlatform string
	ModelBaseURL  string
	ModelAPIKey   string
	ModelName     string

	// 记忆存储配置
	MemoryDSN string // MySQL 数据库连接字符串，如: user:password@tcp(127.0.0.1:3306)/dbname
}

// App 应用实例
type App struct {
	config   *Config
	feishu   *feishu.Client
	agent    adk.Agent
	provider memory.MemoryProvider // 记忆提供者
	schemas  map[string]string     // messageID -> cardID 映射
	mu       sync.RWMutex
}

func main() {
	ctx := context.Background()

	// 应用配置
	cfg := &Config{
		FeishuAppID:     "cli_a95ed4d22138dbde",
		FeishuAppSecret: "jkVzBpNq9YEGd2pPmu2x3f1pHNI8UELh",
		ModelPlatform:   "qwen",
		ModelBaseURL:    "http://10.82.21.25:38444/apiaccess/1774592699_qwen3-5-0327/v1",
		ModelAPIKey:     "2000|sk-kRU8qibVEpKUSLWRt7WdHvAV6P3M0QAf",
		ModelName:       "qwen3-5-122B",
		MemoryDSN:       "arya:123456@tcp(127.0.0.1:3306)/arya",
	}

	// 创建应用实例
	app, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("创建应用失败: %v", err)
	}
	defer app.Close() // 确保资源释放

	// 注册事件处理器
	app.RegisterEventHandlers()

	// 启动长连接
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Println("启动飞书长连接...")
	if err := app.feishu.StartStream(ctx); err != nil {
		log.Fatalf("启动长连接失败: %v", err)
	}
}

func NewApp(cfg *Config) (*App, error) {
	ctx := context.Background()

	feishuClient := feishu.NewClient(&feishu.Config{
		AppID:     cfg.FeishuAppID,
		AppSecret: cfg.FeishuAppSecret,
		Debug:     false,
	})

	cm, err := aryaModel.NewChatModel(ctx,
		aryaModel.WithPlatform(cfg.ModelPlatform),
		aryaModel.WithBaseUrl(cfg.ModelBaseURL),
		aryaModel.WithAPIKey(cfg.ModelAPIKey),
		aryaModel.WithModel(cfg.ModelName),
	)
	if err != nil {
		return nil, fmt.Errorf("创建大语言模型失败: %w", err)
	}

	// 创建记忆存储
	provider, err := createMemoryProvider(cm, cfg.MemoryDSN)
	if err != nil {
		return nil, fmt.Errorf("创建记忆存储失败: %w", err)
	}

	// 创建 Agent（带记忆功能）
	ag, err := aryaAgent.NewAgentBuilder(cm).
		WithInstruction("你是一名28岁的厦门女孩,名叫mary,是一名短视频剪辑师，主要剪辑电影类。喜欢养狗，家里有一只金毛，它叫小黄。喜欢宅在家里刷剧。不喜欢吃水果。性格内向高冷，别人不问的时候，一般都不回答自己的信息").
		WithMemory(provider).
		Build(ctx)
	if err != nil {
		return nil, fmt.Errorf("创建 agent 失败: %w", err)
	}

	return &App{
		config:   cfg,
		feishu:   feishuClient,
		agent:    ag,
		provider: provider,
		schemas:  make(map[string]string),
	}, nil
}

// createMemoryProvider 创建记忆提供者
func createMemoryProvider(cm model.ToolCallingChatModel, dsn string) (memory.MemoryProvider, error) {
	// 使用 MySQL 存储
	gormDB, err := newMysqlGorm(dsn, logger.Silent)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	s, err := storage.NewGormStorage(gormDB)
	if err != nil {
		return nil, fmt.Errorf("创建存储失败: %w", err)
	}

	provider, err := memory.GlobalRegistry().CreateProvider("builtin", &builtin.ProviderConfig{
		ChatModel: cm,
		Storage:   s,
		MemoryConfig: &builtin.MemoryConfig{
			EnableSessionSummary: true,
			EnableUserMemories:   true,
			MemoryLimit:          20,
			Retrieval:            builtin.RetrievalLastN,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("创建记忆提供者失败: %w", err)
	}

	return provider, nil
}

// newMysqlGorm 创建 MySQL GORM 连接
func newMysqlGorm(source string, logLevel logger.LogLevel) (*gorm.DB, error) {
	if !strings.Contains(source, "parseTime") {
		source += "?charset=utf8mb4&parseTime=True&loc=Local"
	}
	gdb, err := gorm.Open(mysql.Open(source), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schemaGrom.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		return nil, err
	}

	// 配置 GORM 日志
	var gormLogger logger.Interface
	if logLevel > 0 {
		gormLogger = logger.Default.LogMode(logLevel)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	gdb.Logger = gormLogger

	return gdb, nil
}

// Close 关闭应用资源
func (app *App) Close() {
	if app.provider != nil {
		app.provider.Close()
	}
}

// RegisterEventHandlers 注册事件处理器
func (app *App) RegisterEventHandlers() {
	app.feishu.RegisterEvents(&feishu.EventConfig{
		OnMessageReceive:    app.handleMessageReceive,
		CustomEventHandlers: map[string]func(ctx context.Context, event *larkevent.EventReq) error{
			// 可添加其他自定义事件处理
		},
	})
}

// handleMessageReceive 处理接收到的消息
func (app *App) handleMessageReceive(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
	// 解析消息内容
	msgContent := parseMessageContent(event.Event.Message.Content)
	if msgContent == "" {
		return nil
	}

	// 获取发送者信息
	senderID := getSenderID(event.Event.Sender)
	if senderID == "" {
		return nil
	}

	// 获取群聊ID
	chatID := getChatID(event.Event.Message)

	log.Printf("收到消息 - 发送者: %s, 群聊: %s, 内容: %s", senderID, chatID, msgContent)

	// 异步处理消息，避免阻塞事件处理
	go app.processMessage(chatID, senderID, msgContent)

	return nil
}

// processMessage 处理消息并响应
// sessionID 用于标识对话会话，这里使用 chatID + senderID 作为会话标识，确保同一用户在同一群聊中的对话历史连续
func (app *App) processMessage(chatID, senderID, content string) {
	ctx := context.Background()

	// 生成会话 ID：同一用户在同一群聊中使用相同会话
	sessionID := fmt.Sprintf("%s_%s", chatID, senderID)

	// 1. 创建卡片实体
	c := card.NewBuilder().
		WithTitle("AI 助手").
		WithStreamingMode(true).
		AddMarkdownElement("markdown", "正在思考...").
		Build()
	cardID, err := app.feishu.Card().Create(ctx, c)
	if err != nil {
		log.Printf("创建卡片失败: %v", err)
		return
	}

	// 2. 发送卡片消息给用户
	messageID, err := app.feishu.Message().SendCardToChat(ctx, chatID, buildCardContent(cardID))
	if err != nil {
		log.Printf("发送卡片消息失败: %v", err)
		return
	}

	// 保存映射关系
	app.mu.Lock()
	app.schemas[messageID] = cardID
	app.mu.Unlock()

	log.Printf("发送卡片成功 - messageID: %s, cardID: %s", messageID, cardID)

	// 3. 调用 AI Agent 流式输出（带会话记忆）
	app.streamAIResponse(ctx, cardID, senderID, sessionID, content)
}

// streamAIResponse 流式获取 AI 响应并更新卡片
func (app *App) streamAIResponse(ctx context.Context, cardID, userID, sessionID, userMessage string) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: app.agent, EnableStreaming: true})

	// 传入会话信息，启用记忆功能
	stream := runner.Run(ctx, []adk.Message{schema.UserMessage(userMessage)},
		adk.WithSessionValues(map[string]any{
			"userID":    userID,
			"sessionID": sessionID,
		}))

	// 收集完整响应
	var fullContent strings.Builder
	var lastContent string

	// 读取流式响应
	for {
		event, ok := stream.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			log.Printf("AI 响应错误: %v", event.Err)
			break
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			if msg, err := event.Output.MessageOutput.GetMessage(); err == nil && msg != nil {
				fullContent.WriteString(msg.Content)

				// 节流更新：避免频繁更新卡片
				currentContent := fullContent.String()
				if len(currentContent)-len(lastContent) > 10 || lastContent == "" {
					lastContent = currentContent
					if err := app.feishu.Card().UpdateContent(ctx, cardID, currentContent); err != nil {
						log.Printf("更新卡片失败: %v", err)
					}
				}
			}
		}
	}

	// 最终更新
	finalContent := fullContent.String()
	if finalContent != lastContent {
		if err := app.feishu.Card().UpdateContent(ctx, cardID, finalContent); err != nil {
			log.Printf("最终更新卡片失败: %v", err)
		}
	}

	log.Printf("AI 响应完成 - cardID: %s, userID: %s, sessionID: %s", cardID, userID, sessionID)
}

// updateCardWithError 更新卡片显示错误信息
func (app *App) updateCardWithError(ctx context.Context, cardID, errMsg string) {
	_ = app.feishu.Card().UpdateContent(ctx, cardID, fmt.Sprintf("❌ %s", errMsg))
}

// parseMessageContent 解析消息内容
func parseMessageContent(content *string) string {
	if content == nil || *content == "" {
		return ""
	}

	// 尝试解析 JSON 格式的消息内容
	var msg struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal([]byte(*content), &msg); err == nil {
		return msg.Text
	}

	// 如果不是 JSON，直接返回原文
	return *content
}

// getSenderID 获取发送者 ID
func getSenderID(sender *larkim.EventSender) string {
	if sender == nil {
		return ""
	}

	fmt.Printf("发送人：%s,%s,%s", *sender.SenderId.OpenId, *sender.SenderId.UnionId, *sender.SenderId.UserId)

	// 优先使用 Open ID
	if sender.SenderId != nil && sender.SenderId.OpenId != nil {
		return *sender.SenderId.OpenId
	}

	// 其次使用 User ID
	if sender.SenderId != nil && sender.SenderId.UserId != nil {
		return *sender.SenderId.UserId
	}

	return ""
}

// getChatID 获取群聊ID
func getChatID(msg *larkim.EventMessage) string {
	if msg == nil {
		return ""
	}

	if msg.ChatId != nil {
		return *msg.ChatId
	}
	return ""
}

// buildCardContent 构建卡片内容
func buildCardContent(cardID string) string {
	// 构建卡片 JSON，引用已创建的卡片实体
	cardData := map[string]interface{}{
		"type": "card",
		"data": map[string]interface{}{
			"card_id": cardID,
		},
	}

	jsonBytes, _ := json.Marshal(cardData)
	return string(jsonBytes)
}
