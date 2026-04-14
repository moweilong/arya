package main

import (
	"context"
	"log"
	"os"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/joho/godotenv"

	"github.com/moweilong/arya/agent"
	"github.com/moweilong/arya/model"
	"github.com/moweilong/arya/tools/shell"
)

func main() {
	ctx := context.Background()

	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Printf("警告: 无法加载 .env 文件: %v", err)
	}

	aryaBaseUrl := os.Getenv("ARYA_BASE_URL")
	aryaApiKey := os.Getenv("ARYA_API_KEY")
	aryaModel := os.Getenv("ARYA_MODEL")
	aryaPlatform := os.Getenv("ARYA_PLATFORM")
	if aryaPlatform == "" || aryaBaseUrl == "" || aryaApiKey == "" || aryaModel == "" {
		log.Fatal("ARYA_PLATFORM, ARYA_BASE_URL, ARYA_API_KEY and ARYA_MODEL environment variables must be set")
		return
	}
	cm, err := model.NewChatModel(ctx, model.WithBaseUrl(aryaBaseUrl),
		model.WithAPIKey(aryaApiKey),
		model.WithModel(aryaModel),
		model.WithPlatform(aryaPlatform),
	)
	if err != nil {
		log.Fatalf("new chat model fail,err:%s", err)
		return
	}

	// 合并所有工具：创建时一次性设置
	allTools := append(shell.GetExecuteTools(), shell.GetTools()...)

	ag, err := agent.NewAgentBuilder(cm).
		WithInstruction("你是一个linux大师").
		WithTools(allTools...).
		Build(ctx)
	if err != nil {
		log.Fatalf("new agent fail,err:%s", err)
		return
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: ag})

	conversations := []string{
		"帮我看一下当前目录有什么文件",
		"帮我看一下内存使用情况",
	}

	for _, conversation := range conversations {
		log.Printf("User: %s", conversation)
		iter := runner.Run(ctx, []*schema.Message{
			schema.UserMessage(conversation),
		})
		var response string
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				log.Fatalf("generate fail,err:%s", event.Err)
			}
			if event.Output != nil && event.Output.MessageOutput != nil {
				if msg, err := event.Output.MessageOutput.GetMessage(); err == nil && msg != nil {
					response = msg.Content
				}
			}
		}
		log.Printf("AI:%s", response)
	}
}
