// Package card 提供飞书卡片的构建和操作功能。
//
// 飞书卡片是一种富媒体消息格式，支持文本、图片、按钮等多种组件，
// 可以用于展示结构化信息、收集用户输入等场景。
//
// # 创建卡片服务
//
// 卡片服务需要依赖 lark.Client 进行 API 调用：
//
//	larkClient := lark.NewClient("app-id", "app-secret")
//	cardService := card.NewCardService(larkClient)
//
// # 使用构建器创建卡片
//
// 使用 Builder 可以灵活构建复杂的卡片结构：
//
//	// 创建一个带流式更新的卡片
//	c := card.NewBuilder().
//	    WithTitle("AI 助手").
//	    WithStreamingMode(true).
//	    WithStreamingConfig(100, 5, "fast").
//	    AddMarkdownElement("content", "正在思考...").
//	    Build()
//
//	cardID, err := cardService.Create(ctx, c)
//
// # 快捷创建简单卡片
//
// 对于简单的 Markdown 卡片，可以使用快捷方法：
//
//	cardID, err := cardService.CreateWithContent(ctx, "Hello World", "标题")
//
// # 更新卡片内容
//
// 更新整个卡片或指定元素：
//
//	// 更新默认 markdown 元素
//	err := cardService.UpdateContent(ctx, cardID, "新的内容")
//
//	// 更新指定元素
//	err := cardService.UpdateElement(ctx, cardID, "content", "更新的内容")
//
// # 卡片结构
//
// 飞书卡片 JSON 2.0 结构包含以下主要部分：
//   - Header: 标题区域
//   - Body: 正文区域，包含多个元素
//   - Config: 配置项，如流式更新、共享设置等
//
// 更多卡片组件和配置，参考飞书官方文档：
// https://open.feishu.cn/document/feishu-cards/card-json-v2-structure
package card
