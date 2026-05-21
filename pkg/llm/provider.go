package llm

import (
	"log"
	"time"
)

// Provider 是 LLM 服务商接口，支持多 provider
type Provider interface {
	Chat(systemPrompt string, messages []Message, tools []ToolDef) (*ChatResult, error)
}

// Provider 常量
const (
	ProviderAnthropic = "anthropic"
	ProviderOllama    = "ollama"
	ProviderOpenAI    = "openai"
)

// Config 是 LLM 客户端的配置
type Config struct {
	Provider  string
	BaseURL   string
	APIKey    string
	Model     string
	MaxTokens int
	Timeout   time.Duration
}

// NewProvider 创建指定 provider 的 LLM 客户端
func NewProvider(cfg Config) Provider {
	switch cfg.Provider {
	case ProviderOllama:
		return NewOllamaProvider(cfg)
	case ProviderOpenAI:
		return NewOpenAIProvider(cfg)
	case ProviderAnthropic, "":
		return NewAnthropicProvider(cfg)
	default:
		log.Printf("Warning: unknown LLM provider %q, falling back to anthropic", cfg.Provider)
		return NewAnthropicProvider(cfg)
	}
}

func nopResult() *ChatResult {
	return &ChatResult{
		Content: []ContentBlock{{Type: ContentTypeText, Text: "nop mode response"}},
		Text:    "nop mode response",
		Role:    RoleAssistant,
	}
}
