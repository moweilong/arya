package card

import "encoding/json"

// https://open.feishu.cn/document/feishu-cards/card-json-v2-structure
type Card struct {
	Schema string `json:"schema"` // 卡片结构的版本声明。默认为 1.0 版本。要使用 JSON 2.0 结构，必须显示声明 2.0。
	Header Header `json:"header"` // 标题组件相关配置。
	Config Config `json:"config"` // 配置卡片的全局行为，包括流式更新模式（JSON 2.0 新增能力）、是否允许被转发、是否为共享卡片等。
	Body   Body   `json:"body"`   // 卡片正文，包含一个名为 elements 的数组，用于放置各类组件。
}

// https://open.feishu.cn/document/feishu-cards/card-json-v2-components/content-components/title
type Header struct {
	Title `json:"title"` // 卡片主标题。必填。要为标题配置多语言，参考配置卡片多语言文档。
}

// Title 标题
type Title struct {
	Tag     string `json:"tag"`     // 文本类型的标签。可选值：plain_text 和 lark_md。
	Content string `json:"content"` // 标题内容。
}

// Config 卡片配置
type Config struct {
	StreamingMode   bool            `json:"streaming_mode"`   // 卡片是否处于流式更新模式，默认值为 false。
	StreamingConfig StreamingConfig `json:"streaming_config"` // 流式更新配置。https://open.feishu.cn/document/cardkit-v1/streaming-updates-openapi-overview
	UpdateMulti     bool            `json:"update_multi"`     // 是否为共享卡片。默认值为 true，JSON 2.0 暂时仅支持设为 true，即更新卡片的内容对所有收到这张卡片的人员可见。
}

// StreamingConfig 流式更新配置
// TODO 字段类型待更正
type StreamingConfig struct {
	PrintFrequencyMs int    `json:"print_frequency_ms,omitempty"` // 流式更新频率，单位：ms
	PrintStep        int    `json:"print_step,omitempty"`         // 流式更新步长，单位：字符数
	PrintStrategy    string `json:"print_strategy,omitempty"`     // 流式更新策略，枚举值，可取：fast/delay
}

// Body 卡片正文
type Body struct {
	Elements []Element `json:"elements"` // 在此传入各个组件的 JSON 数据，组件将按数组顺序纵向流式排列。
}

// Element 卡片元素
type Element struct {
	Tag       string `json:"tag"`        // 组件的标签。
	ElementId string `json:"element_id"` // 操作组件的唯一标识。
	Content   string `json:"content"`    // 内容
}

// ToJSON 序列化为 JSON
func (c *Card) ToJSON() []byte {
	bytes, err := json.Marshal(c)
	if err != nil {
		return []byte("{}")
	}
	return bytes
}

// String 返回 JSON 字符串
func (c *Card) String() string {
	return string(c.ToJSON())
}
