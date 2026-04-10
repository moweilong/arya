package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/moweilong/arya/agent"
	"github.com/moweilong/arya/memory"
	"github.com/moweilong/arya/memory/builtin"
	"github.com/moweilong/arya/memory/builtin/storage"
	"github.com/moweilong/arya/model"
	"github.com/moweilong/arya/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	schemaGrom "gorm.io/gorm/schema"
)

func main() {
	ctx := context.Background()

	// 大语言模型初始化
	aryaPlatform := os.Getenv("ARYA_PLATFORM")
	aryaBaseUrl := os.Getenv("ARYA_BASE_URL")
	aryaApiKey := os.Getenv("ARYA_API_KEY")
	aryaModel := os.Getenv("ARYA_MODEL")
	if aryaPlatform == "" || aryaBaseUrl == "" || aryaApiKey == "" || aryaModel == "" {
		log.Fatal("ARYA_PLATFORM, ARYA_BASE_URL, ARYA_API_KEY and ARYA_MODEL environment variables must be set")
		return
	}
	cm, err := model.NewChatModel(ctx, model.WithPlatform(aryaPlatform),
		model.WithBaseUrl(aryaBaseUrl),
		model.WithAPIKey(aryaApiKey),
		model.WithModel(aryaModel),
	)
	if err != nil {
		log.Fatalf("new chat model fail,err:%s", err)
		return
	}

	// 用户记忆存储
	gormSql, err := NewMysqlGrom("arya:123456@tcp(127.0.0.1:3306)/arya", logger.Silent)
	if err != nil {
		log.Fatalf("创建数据库连接失败: %v", err)
		return
	}
	s, err := storage.NewGormStorage(gormSql)
	if err != nil {
		log.Fatalf("new sql store fail,err:%s", err)
		return
	}
	provider, err := memory.GlobalRegistry().CreateProvider("builtin", &builtin.ProviderConfig{
		ChatModel: cm,
		Storage:   s,
		MemoryConfig: &builtin.MemoryConfig{
			EnableSessionSummary: true,
			EnableUserMemories:   true,
			MemoryLimit:          8,
			Retrieval:            builtin.RetrievalLastN,
		},
	})
	if err != nil {
		log.Fatalf("new provider fail,err:%s", err)
		return
	}
	defer provider.Close()

	sessionID := utils.GetULID()
	ag, err := agent.NewAgentBuilder(cm).
		WithInstruction("你是一名28岁的厦门女孩,名叫mary,是一名短视频剪辑师，主要剪辑电影类。喜欢养狗，家里有一只金毛，它叫小黄。喜欢宅在家里刷剧。不喜欢吃水果。性格内向高冷，别人不问的时候，一般都不回答自己的信息").
		WithMemory(provider).
		Build(ctx)
	if err != nil {
		log.Fatalf("new agent fail,err:%s", err)
		return
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: ag})

	conversations := []*schema.Message{
		schema.UserMessage("你好，我是Alice"),
		schema.UserMessage("我是一名软件工程师，专门做后端开发"),
		schema.UserMessage("我住在北京，今年28岁"),
		schema.UserMessage("你有什么爱好吗?"),
		schema.UserMessage("我喜欢读书和摄影，特别是科幻小说"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{
					Type: "text",
					Text: "这图片里面有什么？",
				},
				{
					Type: "image_url",
					Image: &schema.MessageInputImage{
						MessagePartCommon: schema.MessagePartCommon{
							URL: new("http://10.82.56.134:8111/img/loginBg.6c1d0a90.png"),
						},
					},
				},
			},
		},
		schema.UserMessage("我最近在学习Go语言和云原生技术"),
		schema.UserMessage("我的工作主要涉及微服务架构设计"),
		schema.UserMessage("周末我通常会去公园拍照或者在家看书"),
		schema.UserMessage("你能给我推荐一些适合我的技术书籍吗？"),
		schema.UserMessage("你还记得我之前说过我的职业是什么吗？"),
		schema.UserMessage("基于你对我的了解，你觉得我适合学习什么新技术？"),
		schema.UserMessage("我们年龄相差多少岁呢"),
		schema.UserMessage("你喜欢吃什么水果吗？我喜欢吃苹果"),
		schema.UserMessage("你知道我住哪里吗"),
	}

	for _, conversation := range conversations {
		j, _ := sonic.MarshalString(conversation)
		log.Printf("User: %s", j)
		iter := runner.Run(ctx, []*schema.Message{
			conversation,
		}, adk.WithSessionValues(map[string]any{
			"userID":    sessionID,
			"sessionID": sessionID,
		}))
		var response string
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				log.Fatalf("generate fail,err:%s", event.Err)
				return
			}
			if event.Output != nil && event.Output.MessageOutput != nil {
				if msg, err := event.Output.MessageOutput.GetMessage(); err == nil && msg != nil {
					response = msg.Content
				}
			}
		}
		log.Printf("AI:%s", response)
	}
	time.Sleep(5 * time.Second)
}

func NewMysqlGrom(source string, logLevel logger.LogLevel) (*gorm.DB, error) {
	if !strings.Contains(source, "parseTime") {
		source += "?charset=utf8mb4&parseTime=True&loc=Local"
	}
	gdb, err := gorm.Open(mysql.Open(source), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schemaGrom.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	// 配置GORM日志
	var gormLogger logger.Interface
	if logLevel > 0 {
		gormLogger = logger.Default.LogMode(logLevel)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	gdb.Logger = gormLogger

	return gdb, nil
}
