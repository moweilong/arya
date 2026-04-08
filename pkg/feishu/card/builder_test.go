package card

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.card)
	assert.Equal(t, "2.0", builder.card.Schema)
	assert.True(t, builder.card.Config.UpdateMulti)
}

func TestBuilder_WithSchema(t *testing.T) {
	card := NewBuilder().
		WithSchema("1.0").
		Build()

	assert.Equal(t, "1.0", card.Schema)
}

func TestBuilder_WithTitle(t *testing.T) {
	card := NewBuilder().
		WithTitle("测试标题").
		Build()

	assert.Equal(t, "plain_text", card.Header.Title.Tag)
	assert.Equal(t, "测试标题", card.Header.Title.Content)
}

func TestBuilder_WithTitleMD(t *testing.T) {
	card := NewBuilder().
		WithTitleMD("**粗体标题**").
		Build()

	assert.Equal(t, "lark_md", card.Header.Title.Tag)
	assert.Equal(t, "**粗体标题**", card.Header.Title.Content)
}

func TestBuilder_WithStreamingMode(t *testing.T) {
	card := NewBuilder().
		WithStreamingMode(true).
		Build()

	assert.True(t, card.Config.StreamingMode)
}

func TestBuilder_WithStreamingConfig(t *testing.T) {
	card := NewBuilder().
		WithStreamingConfig(100, 5, "fast").
		Build()

	assert.Equal(t, 100, card.Config.StreamingConfig.PrintFrequencyMs)
	assert.Equal(t, 5, card.Config.StreamingConfig.PrintStep)
	assert.Equal(t, "fast", card.Config.StreamingConfig.PrintStrategy)
}

func TestBuilder_WithUpdateMulti(t *testing.T) {
	card := NewBuilder().
		WithUpdateMulti(false).
		Build()

	assert.False(t, card.Config.UpdateMulti)
}

func TestBuilder_AddElement(t *testing.T) {
	card := NewBuilder().
		AddElement("markdown", "elem1", "内容1").
		AddElement("plain_text", "elem2", "内容2").
		Build()

	assert.Len(t, card.Body.Elements, 2)
	assert.Equal(t, "markdown", card.Body.Elements[0].Tag)
	assert.Equal(t, "elem1", card.Body.Elements[0].ElementId)
	assert.Equal(t, "内容1", card.Body.Elements[0].Content)
	assert.Equal(t, "plain_text", card.Body.Elements[1].Tag)
}

func TestBuilder_AddMarkdownElement(t *testing.T) {
	card := NewBuilder().
		AddMarkdownElement("md1", "**加粗内容**").
		Build()

	assert.Len(t, card.Body.Elements, 1)
	assert.Equal(t, "markdown", card.Body.Elements[0].Tag)
	assert.Equal(t, "md1", card.Body.Elements[0].ElementId)
	assert.Equal(t, "**加粗内容**", card.Body.Elements[0].Content)
}

func TestBuilder_AddPlainTextElement(t *testing.T) {
	card := NewBuilder().
		AddPlainTextElement("text1", "纯文本内容").
		Build()

	assert.Len(t, card.Body.Elements, 1)
	assert.Equal(t, "plain_text", card.Body.Elements[0].Tag)
	assert.Equal(t, "text1", card.Body.Elements[0].ElementId)
	assert.Equal(t, "纯文本内容", card.Body.Elements[0].Content)
}

func TestBuilder_ChainedCalls(t *testing.T) {
	card := NewBuilder().
		WithTitle("链式调用测试").
		WithStreamingMode(true).
		WithStreamingConfig(200, 10, "delay").
		AddMarkdownElement("content", "测试内容").
		AddPlainTextElement("note", "备注").
		Build()

	assert.Equal(t, "链式调用测试", card.Header.Title.Content)
	assert.True(t, card.Config.StreamingMode)
	assert.Equal(t, 200, card.Config.StreamingConfig.PrintFrequencyMs)
	assert.Len(t, card.Body.Elements, 2)
}

func TestCard_ToJSON(t *testing.T) {
	card := NewBuilder().
		WithTitle("测试").
		AddMarkdownElement("md", "内容").
		Build()

	jsonBytes := card.ToJSON()
	assert.NotNil(t, jsonBytes)
	// JSON 字段名根据结构体的 json tag 决定
	assert.Contains(t, string(jsonBytes), "\"Schema\":\"2.0\"")
	assert.Contains(t, string(jsonBytes), "\"content\":\"测试\"")
	assert.Contains(t, string(jsonBytes), "\"tag\":\"markdown\"")
}

func TestCard_String(t *testing.T) {
	card := NewBuilder().
		WithTitle("测试标题").
		Build()

	str := card.String()
	assert.NotEmpty(t, str)
	assert.Contains(t, str, "测试标题")
}

func TestCard_ToJSON_Empty(t *testing.T) {
	// 空卡片仍然会序列化为包含所有零值字段的 JSON
	card := &Card{}
	jsonBytes := card.ToJSON()
	assert.NotEmpty(t, jsonBytes)
	// 不是 {} 而是包含零值字段的 JSON
	assert.Contains(t, string(jsonBytes), "Schema")
}
