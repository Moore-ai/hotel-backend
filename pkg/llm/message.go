package llm

import "encoding/json"

// Role constants (Anthropic: only "user" and "assistant")
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Content block types (Anthropic content blocks)
const (
	ContentTypeText       = "text"
	ContentTypeToolUse    = "tool_use"
	ContentTypeToolResult = "tool_result"
	ContentTypeThinking  = "thinking"
)

// StopReason values
const (
	StopReasonEndTurn = "end_turn"
	StopReasonToolUse = "tool_use"
)

// ContentBlock 表示 Anthropic 响应中的单个内容块
type ContentBlock struct {
	Type string `json:"type"`           // text / tool_use / tool_result
	Text string `json:"text,omitempty"` // text 块的内容

	// tool_use 字段
	ID    string          `json:"id,omitempty"`    // tool_use 的 ID
	Name  string          `json:"name,omitempty"`  // 工具名
	Input json.RawMessage `json:"input,omitempty"` // 工具参数 JSON

	// tool_result 字段
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"` // tool_result 的文本内容

	// thinking（推理内容）：DeepSeek 的 Anthropic 兼容接口要求 content block 中始终存在此字段
	Thinking  string `json:"thinking"`
	Signature string `json:"signature,omitempty"`
}

// ToolCall 提取自 tool_use 内容块
type ToolCall struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// Message 表示 Anthropic Messages API 中的一条消息
// Content 始终使用 content block 数组格式
type Message struct {
	Role    string          `json:"role"`    // "user" 或 "assistant"
	Content json.RawMessage `json:"content"` // []ContentBlock 的 JSON
}

// ToolDef 定义 Anthropic 工具（对应 OpenAI 的 functions）
type ToolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"input_schema"`
}

type InputSchema struct {
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Required   []string       `json:"required,omitempty"`
}

// ChatResult 是 Chat 方法的返回值
type ChatResult struct {
	Content    []ContentBlock // 完整内容块（用于存历史）
	Text       string         // 第一个文本块内容（快捷访问）
	ToolCall   *ToolCall      // 第一个 tool_use（非空表示需要执行工具）
	Role       string         // "assistant"
	StopReason string
}

// NewUserTextMessage 创建用户文本消息
func NewUserTextMessage(text string) Message {
	blocks, _ := json.Marshal([]ContentBlock{
		{Type: ContentTypeText, Text: text},
	})
	return Message{Role: RoleUser, Content: blocks}
}

// NewAssistantMessage 创建 assistant 消息（从 ChatResult）
func NewAssistantMessage(result *ChatResult) Message {
	blocks, _ := json.Marshal(result.Content)
	return Message{Role: RoleAssistant, Content: blocks}
}

// NewToolResultMessage 创建 tool_result 消息
func NewToolResultMessage(toolUseID, content string) Message {
	blocks, _ := json.Marshal([]ContentBlock{
		{Type: ContentTypeToolResult, ToolUseID: toolUseID, Content: content},
	})
	return Message{Role: RoleUser, Content: blocks}
}
