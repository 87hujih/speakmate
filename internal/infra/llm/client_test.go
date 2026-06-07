package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"speakmate/internal/config"
)

func TestOpenAICompatibleClientSendsChatCompletionRequest(t *testing.T) {
	var gotAuth string
	var gotRequest openAIChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"What was your role?"}}]}`))
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClient(config.LLMConfig{
		BaseURL:        server.URL + "/v1",
		APIKey:         "test-key",
		Model:          "test-model",
		TimeoutSeconds: 30,
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleClient returned error: %v", err)
	}

	response, err := client.CreateChatCompletion(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("CreateChatCompletion returned error: %v", err)
	}

	if gotAuth != "Bearer test-key" {
		t.Fatalf("Authorization = %q, want Bearer test-key", gotAuth)
	}
	if gotRequest.Model != "test-model" {
		t.Fatalf("model = %q, want test-model", gotRequest.Model)
	}
	if len(gotRequest.Messages) != 1 || gotRequest.Messages[0].Content != "hello" {
		t.Fatalf("messages = %#v, want user hello", gotRequest.Messages)
	}
	if response.Content != "What was your role?" {
		t.Fatalf("response content = %q, want What was your role?", response.Content)
	}
}

func TestOpenAICompatibleClientReturnsErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClient(config.LLMConfig{
		BaseURL:        server.URL,
		APIKey:         "test-key",
		Model:          "test-model",
		TimeoutSeconds: 30,
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleClient returned error: %v", err)
	}

	_, err = client.CreateChatCompletion(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	})

	if err == nil {
		t.Fatal("CreateChatCompletion error = nil, want error")
	}
}

func TestOpenAICompatibleClientStreamsChatCompletionDeltas(t *testing.T) {
	var gotRequest struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
		Stream   bool      `json:"stream"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q, want /v1/chat/completions", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("request body is not valid JSON: %v", err)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		for _, frame := range []string{
			`{"choices":[{"delta":{"role":"assistant"}}]}`,
			`{"choices":[{"delta":{"content":"What "}}]}`,
			`{"choices":[{"delta":{"content":"was your role?"}}]}`,
		} {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", frame)
		}
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client, err := NewOpenAICompatibleClient(config.LLMConfig{
		BaseURL:        server.URL + "/v1",
		APIKey:         "test-key",
		Model:          "test-model",
		TimeoutSeconds: 30,
	})
	if err != nil {
		t.Fatalf("NewOpenAICompatibleClient returned error: %v", err)
	}

	var deltas []string
	response, err := client.CreateChatCompletionStream(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	}, func(delta ChatStreamDelta) error {
		deltas = append(deltas, delta.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("CreateChatCompletionStream returned error: %v", err)
	}

	if !gotRequest.Stream {
		t.Fatal("stream = false, want true")
	}
	if gotRequest.Model != "test-model" {
		t.Fatalf("model = %q, want test-model", gotRequest.Model)
	}
	if len(deltas) != 2 || deltas[0] != "What " || deltas[1] != "was your role?" {
		t.Fatalf("deltas = %#v, want streamed content chunks", deltas)
	}
	if response.Content != "What was your role?" {
		t.Fatalf("response content = %q, want combined streamed content", response.Content)
	}
}

func TestOpenAICompatibleClientRequiresBaseURLAPIKeyAndModel(t *testing.T) {
	_, err := NewOpenAICompatibleClient(config.LLMConfig{TimeoutSeconds: 30})

	if err == nil {
		t.Fatal("NewOpenAICompatibleClient error = nil, want error")
	}
}
