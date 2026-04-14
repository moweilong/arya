package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func newLogHandler() callbacks.Handler {
	builder := callbacks.NewHandlerBuilder()
	builder.
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			if info.Component == components.ComponentOfTool {
				tci := tool.ConvCallbackInput(input)
				fmt.Printf("\033[32m[callback]: start [%s:%s:%s] input: %v\033[0m\n", info.Component, info.Type, info.Name, tci.ArgumentsInJSON)
			} else {
				fmt.Printf("\033[32m[callback]: start [%s:%s:%s]running...\033[0m\n", info.Component, info.Type, info.Name)
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			if info.Component == components.ComponentOfTool {
				tco := tool.ConvCallbackOutput(output)
				fmt.Printf("\033[32m[callback]: end [%s:%s:%s] result: %v\033[0m\n", info.Component, info.Type, info.Name, tco.Response)
			} else if info.Component == components.ComponentOfChatModel {
				cco := model.ConvCallbackOutput(output)
				fmt.Printf("\033[32m[callback]: end [%s:%s:%s] output: %s\033[0m\n", info.Component, info.Type, info.Name, printMessage(cco.Message))
			} else {
				fmt.Printf("\033[32m[callback]: end [%s:%s:%s]\033[0m\n", info.Component, info.Type, info.Name)
			}
			return ctx
		}).
		OnStartWithStreamInputFn(func(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
			input.Close()
			fmt.Printf("\033[32m[callback]: start stream input [%s:%s:%s]running...\033[0m\n", info.Component, info.Type, info.Name)
			return ctx
		}).
		OnEndWithStreamOutputFn(func(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
			output.Close()
			fmt.Printf("\033[32m[callback]: end stream output [%s:%s:%s]\033[0m\n", info.Component, info.Type, info.Name)
			return ctx
		}).
		OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
			fmt.Printf("\033[31m[callback]: error [%s:%s:%s] - %v\033[0m\n", info.Component, info.Type, info.Name, err)
			return ctx
		})
	return builder.Build()
}

func printMessage(m *schema.Message) string {
	sb := strings.Builder{}
	sb.WriteString("[")
	sb.WriteString(string(m.Role))
	sb.WriteString("]")
	if len(m.Content) > 0 {
		sb.WriteString("Content: \"")
		sb.WriteString(m.Content)
		sb.WriteString("\" ")
	}
	if len(m.ToolCalls) > 0 {
		sb.WriteString("ToolCalls: [")
		for _, toolCall := range m.ToolCalls {
			sb.WriteString("{Name: ")
			sb.WriteString(toolCall.Function.Name)
			sb.WriteString(", Arguments: ")
			sb.WriteString(toolCall.Function.Arguments)
			sb.WriteString("}")
		}
		sb.WriteString("]")
	}
	return sb.String()
}
