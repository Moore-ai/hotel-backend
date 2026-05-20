package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"hotel-backend/internal/database"
	"hotel-backend/pkg/llm"
)

type ChatContext struct {
	UserID uint
}

type ChatFunc struct {
	Definition llm.ToolDef
	Execute    func(ctx *ChatContext, args json.RawMessage) (string, error)
}

// ChatService 处理 AI 智能管家的对话编排
type ChatService struct {
	llmClient    *llm.Client
	systemPrompt string
	tools        []ChatFunc
	toolMap      map[string]ChatFunc
	toolDefsCache []llm.ToolDef
	maxHistory   int
	// 依赖的 Service
	waiterSvc *WaiterService
	orderSvc  *OrderService
	appealSvc *AppealService
}

func NewChatService(llmClient *llm.Client, systemPrompt string, maxHistory int, waiterSvc *WaiterService, orderSvc *OrderService, appealSvc *AppealService) *ChatService {
	s := &ChatService{
		llmClient:    llmClient,
		systemPrompt: systemPrompt,
		maxHistory:   maxHistory,
		waiterSvc:    waiterSvc,
		orderSvc:     orderSvc,
		appealSvc:    appealSvc,
		toolMap:      make(map[string]ChatFunc),
	}
	s.registerTools()
	return s
}

func (s *ChatService) registerTools() {
	s.tools = []ChatFunc{
		{
			Definition: llm.ToolDef{
				Name:        "dispatch_waiter",
				Description: "为客户房间安排一名服务员前往服务",
				InputSchema: llm.InputSchema{
					Type: "object",
					Properties: map[string]any{
						"room_id": map[string]any{"type": "integer", "description": "房间号"},
						"content": map[string]any{"type": "string", "description": "服务内容"},
					},
					Required: []string{"room_id", "content"},
				},
			},
			Execute: func(ctx *ChatContext, args json.RawMessage) (string, error) {
				var req struct {
					RoomID  uint   `json:"room_id"`
					Content string `json:"content"`
				}
				if err := json.Unmarshal(args, &req); err != nil {
					return "", fmt.Errorf("invalid arguments: %w", err)
				}
				waiter, err := s.waiterSvc.Dispatch(req.RoomID, ctx.UserID, req.Content, "")
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("已安排服务员 %s 前往 %d 号房处理「%s」", waiter.Name, req.RoomID, req.Content), nil
			},
		},
		{
			Definition: llm.ToolDef{
				Name:        "get_my_orders",
				Description: "查询当前用户的所有订单列表",
				InputSchema: llm.InputSchema{
					Type:       "object",
					Properties: map[string]any{},
				},
			},
			Execute: func(ctx *ChatContext, args json.RawMessage) (string, error) {
				orders, _, err := s.orderSvc.FindByUserID(ctx.UserID, 1, 20)
				if err != nil {
					return "", err
				}
				if len(orders) == 0 {
					return "您当前没有任何订单", nil
				}
				result := "您的订单如下：\n"
				for _, o := range orders {
					status := o.Status
					roomNumber := ""
					if o.Room != nil {
						roomNumber = o.Room.RoomNumber
					}
					result += fmt.Sprintf("- 订单 #%d：房间 %s，状态：%s，金额：%.2f\n", o.ID, roomNumber, status, o.TotalPrice)
				}
				return result, nil
			},
		},
		{
			Definition: llm.ToolDef{
				Name:        "get_appeal_status",
				Description: "查询当前用户对某个订单的申诉进度",
				InputSchema: llm.InputSchema{
					Type: "object",
					Properties: map[string]any{
						"order_id": map[string]any{"type": "integer", "description": "订单ID"},
					},
					Required: []string{"order_id"},
				},
			},
			Execute: func(ctx *ChatContext, args json.RawMessage) (string, error) {
				var req struct {
					OrderID uint `json:"order_id"`
				}
				if err := json.Unmarshal(args, &req); err != nil {
					return "", fmt.Errorf("invalid arguments: %w", err)
				}
				// 先验证订单属于当前用户
				order, err := s.orderSvc.FindByID(req.OrderID)
				if err != nil || order.UserID != ctx.UserID {
					return "", fmt.Errorf("未找到该订单")
				}
				// 查询申诉
				appeal, err := s.appealSvc.FindByOrderID(req.OrderID)
				if err != nil {
					return "该订单暂无申诉记录", nil
				}
				return fmt.Sprintf("订单 %s 的申诉状态：%s（提交于 %s）", order.OrderCode, appeal.Status, appeal.CreatedAt.Format("2006-01-02 15:04")), nil
			},
		},
		{
			Definition: llm.ToolDef{
				Name:        "create_appeal",
				Description: "对订单被拒绝取消的结果发起申诉",
				InputSchema: llm.InputSchema{
					Type: "object",
					Properties: map[string]any{
						"order_id": map[string]any{"type": "integer", "description": "订单ID"},
						"reason":   map[string]any{"type": "string", "description": "申诉理由"},
					},
					Required: []string{"order_id", "reason"},
				},
			},
			Execute: func(ctx *ChatContext, args json.RawMessage) (string, error) {
				var req struct {
					OrderID uint   `json:"order_id"`
					Reason  string `json:"reason"`
				}
				if err := json.Unmarshal(args, &req); err != nil {
					return "", fmt.Errorf("invalid arguments: %w", err)
				}
				appeal, err := s.appealSvc.Create(req.OrderID, ctx.UserID, req.Reason)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("申诉已提交，申诉编号 %s，目前状态：%s，请耐心等待审核", appeal.AppealCode, appeal.Status), nil
			},
		},
	}

	for _, t := range s.tools {
		s.toolMap[t.Definition.Name] = t
	}

	s.toolDefsCache = make([]llm.ToolDef, len(s.tools))
	for i, t := range s.tools {
		s.toolDefsCache[i] = t.Definition
	}
}

