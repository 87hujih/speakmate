package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"speakmate/internal/config"
)

type Client interface {
	CreateChatCompletion(ctx context.Context, request ChatRequest) (ChatResponse, error)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages []Message
}

type ChatResponse struct {
	Content string
	Raw     string
}

type OpenAICompatibleClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewOpenAICompatibleClient(cfg config.LLMConfig) (*OpenAICompatibleClient, error) {
	if !cfg.HasRequiredFields() {
		return nil, errors.New("llm base url, api key, and model are required")
	}

	timeoutSeconds := cfg.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	return &OpenAICompatibleClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}, nil
}

func (c *OpenAICompatibleClient) CreateChatCompletion(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	body, err := json.Marshal(openAIChatRequest{
		Model:       c.model,
		Messages:    request.Messages,
		Temperature: 0.7,
	})
	if err != nil {
		return ChatResponse{}, err
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/chat/completions",
		bytes.NewReader(body),
	)
	if err != nil {
		return ChatResponse{}, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return ChatResponse{}, err
	}
	defer httpResponse.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(httpResponse.Body, 1024*1024))
	if err != nil {
		return ChatResponse{}, err
	}
	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		return ChatResponse{}, fmt.Errorf("llm request failed with status %d: %s", httpResponse.StatusCode, strings.TrimSpace(string(raw)))
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ChatResponse{}, err
	}
	if len(parsed.Choices) == 0 {
		return ChatResponse{}, errors.New("llm response has no choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return ChatResponse{}, errors.New("llm response content is empty")
	}

	return ChatResponse{
		Content: content,
		Raw:     string(raw),
	}, nil
}

type openAIChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}
