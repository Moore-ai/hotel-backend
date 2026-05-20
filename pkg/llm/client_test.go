package llm

import (
	"encoding/json"
	"testing"
)

func TestNopMode(t *testing.T) {
	cfg := Config{BaseURL: "", Model: "test-model", MaxTokens: 100}
	client := NewClient(cfg)

	result, err := client.Chat("system prompt", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != RoleAssistant {
		t.Fatalf("expected role assistant, got %s", result.Role)
	}
	if result.Text != "nop mode response" {
		t.Fatalf("expected 'nop mode response', got %s", result.Text)
	}
}

func TestNewUserTextMessage(t *testing.T) {
	msg := NewUserTextMessage("hello")
	if msg.Role != RoleUser {
		t.Fatalf("expected role user, got %s", msg.Role)
	}
	var blocks []ContentBlock
	if err := json.Unmarshal(msg.Content, &blocks); err != nil {
		t.Fatalf("failed to unmarshal content: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Type != ContentTypeText || blocks[0].Text != "hello" {
		t.Fatalf("unexpected content blocks: %+v", blocks)
	}
}

func TestNewToolResultMessage(t *testing.T) {
	msg := NewToolResultMessage("toolu_abc", "result ok")
	if msg.Role != RoleUser {
		t.Fatalf("expected role user, got %s", msg.Role)
	}
	var blocks []ContentBlock
	if err := json.Unmarshal(msg.Content, &blocks); err != nil {
		t.Fatalf("failed to unmarshal content: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Type != ContentTypeToolResult {
		t.Fatalf("unexpected content blocks: %+v", blocks)
	}
	if blocks[0].ToolUseID != "toolu_abc" || blocks[0].Content != "result ok" {
		t.Fatalf("unexpected tool result fields")
	}
}

func TestNewAssistantMessage(t *testing.T) {
	result := &ChatResult{
		Content: []ContentBlock{
			{Type: ContentTypeText, Text: "hello"},
		},
		Role: RoleAssistant,
	}
	msg := NewAssistantMessage(result)
	if msg.Role != RoleAssistant {
		t.Fatalf("expected role assistant, got %s", msg.Role)
	}
	var blocks []ContentBlock
	if err := json.Unmarshal(msg.Content, &blocks); err != nil {
		t.Fatalf("failed to unmarshal content: %v", err)
	}
	if len(blocks) != 1 || blocks[0].Type != ContentTypeText || blocks[0].Text != "hello" {
		t.Fatalf("unexpected content blocks: %+v", blocks)
	}
}
