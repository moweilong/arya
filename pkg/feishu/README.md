# 飞书功能封装

## 使用

```go
    client := feishu.NewClient(&feishu.Config{
      AppID:     "your-app-id",
      AppSecret: "your-app-secret",
  })

  // 构建卡片
  c := card.NewBuilder().
      WithTitle("AI 助手").
      AddMarkdownElement("content", "Hello World").
      Build()

  // 发送卡片给用户
  messageID, err := client.Message().SendCardToUser(ctx, "user-id", c.String())

  // 发送文本消息
  messageID, err := client.Message().SendText(ctx, message.ReceiveIdTypeUserID, "user-id", "Hello")
```


```go
// 创建客户端
  client := feishu.NewClient(&feishu.Config{
      AppID:     "your-app-id",
      AppSecret: "your-app-secret",
      Debug:     true,
  })

  // 注册事件处理器
  client.RegisterEvents(&feishu.EventConfig{
      OnMessageReceive: func(ctx context.Context, event *larkim.P2MessageReceiveV1) error {
          // 处理接收到的消息
          fmt.Printf("收到消息: %s\n", event.Event.Message.Content)
          return nil
      },
      CustomEventHandlers: map[string]func(ctx context.Context, event *larkevent.EventReq) error{
          "approval": func(ctx context.Context, event *larkevent.EventReq) error {
              // 处理审批事件
              return nil
          },
      },
  })

  // 启动长连接
  ctx, cancel := context.WithCancel(context.Background())
  defer cancel()

  if err := client.StartStream(ctx); err != nil {
      log.Fatalf("启动长连接失败: %v", err)
  }
```
