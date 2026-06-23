package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"speakmate/internal/infra/llm"
	"speakmate/internal/model"
)

func TestLLMConversationAgentUsesClientAndReturnsStageAndNextGoal(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{Content: "That sounds interesting. What was your personal contribution?"},
	}
	agent := NewLLMConversationAgent(client)

	output, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{
			Code:     "interview",
			Name:     "英语面试",
			AIRole:   "technical interviewer",
			UserGoal: "explain project experience clearly",
			Stages: []model.ScenarioStage{
				{Name: "self introduction", Description: "introduce background"},
				{Name: "project experience", Description: "describe role and impact"},
			},
		},
		Session:     model.Session{TurnCount: 0},
		UserContent: "I built a robot control project.",
	})
	if err != nil {
		t.Fatalf("GenerateReply returned error: %v", err)
	}

	if output.Reply != "That sounds interesting. What was your personal contribution?" {
		t.Fatalf("Reply = %q, want fake LLM content", output.Reply)
	}
	if output.Stage != "project experience" {
		t.Fatalf("Stage = %q, want project experience", output.Stage)
	}
	if output.NextGoal == "" {
		t.Fatal("NextGoal is empty")
	}
	if len(client.requests) != 1 {
		t.Fatalf("client request count = %d, want 1", len(client.requests))
	}
	if !promptContains(client.requests[0].Messages, "technical interviewer") {
		t.Fatalf("client prompt does not contain scenario role: %#v", client.requests[0].Messages)
	}
}

func TestLLMConversationAgentStripsLeadingStageLabelFromReply(t *testing.T) {
	client := &fakeLLMClient{
		response: llm.ChatResponse{Content: "[澄清确认] Let me confirm the owner and timeline. Who will follow up?"},
	}
	agent := NewLLMConversationAgent(client)

	output, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{
			Code:   "meeting",
			AIRole: "project manager",
			Stages: []model.ScenarioStage{
				{Name: "进度同步"},
				{Name: "澄清确认"},
			},
		},
		Session:     model.Session{TurnCount: 0},
		UserContent: "Alice will own the follow-up.",
	})
	if err != nil {
		t.Fatalf("GenerateReply returned error: %v", err)
	}

	if output.Reply != "Let me confirm the owner and timeline. Who will follow up?" {
		t.Fatalf("Reply = %q, want leading stage label removed", output.Reply)
	}
}

func TestLLMConversationAgentFallsBackToMockWhenClientFails(t *testing.T) {
	client := &fakeLLMClient{err: errors.New("up实时事件流不可用")}
	agent := NewLLMConversationAgent(client, WithFallbackAgent(NewMockConversationAgent()))

	output, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{Code: "restaurant", Stages: []model.ScenarioStage{{Name: "menu"}, {Name: "preference"}}},
		Session:  model.Session{TurnCount: 0},
	})
	if err != nil {
		t.Fatalf("GenerateReply returned error: %v", err)
	}

	if output.Reply == "" {
		t.Fatal("fallback reply is empty")
	}
	if output.Stage != "preference" {
		t.Fatalf("fallback Stage = %q, want preference", output.Stage)
	}
}

func TestLLMConversationAgentStreamsDeltasAndReturnsCombinedReply(t *testing.T) {
	client := &fakeStreamingLLMClient{
		deltas: []llm.ChatStreamDelta{
			{Content: "What "},
			{Content: "was your role?"},
		},
		response: llm.ChatResponse{Content: "What was your role?", Raw: `{"stream":true}`},
	}
	agent := NewLLMConversationAgent(client)

	var deltas []string
	output, err := agent.StreamReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{
			Code:   "interview",
			AIRole: "technical interviewer",
			Stages: []model.ScenarioStage{
				{Name: "self introduction"},
				{Name: "project experience"},
			},
		},
		Session:     model.Session{TurnCount: 0},
		UserContent: "I built a robot control project.",
	}, func(delta ConversationDelta) error {
		deltas = append(deltas, delta.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamReply returned error: %v", err)
	}

	if len(client.streamRequests) != 1 {
		t.Fatalf("stream request count = %d, want 1", len(client.streamRequests))
	}
	if !promptContains(client.streamRequests[0].Messages, "technical interviewer") {
		t.Fatalf("stream prompt does not contain scenario role: %#v", client.streamRequests[0].Messages)
	}
	if len(deltas) != 2 || deltas[0] != "What " || deltas[1] != "was your role?" {
		t.Fatalf("deltas = %#v, want streamed LLM chunks", deltas)
	}
	if output.Reply != "What was your role?" {
		t.Fatalf("Reply = %q, want combined streamed reply", output.Reply)
	}
	if output.Stage != "project experience" {
		t.Fatalf("Stage = %q, want project experience", output.Stage)
	}
	if output.Raw != `{"stream":true}` {
		t.Fatalf("Raw = %q, want stream raw", output.Raw)
	}
}

