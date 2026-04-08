package card

// Builder 卡片构建器
type Builder struct {
	card *Card
}

// NewBuilder 创建卡片构建器
func NewBuilder() *Builder {
	return &Builder{
		card: &Card{
			Schema: "2.0",
			Config: Config{
				UpdateMulti: true,
			},
		},
	}
}

// WithSchema 设置 schema 版本
func (b *Builder) WithSchema(schema string) *Builder {
	b.card.Schema = schema
	return b
}

// WithTitle 设置标题
func (b *Builder) WithTitle(title string) *Builder {
	b.card.Header = Header{
		Title: Title{
			Tag:     "plain_text",
			Content: title,
		},
	}
	return b
}

// WithTitleMD 设置 Markdown 标题
func (b *Builder) WithTitleMD(title string) *Builder {
	b.card.Header = Header{
		Title: Title{
			Tag:     "lark_md",
			Content: title,
		},
	}
	return b
}

// WithStreamingMode 设置流式更新模式
func (b *Builder) WithStreamingMode(enabled bool) *Builder {
	b.card.Config.StreamingMode = enabled
	return b
}

// WithStreamingConfig 设置流式更新配置
func (b *Builder) WithStreamingConfig(frequencyMs, step int, strategy string) *Builder {
	b.card.Config.StreamingConfig = StreamingConfig{
		PrintFrequencyMs: frequencyMs,
		PrintStep:        step,
		PrintStrategy:    strategy,
	}
	return b
}

// WithUpdateMulti 设置是否为共享卡片
func (b *Builder) WithUpdateMulti(updateMulti bool) *Builder {
	b.card.Config.UpdateMulti = updateMulti
	return b
}

// AddElement 添加卡片元素
func (b *Builder) AddElement(tag, elementID, content string) *Builder {
	b.card.Body.Elements = append(b.card.Body.Elements, Element{
		Tag:       tag,
		ElementId: elementID,
		Content:   content,
	})
	return b
}

// AddMarkdownElement 添加 Markdown 元素
func (b *Builder) AddMarkdownElement(elementID, content string) *Builder {
	return b.AddElement("markdown", elementID, content)
}

// AddPlainTextElement 添加纯文本元素
func (b *Builder) AddPlainTextElement(elementID, content string) *Builder {
	return b.AddElement("plain_text", elementID, content)
}

// Build 构建卡片
func (b *Builder) Build() *Card {
	return b.card
}
