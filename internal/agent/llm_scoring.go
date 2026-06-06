package agent

import (
	"context"
	"errors"
	"fmt"

	"speakmate/internal/infra/llm"
	"speakmate/internal/model"
)

type LLMScoringAgent struct {
	client   llm.Client
	fallback ScoringAgent
}

type LLMScoringOption func(*LLMScoringAgent)

func NewLLMScoringAgent(client llm.Client, opts ...LLMScoringOption) *LLMScoringAgent {
	agent := &LLMScoringAgent{
		client: client,
	}
	for _, opt := range opts {
		opt(agent)
	}

	return agent
}

func WithScoringFallbackAgent(fallback ScoringAgent) LLMScoringOption {
	return func(agent *LLMScoringAgent) {
		agent.fallback = fallback
	}
}

func (a *LLMScoringAgent) Score(input ScoringInput) (ScoringOutput, error) {
	if a.client == nil {
		return a.fallbackOrError(input, errors.New("llm client is nil"))
	}

	response, err := a.client.CreateChatCompletion(context.Background(), llm.ChatRequest{
		Messages: toLLMMessages(BuildScoringPrompt(input)),
	})
	if err != nil {
		return a.fallbackOrError(input, fmt.Errorf("create chat completion: %w", err))
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

func (a *LLMScoringAgent) fallbackOrError(input ScoringInput, err error) (ScoringOutput, error) {
	if a.fallback != nil {
		return a.fallback.Score(input)
	}

	return ScoringOutput{}, err
}

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

func parseScoreJSON(content string, input ScoringInput) (model.ScoreResult, error) {
	var payload scorePayload
	if err := decodeStrictJSONObject(content, &payload); err != nil {
		return model.ScoreResult{}, fmt.Errorf("parse score json: %w", err)
	}

	expectedMessageID := messageIDFromScoringInput(input)
	if payload.MessageID == nil || *payload.MessageID <= 0 {
		return model.ScoreResult{}, errors.New("score message_id is required")
	}
	if expectedMessageID > 0 && *payload.MessageID != expectedMessageID {
		return model.ScoreResult{}, fmt.Errorf("score message_id = %d, want %d", *payload.MessageID, expectedMessageID)
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
			return model.ScoreResult{}, errors.New("score session_id must be positive")
		}
		if sessionID > 0 && *payload.SessionID != sessionID {
			return model.ScoreResult{}, fmt.Errorf("score session_id = %d, want %d", *payload.SessionID, sessionID)
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

func requiredScore(value *int, field string) (int, error) {
	if value == nil {
		return 0, errors.New(field + " is required")
	}
	if *value < 0 || *value > 100 {
		return 0, fmt.Errorf("%s = %d is outside 0 to 100", field, *value)
	}

	return *value, nil
}

func messageIDFromScoringInput(input ScoringInput) int {
	if input.Correction.MessageID > 0 {
		return input.Correction.MessageID
	}

	return input.UserMessage.ID
}

func sessionIDFromScoringInput(input ScoringInput) int {
	if input.Correction.SessionID > 0 {
		return input.Correction.SessionID
	}
	return sessionIDFromFeedbackInput(input.Session.ID, input.UserMessage.SessionID)
}
