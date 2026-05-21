package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OpenAIProvider 是与 OpenAI Chat Completions API 通信的客户端
type OpenAIProvider struct {
	config Config
	http   *http.Client
}

// NewOpenAIProvider 创建 OpenAI 客户端
func NewOpenAIProvider(cfg Config) *OpenAIProvider {
	return &OpenAIProvider{
		config: cfg,
		http:   &http.Client{Timeout: cfg.Timeout},
	}
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    *string          `json:"content"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openaiToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openaiFunctionCall `json:"function"`
}

type openaiFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string             `json:"type"`
	Function openaiToolFunction `json:"function"`
}

type openaiToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

type openaiRequest struct {
	Model     string          `json:"model"`
	Messages  []openaiMessage `json:"messages"`
	Tools     []openaiTool    `json:"tools,omitempty"`
	MaxTokens int             `json:"max_tokens,omitempty"`
}

type openaiChoice struct {
	Index        int           `json:"index"`
	FinishReason string        `json:"finish_reason"`
	Message      openaiMessage `json:"message"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func convertTools(tools []ToolDef) []openaiTool {
	if len(tools) == 0 {
		return nil
	}
	result := make([]openaiTool, len(tools))
	for i, t := range tools {
		result[i] = openaiTool{
			Type: "function",
			Function: openaiToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		}
	}
	return result
}

// Chat 发送消息到 OpenAI API 并返回结果
func (c *OpenAIProvider) Chat(systemPrompt string, messages []Message, tools []ToolDef) (*ChatResult, error) {
	if c.config.BaseURL == "" {
		return nopResult(), nil
	}

	openaiMsgs := make([]openaiMessage, 0, len(messages)+1)
	if systemPrompt != "" {
		text := systemPrompt
		openaiMsgs = append(openaiMsgs, openaiMessage{Role: RoleSystem, Content: &text})
	}
	for _, msg := range messages {
		text, toolUseID := extractContent(msg)

		if toolUseID != "" {
			openaiMsgs = append(openaiMsgs, openaiMessage{
				Role:       RoleTool,
				Content:    &text,
				ToolCallID: toolUseID,
			})
		} else {
			openaiMsgs = append(openaiMsgs, openaiMessage{Role: msg.Role, Content: &text})
		}
	}

	reqBody := openaiRequest{
		Model:     c.config.Model,
		Messages:  openaiMsgs,
		Tools:     convertTools(tools),
		MaxTokens: c.config.MaxTokens,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.config.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var or openaiResponse
	if err := json.Unmarshal(respBody, &or); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if or.Error != nil {
		return nil, fmt.Errorf("openai error: %s", or.Error.Message)
	}

	if len(or.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	choice := or.Choices[0]
	m := choice.Message

	content := []ContentBlock{
		{Type: ContentTypeText, Text: strPtrOrEmpty(m.Content)},
	}

	result := &ChatResult{
		Content:    content,
		Text:       strPtrOrEmpty(m.Content),
		Role:       RoleAssistant,
		StopReason: choice.FinishReason,
	}

	if len(m.ToolCalls) > 0 {
		tc := m.ToolCalls[0]
		result.ToolCall = &ToolCall{
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: json.RawMessage(tc.Function.Arguments),
		}
	}

	return result, nil
}

// extractContent 从 Message 的 content blocks 中提取文本和 tool_use_id（单次解析）
func extractContent(msg Message) (text string, toolUseID string) {
	var blocks []ContentBlock
	if err := json.Unmarshal(msg.Content, &blocks); err != nil {
		return "", ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == ContentTypeText && b.Text != "" {
			parts = append(parts, b.Text)
		}
		if b.Type == ContentTypeToolResult && b.Content != "" {
			parts = append(parts, b.Content)
		}
		if b.ToolUseID != "" {
			toolUseID = b.ToolUseID
		}
	}
	if len(parts) > 0 {
		text = parts[0]
		for _, p := range parts[1:] {
			text += "\n" + p
		}
	}
	return
}

func strPtrOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
