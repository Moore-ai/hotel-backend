package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicClient 是与 Anthropic Messages API 通信的客户端
type AnthropicClient struct {
	config Config
	http   *http.Client
}

// Config 是 LLM 客户端的配置
type Config struct {
	BaseURL   string
	APIKey    string
	Model     string
	MaxTokens int
	Timeout   time.Duration
}

// NewAnthropicClient 创建新的 LLM 客户端
func NewAnthropicClient(cfg Config) *AnthropicClient {
	return &AnthropicClient{
		config: cfg,
		http: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// anthropicRequest 是 Anthropic Messages API 请求体
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Tools     []ToolDef `json:"tools,omitempty"`
}

// anthropicResponse 是 Anthropic Messages API 响应体
type anthropicResponse struct {
	Content    []ContentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Role       string         `json:"role"`
	Error      *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat 发送消息到 Anthropic API 并返回结果
func (c *AnthropicClient) Chat(systemPrompt string, messages []Message, tools []ToolDef) (*ChatResult, error) {
	// nop 模式：baseURL 为空时跳过 API 调用
	if c.config.BaseURL == "" {
		return &ChatResult{
			Content: []ContentBlock{{Type: ContentTypeText, Text: "nop mode response"}},
			Text:    "nop mode response",
			Role:    RoleAssistant,
		}, nil
	}

	reqBody := anthropicRequest{
		Model:     c.config.Model,
		MaxTokens: c.config.MaxTokens,
		System:    systemPrompt,
		Messages:  messages,
		Tools:     tools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.config.BaseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("llm returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var ar anthropicResponse
	if err := json.Unmarshal(respBody, &ar); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if ar.Error != nil {
		return nil, fmt.Errorf("llm error: %s", ar.Error.Message)
	}

	// 解析内容块
	result := &ChatResult{
		Content:    ar.Content,
		Role:       ar.Role,
		StopReason: ar.StopReason,
	}

	for _, block := range ar.Content {
		if block.Type == ContentTypeText && result.Text == "" {
			result.Text = block.Text
		}
		if block.Type == ContentTypeToolUse {
			result.ToolCall = &ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			}
			break // 只取第一个 tool use
		}
	}

	return result, nil
}
