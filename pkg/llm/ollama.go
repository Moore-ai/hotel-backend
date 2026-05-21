package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaProvider 是与 Ollama API 通信的客户端
type OllamaProvider struct {
	config Config
	http   *http.Client
}

// NewOllamaProvider 创建 Ollama 客户端
func NewOllamaProvider(cfg Config) *OllamaProvider {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 120 * time.Second
	}
	return &OllamaProvider{
		config: cfg,
		http:   &http.Client{Timeout: timeout},
	}
}

// ollamaMessage 是 Ollama 的消息格式（content 为简单字符串）
type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaResponse 是 Ollama Chat API 响应体
type ollamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
}

// Chat 发送消息到 Ollama API 并返回结果
func (c *OllamaProvider) Chat(systemPrompt string, messages []Message, tools []ToolDef) (*ChatResult, error) {
	if c.config.BaseURL == "" {
		return nopResult(), nil
	}

	// 构建 Ollama 格式的消息列表
	ollamaMsgs := make([]ollamaMessage, 0, len(messages)+1)
	if systemPrompt != "" {
		ollamaMsgs = append(ollamaMsgs, ollamaMessage{Role: RoleSystem, Content: systemPrompt})
	}
	for _, msg := range messages {
		text := extractText(msg)
		ollamaMsgs = append(ollamaMsgs, ollamaMessage{Role: msg.Role, Content: text})
	}

	body, err := json.Marshal(map[string]any{
		"model":    c.config.Model,
		"messages": ollamaMsgs,
		"stream":   false,
		"tools":    tools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.config.BaseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var or ollamaResponse
	if err := json.Unmarshal(respBody, &or); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	content := []ContentBlock{
		{Type: ContentTypeText, Text: or.Message.Content},
	}

	return &ChatResult{
		Content:    content,
		Text:       or.Message.Content,
		Role:       RoleAssistant,
		StopReason: StopReasonEndTurn,
	}, nil
}

// extractText 从 Message 的 content blocks 中提取文本内容
func extractText(msg Message) string {
	text, _ := extractContent(msg)
	return text
}
