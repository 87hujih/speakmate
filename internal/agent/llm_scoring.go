package agent

import (
	"context"
	"errors"
	"fmt"

	"speakmate/internal/infra/llm"
	"speakmate/internal/model"
)

// LLMScoringAgent 使用 LLM 生成用户表达评分结果。
type LLMScoringAgent struct {
	client   llm.Client
	fallback ScoringAgent
}

// LLMScoringOption 用于配置 LLMScoringAgent。
type LLMScoringOption func(*LLMScoringAgent)

// NewLLMScoringAgent 创建并返回对应组件实例。
func NewLLMScoringAgent(client llm.Client, opts ...LLMScoringOption) *LLMScoringAgent {
	agent := &LLMScoringAgent{
		client: client,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// WithScoringFallbackAgent 返回用于覆盖默认行为的配置选项。
func WithScoringFallbackAgent(fallback ScoringAgent) LLMScoringOption {
	return func(agent *LLMScoringAgent) {
		agent.fallback = fallback
	}
}

// Score 调用 LLM 生成结构化评分结果，并在失败时按配置降级。
func (a *LLMScoringAgent) Score(input ScoringInput) (ScoringOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(input, errors.New("LLM 客户端不能为空"))
	}

	response, err := a.client.CreateChatCompletion(context.Background(), llm.ChatRequest{
		Messages: toLLMMessages(BuildScoringPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(input, fmt.Errorf("创建聊天补全失败：%w", err))
	}

	result, err := parseScoreJSON(response.Content, input)
	if err != nil {
		return a.fallbackOrError(input, err)
	}

	return ScoringOutput{
		Result: result,
		Raw:    response.Raw,
	}, nil
}

// fallbackOrError 在配置 fallback 时降级处理，否则返回原始错误。
func (a *LLMScoringAgent) fallbackOrError(input ScoringInput, err error) (ScoringOutput, error) {
	if a.fallback != nil {
		return a.fallback.Score(input)
	}

	return ScoringOutput{}, err
}

// scorePayload 是 LLM 评分 JSON 的内部解析结构。
type scorePayload struct {
	MessageID  *int    `json:"message_id"`
	SessionID  *int    `json:"session_id,omitempty"`
	Fluency    *int    `json:"fluency"`
	Grammar    *int    `json:"grammar"`
	Expression *int    `json:"expression"`
	Vocabulary *int    `json:"vocabulary"`
	Completion *int    `json:"completion"`
	TotalScore *int    `json:"total_score"`
	Comment    *string `json:"comment"`
}

// parseScoreJSON 严格解析 LLM 返回的评分 JSON。
func parseScoreJSON(content string, input ScoringInput) (model.ScoreResult, error) {
	var payload scorePayload
	if err := decodeStrictJSONObject(content, &payload); err != nil {
		return model.ScoreResult{}, fmt.Errorf("解析评分 JSON 失败：%w", err)
	}

	expectedMessageID := messageIDFromScoringInput(input)
	if payload.MessageID == nil || *payload.MessageID <= 0 {
		return model.ScoreResult{}, errors.New("评分 message_id 不能为空")
	}
	if expectedMessageID > 0 && *payload.MessageID != expectedMessageID {
		return model.ScoreResult{}, fmt.Errorf("评分 message_id = %d，期望 %d", *payload.MessageID, expectedMessageID)
	}

	fluency, err := requiredScore(payload.Fluency, "score fluency")
	if err != nil {
		return model.ScoreResult{}, err
	}
	grammar, err := requiredScore(payload.Grammar, "score grammar")
	if err != nil {
		return model.ScoreResult{}, err
	}
	expression, err := requiredScore(payload.Expression, "score expression")
	if err != nil {
		return model.ScoreResult{}, err
	}
	vocabulary, err := requiredScore(payload.Vocabulary, "score vocabulary")
	if err != nil {
		return model.ScoreResult{}, err
	}
	completion, err := requiredScore(payload.Completion, "score completion")
	if err != nil {
		return model.ScoreResult{}, err
	}
	totalScore, err := requiredScore(payload.TotalScore, "score total_score")
	if err != nil {
		return model.ScoreResult{}, err
	}
	comment, err := requiredTrimmedString(payload.Comment, "score comment")
	if err != nil {
		return model.ScoreResult{}, err
	}

	sessionID := sessionIDFromScoringInput(input)
	if payload.SessionID != nil {
		if *payload.SessionID <= 0 {
			return model.ScoreResult{}, errors.New("评分 session_id 必须为正数")
		}
		if sessionID > 0 && *payload.SessionID != sessionID {
			return model.ScoreResult{}, fmt.Errorf("评分 session_id = %d，期望 %d", *payload.SessionID, sessionID)
		}
		sessionID = *payload.SessionID
	}

	return model.ScoreResult{
		MessageID:  *payload.MessageID,
		SessionID:  sessionID,
		Fluency:    fluency,
		Grammar:    grammar,
		Expression: expression,
		Vocabulary: vocabulary,
		Completion: completion,
		TotalScore: totalScore,
		Comment:    comment,
	}, nil
}

// requiredScore 校验必填分数字段并限制在 0 到 100。
func requiredScore(value *int, field string) (int, error) {
	if value == nil {
		return 0, fmt.Errorf("%s 不能为空", field)
	}
	if *value < 0 || *value > 100 {
		return 0, fmt.Errorf("%s = %d 超出 0 到 100 范围", field, *value)
	}

	return *value, nil
}

// messageIDFromScoringInput 从评分输入中提取用户消息 ID。
func messageIDFromScoringInput(input ScoringInput) int {
	if input.Correction.MessageID > 0 {
		return input.Correction.MessageID
	}

	return input.UserMessage.ID
}

// sessionIDFromScoringInput 从评分输入中提取 Session ID。
func sessionIDFromScoringInput(input ScoringInput) int {
	if input.Correction.SessionID > 0 {
		return input.Correction.SessionID
	}
	return sessionIDFromFeedbackInput(input.Session.ID, input.UserMessage.SessionID)
}
