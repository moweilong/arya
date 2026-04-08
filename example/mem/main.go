package main

import (
	"context"
	"log"
	"os"
)

func main() {
	ctx := context.Background()
	baseUrl := os.Getenv("ARYA_BASE_URL")
	apiKey := os.Getenv("ARYA_API_KEY")
	model := os.Getenv("ARYA_MODEL")
	if baseUrl == "" || apiKey == "" || model == "" {
		log.Fatal("BaseUrl and APIKey environment variables must be set")
		return
	}

	cm, err := model.NewChatModel(model.WithBaseUrl(baseUrl),
		model.WithAPIKey(apiKey),
		model.WithModel("gpt-5-nano"),
	)
	if err != nil {
		log.Fatalf("new chat model fail,err:%s", err)
		return
	}
}
