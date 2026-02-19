package ai

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

var client *anthropic.Client

func Init() {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Println("ANTHROPIC_API_KEY is not set — AI features will be unavailable")
		return
	}

	c := anthropic.NewClient(option.WithAPIKey(apiKey))
	client = &c
	log.Println("Anthropic AI 기능 활성화 완료")
}

func Close() {
	client = nil
}

func IsAvailable() bool {
	return client != nil
}

type HistoryEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func ChatStream(ctx context.Context, history []HistoryEntry, message string) (<-chan string, <-chan error) {
	chunks := make(chan string, 16)
	errc := make(chan error, 1)

	go func() {
		defer close(chunks)
		defer close(errc)

		if client == nil {
			errc <- fmt.Errorf("Anthropic API key is not configured")
			return
		}

		var systemPrompt string
		var messages []anthropic.MessageParam

		for _, h := range history {
			switch h.Role {
			case "system":
				systemPrompt = h.Content
			case "user":
				messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Content)))
			case "assistant", "model":
				messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Content)))
			}
		}

		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(message)))

		params := anthropic.MessageNewParams{
			Model:     anthropic.ModelClaude4Sonnet20250514,
			MaxTokens: 4096,
			Messages:  messages,
		}

		if systemPrompt != "" {
			params.System = []anthropic.TextBlockParam{
				{Text: systemPrompt},
			}
		}

		stream := client.Messages.NewStreaming(ctx, params)

		for stream.Next() {
			event := stream.Current()
			switch variant := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch delta := variant.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					if delta.Text != "" {
						select {
						case chunks <- delta.Text:
						case <-ctx.Done():
							return
						}
					}
				}
			}
		}

		if stream.Err() != nil {
			errc <- fmt.Errorf("스트리밍 오류: %w", stream.Err())
		}
	}()

	return chunks, errc
}
