package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"speakmate/internal/infra/llm"
	"speakmate/internal/model"
)

// LLMCorrectionAgent 使用 LLM 生成用户表达纠错结果。
type LLMCorrectionAgent struct {
	client   llm.Client
	fallback CorrectionAgent
}

// LLMCorrectionOption 用于配置 LLMCorrectionAgent。
type LLMCorrectionOption func(*LLMCorrectionAgent)

// NewLLMCorrectionAgent 创建并返回对应组件实例。
func NewLLMCorrectionAgent(client llm.Client, opts ...LLMCorrectionOption) *LLMCorrectionAgent {
	agent := &LLMCorrectionAgent{
		client: client,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

// WithCorrectionFallbackAgent 返回用于覆盖默认行为的配置选项。
func WithCorrectionFallbackAgent(fallback CorrectionAgent) LLMCorrectionOption {
	return func(agent *LLMCorrectionAgent) {
		agent.fallback = fallback
	}
}

// Correct 调用 LLM 生成结构化纠错结果，并在失败时按配置降级。
func (a *LLMCorrectionAgent) Correct(input CorrectionInput) (CorrectionOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(input, errors.New("LLM 客户端不能为空"))
	}

	response, err := a.client.CreateChatCompletion(context.Background(), llm.ChatRequest{
		Messages: toLLMMessages(BuildCorrectionPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(input, fmt.Errorf("创建聊天补全失败：%w", err))
	}

	result, err := parseCorrectionJSON(response.Content, input)
	if err != nil {
		return a.fallbackOrError(input, err)
	}

	return CorrectionOutput{
		Result: result,
		Raw:    response.Raw,
	}, nil
}

// fallbackOrError 在配置 fallback 时降级处理，否则返回原始错误。
func (a *LLMCorrectionAgent) fallbackOrError(input CorrectionInput, err error) (CorrectionOutput, error) {
	if a.fallback != nil {
		return a.fallback.Correct(input)
	}

	return CorrectionOutput{}, err
}

// correctionPayload 是 LLM 纠错 JSON 的内部解析结构。
type correctionPayload struct {
	MessageID         *int                      `json:"message_id"`
	SessionID         *int                      `json:"session_id,omitempty"`
	OriginalText      *string                   `json:"original_text"`
	CorrectedText     *string                   `json:"corrected_text"`
	Errors            *[]correctionErrorPayload `json:"errors"`
	BetterExpressions *[]string                 `json:"better_expressions"`
}

// correctionErrorPayload 是 LLM 纠错错误项的内部解析结构。
type correctionErrorPayload struct {
	Type        *string `json:"type"`
	Span        *string `json:"span"`
	Suggestion  *string `json:"suggestion"`
	Explanation *string `json:"explanation"`
}

// parseCorrectionJSON 严格解析 LLM 返回的纠错 JSON。
func parseCorrectionJSON(content string, input CorrectionInput) (model.CorrectionResult, error) {
	var payload correctionPayload
	if err := decodeStrictJSONObject(content, &payload); err != nil {
		return model.CorrectionResult{}, fmt.Errorf("解析纠错 JSON 失败：%w", err)
	}

	if payload.MessageID == nil || *payload.MessageID <= 0 {
		return model.CorrectionResult{}, errors.New("纠错 message_id 不能为空")
	}
	if input.UserMessage.ID > 0 && *payload.MessageID != input.UserMessage.ID {
		return model.CorrectionResult{}, fmt.Errorf("纠错 message_id = %d，期望 %d", *payload.MessageID, input.UserMessage.ID)
	}
	originalText, err := requiredTrimmedString(payload.OriginalText, "correction original_text")
	if err != nil {
		return model.CorrectionResult{}, err
	}
	correctedText, err := requiredTrimmedString(payload.CorrectedText, "correction corrected_text")
	if err != nil {
		return model.CorrectionResult{}, err
	}
	if payload.Errors == nil {
		return model.CorrectionResult{}, errors.New("纠错 errors 不能为空")
	}
	if payload.BetterExpressions == nil {
		return model.CorrectionResult{}, errors.New("纠错 better_expressions 不能为空")
	}

	correctionErrors := make([]model.CorrectionError, 0, len(*payload.Errors))
	for i, item := range *payload.Errors {
		correctionError, err := normalizeCorrectionErrorPayload(i, item)
		if err != nil {
			return model.CorrectionResult{}, err
		}
		correctionErrors = append(correctionErrors, correctionError)
	}

	betterExpressions := make([]string, 0, len(*payload.BetterExpressions))
	for i, expression := range *payload.BetterExpressions {
		expression = strings.TrimSpace(expression)
		if expression == "" {
			return model.CorrectionResult{}, fmt.Errorf("纠错 better_expressions[%d] 不能为空", i)
		}
		betterExpressions = append(betterExpressions, expression)
	}

	sessionID := sessionIDFromFeedbackInput(input.Session.ID, input.UserMessage.SessionID)
	if payload.SessionID != nil {
		if *payload.SessionID <= 0 {
			return model.CorrectionResult{}, errors.New("纠错 session_id 必须为正数")
		}
		if sessionID > 0 && *payload.SessionID != sessionID {
			return model.CorrectionResult{}, fmt.Errorf("纠错 session_id = %d，期望 %d", *payload.SessionID, sessionID)
		}
		sessionID = *payload.SessionID
	}

	return model.CorrectionResult{
		MessageID:         *payload.MessageID,
		SessionID:         sessionID,
		OriginalText:      originalText,
		CorrectedText:     correctedText,
		Errors:            correctionErrors,
		BetterExpressions: betterExpressions,
	}, nil
}

// normalizeCorrectionErrorPayload 校验并归一化单条纠错错误项。
func normalizeCorrectionErrorPayload(index int, payload correctionErrorPayload) (model.CorrectionError, error) {
	errorType, err := requiredTrimmedString(payload.Type, fmt.Sprintf("correction errors[%d].type", index))
	if err != nil {
		return model.CorrectionError{}, err
	}
	correctionErrorType := model.CorrectionErrorType(errorType)
	if !validCorrectionErrorType(correctionErrorType) {
		return model.CorrectionError{}, fmt.Errorf("纠错 errors[%d].type = %q 无效", index, errorType)
	}
	span, err := requiredTrimmedString(payload.Span, fmt.Sprintf("correction errors[%d].span", index))
	if err != nil {
		return model.CorrectionError{}, err
	}
	suggestion, err := requiredTrimmedString(payload.Suggestion, fmt.Sprintf("correction errors[%d].suggestion", index))
	if err != nil {
		return model.CorrectionError{}, err
	}
	explanation, err := requiredTrimmedString(payload.Explanation, fmt.Sprintf("correction errors[%d].explanation", index))
	if err != nil {
		return model.CorrectionError{}, err
	}

	return model.CorrectionError{
		Type:        correctionErrorType,
		Span:        span,
		Suggestion:  suggestion,
		Explanation: explanation,
	}, nil
}

// validCorrectionErrorType 判断纠错类型是否属于约定枚举。
func validCorrectionErrorType(errorType model.CorrectionErrorType) bool {
	switch errorType {
	case model.CorrectionErrorTypeGrammar,
		model.CorrectionErrorTypeVocabulary,
		model.CorrectionErrorTypeExpression,
		model.CorrectionErrorTypeStructure,
		model.CorrectionErrorTypeScenario:
		return true
	default:
		return false
	}
}

// decodeStrictJSONObject 只接受单个 JSON 对象，拒绝尾随数据。
func decodeStrictJSONObject(content string, value any) error {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("响应包含多余数据")
	}

	return nil
}

// requiredTrimmedString 校验必填字符串并去除首尾空白。
func requiredTrimmedString(value *string, field string) (string, error) {
	if value == nil {
		return "", fmt.Errorf("%s 不能为空", field)
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return "", fmt.Errorf("%s 不能为空", field)
	}

	return trimmed, nil
}

// sessionIDFromFeedbackInput 从反馈输入中提取 Session ID。
func sessionIDFromFeedbackInput(sessionID int, messageSessionID int) int {
	if messageSessionID > 0 {
		return messageSessionID
	}

	return sessionID
}
