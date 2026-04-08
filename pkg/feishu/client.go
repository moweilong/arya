package feishu

import (
	"context"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"

	"github.com/moweilong/arya/pkg/feishu/card"
	"github.com/moweilong/arya/pkg/feishu/message"
)

// Client 飞书客户端封装
type Client struct {
	httpClient *lark.Client
	wsClient   *larkws.Client
	config     *Config

	// 服务
	cardService    *card.CardService
	messageService *message.MessageService

	// 事件处理器
	eventHandler *dispatcher.EventDispatcher
}

// Config 飞书客户端配置
type Config struct {
	AppID     string
	AppSecret string
	Debug     bool
}

// NewClient 创建飞书客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = &Config{}
	}

	c := &Client{
		config: config,
	}

	// 创建 lark 客户端
	if config.AppID != "" && config.AppSecret != "" {
		var opts []lark.ClientOptionFunc
		if config.Debug {
			opts = append(opts, lark.WithLogLevel(larkcore.LogLevelDebug))
		}
		c.httpClient = lark.NewClient(config.AppID, config.AppSecret, opts...)
		c.cardService = card.NewCardService(c.httpClient)
		c.messageService = message.NewMessageService(c.httpClient)
	}

	return c
}

// NewClientWithOptions 使用选项创建飞书客户端
func NewClientWithOptions(opts ...Option) *Client {
	config := &Config{}
	for _, opt := range opts {
		opt(config)
	}
	return NewClient(config)
}

// GetLarkClient 获取底层 lark 客户端
func (c *Client) GetLarkClient() *lark.Client {
	return c.httpClient
}

// Card 返回卡片服务
func (c *Client) Card() *card.CardService {
	return c.cardService
}

// Message 返回消息服务
func (c *Client) Message() *message.MessageService {
	return c.messageService
}

// EventConfig 事件配置
type EventConfig struct {
	// 消息接收处理
	OnMessageReceive func(ctx context.Context, event *larkim.P2MessageReceiveV1) error

	// 自定义事件处理
	CustomEventHandlers map[string]func(ctx context.Context, event *larkevent.EventReq) error
}

// RegisterEvents 注册事件处理器
func (c *Client) RegisterEvents(cfg *EventConfig) {
	if cfg == nil {
		return
	}

	builder := dispatcher.NewEventDispatcher(c.config.AppID, c.config.AppSecret)

	// 注册消息接收事件
	if cfg.OnMessageReceive != nil {
		builder.OnP2MessageReceiveV1(cfg.OnMessageReceive)
	}

	// 注册自定义事件
	for eventType, handler := range cfg.CustomEventHandlers {
		builder.OnCustomizedEvent(eventType, handler)
	}

	c.eventHandler = builder
}

// StartStream 启动长连接接收事件
// ctx 用于控制生命周期，调用 cancel() 可停止长连接
func (c *Client) StartStream(ctx context.Context) error {
	if c.config.AppID == "" || c.config.AppSecret == "" {
		return fmt.Errorf("AppID 和 AppSecret 不能为空")
	}

	// 如果没有注册事件处理器，使用默认的空处理器
	if c.eventHandler == nil {
		c.eventHandler = dispatcher.NewEventDispatcher(c.config.AppID, c.config.AppSecret)
	}

	// 创建 WebSocket 客户端选项
	opts := []larkws.ClientOption{
		larkws.WithEventHandler(c.eventHandler),
	}
	if c.config.Debug {
		opts = append(opts, larkws.WithLogLevel(larkcore.LogLevelDebug))
	} else {
		opts = append(opts, larkws.WithLogLevel(larkcore.LogLevelInfo))
	}

	// 创建 WebSocket 客户端
	c.wsClient = larkws.NewClient(c.config.AppID, c.config.AppSecret, opts...)

	// 启动客户端（阻塞直到 ctx 取消）
	return c.wsClient.Start(ctx)
}
