// Package message 提供飞书消息发送功能。
//
// 支持向不同用户发送不同类型的消息，包括文本、卡片等。
//
// # 创建消息服务
//
// 消息服务通常通过 feishu.Client 获取：
//
//	client := feishu.NewClient(&feishu.Config{
//	    AppID:     "your-app-id",
//	    AppSecret: "your-app-secret",
//	})
//	msgService := client.Message()
//
// # 发送卡片消息
//
// 结合 card 包发送卡片消息：
//
//	// 构建卡片
//	c := card.NewBuilder().
//	    WithTitle("AI 助手").
//	    AddMarkdownElement("content", "Hello World").
//	    Build()
//
//	// 发送给用户
//	messageID, err := msgService.SendCardToUser(ctx, "user-id", c.String())
//
// # 接收人类型
//
// 支持多种接收人类型：
//   - ReceiveIdTypeUserID: 用户ID
//   - ReceiveIdTypeOpenID: Open ID
//   - ReceiveIdTypeChatID: 群聊ID
//   - ReceiveIdTypeEmail: 邮箱地址
//
// # 发送文本消息
//
//	messageID, err := msgService.SendText(ctx, message.ReceiveIdTypeUserID, "user-id", "Hello")
//
// # 联动卡片更新
//
// 发送卡片后，可以使用返回的 messageID 更新卡片内容：
//
//	// 发送卡片，获取 messageID
//	messageID, err := msgService.SendCardToUser(ctx, "user-id", cardContent)
//
//	// 后续更新卡片内容（需要 card 模块支持）
//	// 注意：更新卡片需要使用 card_id，而非 message_id
//	// card_id 可从发送卡片的响应中获取，或通过卡片回传获取
//
// 更多消息类型和配置，参考飞书官方文档：
// https://open.feishu.cn/document/server-docs/im-v1/message/create
package message
