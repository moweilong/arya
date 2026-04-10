package model

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
)

func NewChatModel(ctx context.Context, opts ...OptionFunc) (model.ToolCallingChatModel, error) {
	o := &Options{}
	for _, opt := range opts {
		opt(o)
	}

	switch o.Platform {
	case "openai":
		return getChatByOpenai(ctx, o)
	case "qwen":
		return getChatByQwen(ctx, o)
	}
	return getChatByOpenai(ctx, o)
}

func getChatByOpenai(ctx context.Context, o *Options) (model.ToolCallingChatModel, error) {
	param := &openai.ChatModelConfig{
		APIKey:          o.APIKey, // OpenAI API 密钥
		BaseURL:         o.BaseUrl,
		Model:           o.Model,
		ReasoningEffort: o.ReasoningEffortLevel,
	}

	if o.ReasoningEffortLevel != "" {
		param.ReasoningEffort = o.ReasoningEffortLevel
	}

	if o.MaxTokens > 0 {
		param.MaxTokens = &o.MaxTokens
	}

	cm, err := openai.NewChatModel(ctx, param)
	return cm, err
}

func getChatByQwen(ctx context.Context, o *Options) (model.ToolCallingChatModel, error) {
	param := &qwen.ChatModelConfig{
		APIKey:         o.APIKey,
		BaseURL:        o.BaseUrl,
		Model:          o.Model,
		EnableThinking: &o.EnableThinking,
	}

	if o.MaxTokens > 0 {
		param.MaxTokens = &o.MaxTokens
	}

	cm, err := qwen.NewChatModel(ctx, param)
	return cm, err
}
