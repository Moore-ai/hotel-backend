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

// ChatContext 是 tool execute 的上下文
type ChatContext struct {
	UserID uint
}

// ChatFunc 注册一个可供 LLM 调用的工具
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

// registerTools 注册所有可供 LLM 调用的工具（对应 Anthropic Tool Use）
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
				appeal, err := s.appealSvc.FindByOrderID(req.OrderID)
				if err != nil {
					return "", err
				}
				if appeal.UserID != ctx.UserID {
					return "", fmt.Errorf("无权查看此申诉")
				}
				status := appeal.Status
				return fmt.Sprintf("订单 #%d 的申诉状态：%s（提交于 %s）", req.OrderID, status, appeal.CreatedAt.Format("2006-01-02 15:04")), nil
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
				return fmt.Sprintf("申诉已提交，申诉编号 #%d，目前状态：%s，请耐心等待审核", appeal.ID, appeal.Status), nil
			},
		},
	}

	for _, t := range s.tools {
		s.toolMap[t.Definition.Name] = t
	}
}

// toolDefs 返回给 LLM 的工具定义列表
func (s *ChatService) toolDefs() []llm.ToolDef {
	defs := make([]llm.ToolDef, len(s.tools))
	for i, t := range s.tools {
		defs[i] = t.Definition
	}
	return defs
}

// newConversationID 生成一个随机对话 ID（标准库，无外部依赖）
func newConversationID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// HandleMessage 处理用户的一条消息，返回 AI 回复
func (s *ChatService) HandleMessage(userID uint, message string, conversationID string) (reply string, newConvID string, action string, err error) {
	ctx := context.Background()

	// 处理 conversation_id
	if conversationID == "" {
		conversationID = newConversationID()
	} else {
		activeKey := fmt.Sprintf("chat:active:%d", userID)
		actual, _ := database.RDB.Get(ctx, activeKey).Result()
		if actual != "" && actual != conversationID {
			return "", "", "", fmt.Errorf("无效的 conversation_id")
		}
	}

	convKey := fmt.Sprintf("chat:conv:%d:%s", userID, conversationID)
	activeKey := fmt.Sprintf("chat:active:%d", userID)

	// 设置活跃对话指针
	database.RDB.Set(ctx, activeKey, conversationID, 24*time.Hour)

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

	// 保存用户消息到 Redis
	userJson, _ := json.Marshal(userMsg)
	database.RDB.RPush(ctx, convKey, string(userJson))
	database.RDB.LTrim(ctx, convKey, 0, int64(s.maxHistory-1))
	database.RDB.Expire(ctx, convKey, 24*time.Hour)

	// 调用 Anthropic API，最多 3 轮 tool use
	sysPrompt := s.systemPrompt
	maxRounds := 3
	for round := 0; round < maxRounds; round++ {
		result, err := s.llmClient.Chat(sysPrompt, messages, s.toolDefs())
		if err != nil {
			return "", "", "", err
		}

		// 保存 assistant 回复到 Redis
		aiMsg := llm.NewAssistantMessage(result)
		aiJson, _ := json.Marshal(aiMsg)
		database.RDB.RPush(ctx, convKey, string(aiJson))
		database.RDB.LTrim(ctx, convKey, 0, int64(s.maxHistory-1))
		database.RDB.Expire(ctx, convKey, 24*time.Hour)

		// 检查是否需要执行工具
		if result.ToolCall == nil {
			return result.Text, conversationID, "", nil
		}

		// 执行 Tool Use
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

		// 将 tool_result 加入对话
		trMsg := llm.NewToolResultMessage(tc.ID, toolResult)
		messages = append(messages, aiMsg, trMsg)

		// 保存 tool_result 到 Redis
		trJson, _ := json.Marshal(trMsg)
		database.RDB.RPush(ctx, convKey, string(trJson))
		database.RDB.LTrim(ctx, convKey, 0, int64(s.maxHistory-1))
		database.RDB.Expire(ctx, convKey, 24*time.Hour)

		// 最后一轮直接返回工具结果
		if round == maxRounds-1 {
			return toolResult, conversationID, tc.Name, nil
		}
	}

	return "请求处理超时，请重试", conversationID, "", nil
}
