package card

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCard_MarshalJSON(t *testing.T) {
	card := &Card{
		Schema: "2.0",
		Header: Header{
			Title: Title{
				Tag:     "plain_text",
				Content: "测试标题",
			},
		},
		Config: Config{
			StreamingMode: true,
			UpdateMulti:   true,
			StreamingConfig: StreamingConfig{
				PrintFrequencyMs: 100,
				PrintStep:        5,
				PrintStrategy:    "fast",
			},
		},
		Body: Body{
			Elements: []Element{
				{
					Tag:       "markdown",
					ElementId: "content",
					Content:   "测试内容",
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(card)
	assert.NoError(t, err)

	// 验证 JSON 包含关键字段
	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, "\"Schema\":\"2.0\"")
	assert.Contains(t, jsonStr, "\"tag\":\"plain_text\"")
	assert.Contains(t, jsonStr, "\"content\":\"测试标题\"")
	assert.Contains(t, jsonStr, "\"streaming_mode\":true")
	assert.Contains(t, jsonStr, "\"update_multi\":true")
	assert.Contains(t, jsonStr, "\"tag\":\"markdown\"")
	assert.Contains(t, jsonStr, "\"element_id\":\"content\"")
}

func TestTitle_JSONTags(t *testing.T) {
	title := Title{
		Tag:     "lark_md",
		Content: "**粗体**",
	}

	jsonBytes, err := json.Marshal(title)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), `"tag":"lark_md"`)
	assert.Contains(t, string(jsonBytes), `"content":"**粗体**"`)
}

func TestConfig_JSONTags(t *testing.T) {
	config := Config{
		StreamingMode: true,
		UpdateMulti:   false,
		StreamingConfig: StreamingConfig{
			PrintFrequencyMs: 50,
			PrintStep:        10,
			PrintStrategy:    "delay",
		},
	}

	jsonBytes, err := json.Marshal(config)
	assert.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"streaming_mode":true`)
	assert.Contains(t, jsonStr, `"update_multi":false`)
	assert.Contains(t, jsonStr, `"print_frequency_ms":50`)
	assert.Contains(t, jsonStr, `"print_step":10`)
	assert.Contains(t, jsonStr, `"print_strategy":"delay"`)
}

func TestElement_JSONTags(t *testing.T) {
	elem := Element{
		Tag:       "markdown",
		ElementId: "elem-1",
		Content:   "元素内容",
	}

	jsonBytes, err := json.Marshal(elem)
	assert.NoError(t, err)

	jsonStr := string(jsonBytes)
	assert.Contains(t, jsonStr, `"tag":"markdown"`)
	assert.Contains(t, jsonStr, `"element_id":"elem-1"`)
	assert.Contains(t, jsonStr, `"content":"元素内容"`)
}

func TestStreamingConfig_Omitempty(t *testing.T) {
	// 测试零值字段是否被省略
	config := StreamingConfig{}

	jsonBytes, err := json.Marshal(config)
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(jsonBytes))

	// 测试有值时包含字段
	config = StreamingConfig{
		PrintFrequencyMs: 100,
	}
	jsonBytes, err = json.Marshal(config)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "print_frequency_ms")
	assert.NotContains(t, string(jsonBytes), "print_step")
}
