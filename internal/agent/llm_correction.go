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

type LLMCorrectionAgent struct {
	client   llm.Client
	fallback CorrectionAgent
}

type LLMCorrectionOption func(*LLMCorrectionAgent)

func NewLLMCorrectionAgent(client llm.Client, opts ...LLMCorrectionOption) *LLMCorrectionAgent {
	agent := &LLMCorrectionAgent{
		client: client,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func WithCorrectionFallbackAgent(fallback CorrectionAgent) LLMCorrectionOption {
	return func(agent *LLMCorrectionAgent) {
		agent.fallback = fallback
	}
}

func (a *LLMCorrectionAgent) Correct(input CorrectionInput) (CorrectionOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(input, errors.New("llm client is nil"))
	}

	response, err := a.client.CreateChatCompletion(context.Background(), llm.ChatRequest{
		Messages: toLLMMessages(BuildCorrectionPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(input, fmt.Errorf("create chat completion: %w", err))
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

func (a *LLMCorrectionAgent) fallbackOrError(input CorrectionInput, err error) (CorrectionOutput, error) {
	if a.fallback != nil {
		return a.fallback.Correct(input)
	}

	return CorrectionOutput{}, err
}

type correctionPayload struct {
	MessageID         *int                      `json:"message_id"`
	SessionID         *int                      `json:"session_id,omitempty"`
	OriginalText      *string                   `json:"original_text"`
	CorrectedText     *string                   `json:"corrected_text"`
	Errors            *[]correctionErrorPayload `json:"errors"`
	BetterExpressions *[]string                 `json:"better_expressions"`
}

type correctionErrorPayload struct {
	Type        *string `json:"type"`
	Span        *string `json:"span"`
	Suggestion  *string `json:"suggestion"`
	Explanation *string `json:"explanation"`
}

func parseCorrectionJSON(content string, input CorrectionInput) (model.CorrectionResult, error) {
	var payload correctionPayload
	if err := decodeStrictJSONObject(content, &payload); err != nil {
		return model.CorrectionResult{}, fmt.Errorf("parse correction json: %w", err)
	}

	if payload.MessageID == nil || *payload.MessageID <= 0 {
		return model.CorrectionResult{}, errors.New("correction message_id is required")
	}
	if input.UserMessage.ID > 0 && *payload.MessageID != input.UserMessage.ID {
		return model.CorrectionResult{}, fmt.Errorf("correction message_id = %d, want %d", *payload.MessageID, input.UserMessage.ID)
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
		return model.CorrectionResult{}, errors.New("correction errors is required")
	}
	if payload.BetterExpressions == nil {
		return model.CorrectionResult{}, errors.New("correction better_expressions is required")
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
			return model.CorrectionResult{}, fmt.Errorf("correction better_expressions[%d] is empty", i)
		}
		betterExpressions = append(betterExpressions, expression)
	}

	sessionID := sessionIDFromFeedbackInput(input.Session.ID, input.UserMessage.SessionID)
	if payload.SessionID != nil {
		if *payload.SessionID <= 0 {
			return model.CorrectionResult{}, errors.New("correction session_id must be positive")
		}
		if sessionID > 0 && *payload.SessionID != sessionID {
			return model.CorrectionResult{}, fmt.Errorf("correction session_id = %d, want %d", *payload.SessionID, sessionID)
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

func normalizeCorrectionErrorPayload(index int, payload correctionErrorPayload) (model.CorrectionError, error) {
	errorType, err := requiredTrimmedString(payload.Type, fmt.Sprintf("correction errors[%d].type", index))
	if err != nil {
		return model.CorrectionError{}, err
	}
	correctionErrorType := model.CorrectionErrorType(errorType)
	if !validCorrectionErrorType(correctionErrorType) {
		return model.CorrectionError{}, fmt.Errorf("correction errors[%d].type = %q is invalid", index, errorType)
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

func decodeStrictJSONObject(content string, value any) error {
	decoder := json.NewDecoder(strings.NewReader(strings.TrimSpace(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("response contains trailing data")
	}

	return nil
}

func requiredTrimmedString(value *string, field string) (string, error) {
	if value == nil {
		return "", errors.New(field + " is required")
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return "", errors.New(field + " is empty")
	}

	return trimmed, nil
}

func sessionIDFromFeedbackInput(sessionID int, messageSessionID int) int {
	if messageSessionID > 0 {
		return messageSessionID
	}

	return sessionID
}