func TestLLMConversationAgentStripsLeadingStageLabelFromStreamedDeltas(t *testing.T) {
	client := &fakeStreamingLLMClient{
		deltas: []llm.ChatStreamDelta{
			{Content: "[澄清确认] "},
			{Content: "Let me confirm the owner and timeline."},
		},
		response: llm.ChatResponse{Content: "[澄清确认] Let me confirm the owner and timeline."},
	}
	agent := NewLLMConversationAgent(client)

	var deltas []string
	output, err := agent.StreamReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{
			Code:   "meeting",
			AIRole: "project manager",
			Stages: []model.ScenarioStage{
				{Name: "进度同步"},
				{Name: "澄清确认"},
			},
		},
		Session:     model.Session{TurnCount: 0},
		UserContent: "Alice will own the follow-up.",
	}, func(delta ConversationDelta) error {
		deltas = append(deltas, delta.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamReply returned error: %v", err)
	}

	if got := strings.Join(deltas, ""); got != "Let me confirm the owner and timeline." {
		t.Fatalf("streamed content = %q, want leading stage label removed", got)
	}
	if output.Reply != "Let me confirm the owner and timeline." {
		t.Fatalf("Reply = %q, want leading stage label removed", output.Reply)
	}
}

func TestLLMConversationAgentStreamsFallbackWhenClientFails(t *testing.T) {
	client := &fakeStreamingLLMClient{streamErr: errors.New("up实时事件流不可用")}
	agent := NewLLMConversationAgent(client, WithFallbackAgent(NewMockConversationAgent()))

	var chunks []string
	output, err := agent.StreamReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{
			Code:   "restaurant",
			Stages: []model.ScenarioStage{{Name: "menu"}, {Name: "preference"}},
		},
		Session: model.Session{TurnCount: 0},
	}, func(delta ConversationDelta) error {
		chunks = append(chunks, delta.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("StreamReply returned error: %v", err)
	}

	if output.Reply == "" {
		t.Fatal("fallback reply is empty")
	}
	if output.Stage != "preference" {
		t.Fatalf("fallback stage = %q, want preference", output.Stage)
	}
	if strings.Join(chunks, "") != output.Reply {
		t.Fatalf("joined fallback chunks = %q, want reply %q", strings.Join(chunks, ""), output.Reply)
	}
}

func TestLLMConversationAgentReturnsStreamErrorWithoutFallback(t *testing.T) {
	client := &fakeStreamingLLMClient{streamErr: errors.New("up实时事件流不可用")}
	agent := NewLLMConversationAgent(client)

	_, err := agent.StreamReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{Code: "meeting"},
		Session:  model.Session{TurnCount: 0},
	}, func(delta ConversationDelta) error {
		return nil
	})

	if err == nil {
		t.Fatal("StreamReply error = nil, want up事件流错误")
	}
}

func TestLLMConversationAgentReturnsErrorWhenClientFailsWithoutFallback(t *testing.T) {
	client := &fakeLLMClient{err: errors.New("up实时事件流不可用")}
	agent := NewLLMConversationAgent(client)

	_, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{Code: "meeting"},
		Session:  model.Session{TurnCount: 0},
	})

	if err == nil {
		t.Fatal("GenerateReply error = nil, want error")
	}
}

func TestLLMConversationAgentRejectsEmptyLLMReply(t *testing.T) {
	client := &fakeLLMClient{response: llm.ChatResponse{Content: "   "}}
	agent := NewLLMConversationAgent(client)

	_, err := agent.GenerateReply(context.Background(), ConversationInput{
		Scenario: model.Scenario{Code: "meeting"},
		Session:  model.Session{TurnCount: 0},
	})

	if err == nil {
		t.Fatal("GenerateReply error = nil, want error")
	}
}

type fakeLLMClient struct {
	response llm.ChatResponse
	err      error
	requests []llm.ChatRequest
}

func (c *fakeLLMClient) CreateChatCompletion(ctx context.Context, request llm.ChatRequest) (llm.ChatResponse, error) {
	c.requests = append(c.requests, request)
	if c.err != nil {
		return llm.ChatResponse{}, c.err
	}

	return c.response, nil
}

type fakeStreamingLLMClient struct {
	fakeLLMClient
	deltas         []llm.ChatStreamDelta
	streamErr      error
	response       llm.ChatResponse
	streamRequests []llm.ChatRequest
}

func (c *fakeStreamingLLMClient) CreateChatCompletionStream(ctx context.Context, request llm.ChatRequest, onDelta func(llm.ChatStreamDelta) error) (llm.ChatResponse, error) {
	c.streamRequests = append(c.streamRequests, request)
	if c.streamErr != nil {
		return llm.ChatResponse{}, c.streamErr
	}
	for _, delta := range c.deltas {
		if err := onDelta(delta); err != nil {
			return llm.ChatResponse{}, err
		}
	}

	return c.response, nil
}

func promptContains(messages []llm.Message, value string) bool {
	for _, message := range messages {
		if strings.Contains(message.Content, value) {
			return true
		}
	}

	return false
}
