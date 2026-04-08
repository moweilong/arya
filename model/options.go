package model

import "github.com/cloudwego/eino-ext/components/model/openai"

// 模型配置选项
type Options struct {
	Platform   string
	Model      string
	BaseUrl    string
	APIKey     string `json:"apiKey"`
	Dimensions int
	MaxTokens  int

	// openai参数
	ReasoningEffortLevel openai.ReasoningEffortLevel // 推理强度: low, medium, high
	// 深度思考参数
	EnableThinking bool `json:"enable_thinking"`
}

type OptionFunc func(opt *Options)

func WithPlatform(platform string) OptionFunc {
	return func(opt *Options) {
		opt.Platform = platform
	}
}

func WithModel(model string) OptionFunc {
	return func(opt *Options) {
		opt.Model = model
	}
}

func WithBaseUrl(baseUrl string) OptionFunc {
	return func(opt *Options) {
		opt.BaseUrl = baseUrl
	}
}

func WithAPIKey(apiKey string) OptionFunc {
	return func(opt *Options) {
		opt.APIKey = apiKey
	}
}

func WithMaxTokens(maxTokens int) OptionFunc {
	return func(opt *Options) {
		opt.MaxTokens = maxTokens
	}
}

func WithDimensions(dimensions int) OptionFunc {
	return func(opt *Options) {
		opt.Dimensions = dimensions
	}
}

func WithReasoningEffortLevel(reasoningEffortLevel openai.ReasoningEffortLevel) OptionFunc {
	return func(opt *Options) {
		opt.ReasoningEffortLevel = reasoningEffortLevel
	}
}

func WithEnableThinking(enableThinking bool) OptionFunc {
	return func(opt *Options) {
		opt.EnableThinking = enableThinking
	}
}
