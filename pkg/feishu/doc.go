// Package feishu 提供飞书 API 的 Go 封装
//
// 飞书是一个企业协作平台，提供消息、卡片、机器人、用户等多种能力。
// 本包按照功能模块进行组织：
//   - card: 卡片相关功能
//   - message: 消息相关功能（预留）
//   - bot: 机器人相关功能（预留）
//
// 使用示例:
//
//	client := feishu.NewClient(&feishu.Config{
//	    AppID:     "your-app-id",
//	    AppSecret: "your-app-secret",
//	})
//
//	// 使用卡片服务
//	cardID, err := client.Card().CreateWithContent(ctx, "Hello World", "标题")
package feishu