func (s *ChatService) toolDefs() []llm.ToolDef {
	return s.toolDefsCache
}

func (s *ChatService) saveToHistory(ctx context.Context, convKey string, msg llm.Message) {
	data, _ := json.Marshal(msg)
	database.RDB.RPush(ctx, convKey, string(data))
	database.RDB.LTrim(ctx, convKey, 0, int64(s.maxHistory-1))
	database.RDB.Expire(ctx, convKey, 24*time.Hour)
}

// newConversationID 无外部依赖的随机 ID
func newConversationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (s *ChatService) HandleMessage(userID uint, message string, conversationID string) (reply string, newConvID string, action string, err error) {
	ctx := context.Background()

	if conversationID == "" {
		conversationID = newConversationID()
		activeKey := fmt.Sprintf("chat:active:%d", userID)
		database.RDB.Set(ctx, activeKey, conversationID, 24*time.Hour)
	} else {
		activeKey := fmt.Sprintf("chat:active:%d", userID)
		actual, _ := database.RDB.Get(ctx, activeKey).Result()
		if actual != "" && actual != conversationID {
			return "", "", "", fmt.Errorf("无效的 conversation_id")
		}
	}

	convKey := fmt.Sprintf("chat:conv:%d:%s", userID, conversationID)

	// 从 Redis 取历史消息（Anthropic 格式：user/assistant 交替）
	history, _ := database.RDB.LRange(ctx, convKey, 0, int64(s.maxHistory-1)).Result()
	var messages []llm.Message
	for _, h := range history {
		var msg llm.Message
		if err := json.Unmarshal([]byte(h), &msg); err != nil {
			log.Printf("Failed to unmarshal history: %v", err)
			continue
		}
		messages = append(messages, msg)
	}

	// 追加用户消息（Anthropic content block 格式）
	userMsg := llm.NewUserTextMessage(message)
	messages = append(messages, userMsg)

	s.saveToHistory(ctx, convKey, userMsg)

	// 调用 Anthropic API，最多 3 轮 tool use
	sysPrompt := s.systemPrompt
	maxRounds := 3
	for round := 0; round < maxRounds; round++ {
		result, err := s.llmClient.Chat(sysPrompt, messages, s.toolDefs())
		if err != nil {
			return "", "", "", err
		}

		aiMsg := llm.NewAssistantMessage(result)
		s.saveToHistory(ctx, convKey, aiMsg)

		if result.ToolCall == nil {
			return result.Text, conversationID, "", nil
		}

		tc := result.ToolCall
		fn, ok := s.toolMap[tc.Name]
		if !ok {
			// 未知工具，告诉 LLM
			msg := llm.NewToolResultMessage(tc.ID, fmt.Sprintf("未知工具: %s", tc.Name))
			messages = append(messages, aiMsg, msg)
			continue
		}

		toolResult, fnErr := fn.Execute(&ChatContext{UserID: userID}, tc.Input)
		if fnErr != nil {
			toolResult = fmt.Sprintf("执行失败: %s", fnErr.Error())
		}

		trMsg := llm.NewToolResultMessage(tc.ID, toolResult)
		messages = append(messages, aiMsg, trMsg)

		s.saveToHistory(ctx, convKey, trMsg)

		// 最后一轮直接返回工具结果
		if round == maxRounds-1 {
			return toolResult, conversationID, tc.Name, nil
		}
	}

	return "请求处理超时，请重试", conversationID, "", nil
}
