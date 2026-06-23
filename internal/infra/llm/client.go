package llm

import (
	"bufio"
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

// Client 定义非流式 LLM 聊天补全能力。
type Client interface {
	CreateChatCompletion(ctx context.Context, request ChatRequest) (ChatResponse, error)
}

// StreamingClient 定义流式 LLM 聊天补全能力。
type StreamingClient interface {
	CreateChatCompletionStream(ctx context.Context, request ChatRequest, onDelta func(ChatStreamDelta) error) (ChatResponse, error)
}

// Message 表示发送给 OpenAI-compatible 接口的一条消息。
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 是 OpenAI-compatible 聊天补全请求。
type ChatRequest struct {
	Messages []Message
}

// ChatResponse 是聊天补全的聚合响应。
type ChatResponse struct {
	Content string
	Raw     string
}

// ChatStreamDelta 表示流式聊天补全的单个文本片段。
type ChatStreamDelta struct {
	Content string
	Raw     string
}

// OpenAICompatibleClient 封装 OpenAI-compatible HTTP 调用。
type OpenAICompatibleClient struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAICompatibleClient 创建并返回对应组件实例。
func NewOpenAICompatibleClient(cfg config.LLMConfig) (*OpenAICompatibleClient, error) {
	if !cfg.HasRequiredFields() {
		return nil, errors.New("LLM BaseURL、API Key 和模型不能为空")
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

// CreateChatCompletion 封装当前文件中的辅助处理逻辑。
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
		return ChatResponse{}, fmt.Errorf("LLM 请求失败，状态码 %d：%s", httpResponse.StatusCode, strings.TrimSpace(string(raw)))
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return ChatResponse{}, err
	}
	if len(parsed.Choices) == 0 {
		return ChatResponse{}, errors.New("LLM 响应缺少 choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return ChatResponse{}, errors.New("LLM 响应内容为空")
	}

	return ChatResponse{
		Content: content,
		Raw:     string(raw),
	}, nil
}

// CreateChatCompletionStream 封装当前文件中的辅助处理逻辑。
func (c *OpenAICompatibleClient) CreateChatCompletionStream(ctx context.Context, request ChatRequest, onDelta func(ChatStreamDelta) error) (ChatResponse, error) {
	if onDelta == nil {
		onDelta = func(ChatStreamDelta) error { return nil }
	}

	body, err := json.Marshal(openAIChatRequest{
		Model:       c.model,
		Messages:    request.Messages,
		Temperature: 0.7,
		Stream:      true,
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
	httpRequest.Header.Set("Accept", "text/event-stream")

	httpResponse, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return ChatResponse{}, err
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode < http.StatusOK || httpResponse.StatusCode >= http.StatusMultipleChoices {
		raw, readErr := io.ReadAll(io.LimitReader(httpResponse.Body, 1024*1024))
		if readErr != nil {
			return ChatResponse{}, readErr
		}

		return ChatResponse{}, fmt.Errorf("LLM 流式请求失败，状态码 %d：%s", httpResponse.StatusCode, strings.TrimSpace(string(raw)))
	}

	var content strings.Builder
	var raw strings.Builder
	reader := bufio.NewReader(httpResponse.Body)
	for {
		line, readErr := reader.ReadString('\n')
		if len(line) > 0 {
			done, err := handleOpenAIStreamLine(strings.TrimSpace(line), &content, &raw, onDelta)
			if err != nil {
				return ChatResponse{}, err
			}
			if done {
				break
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}

			return ChatResponse{}, readErr
		}
	}

	reply := strings.TrimSpace(content.String())
	if reply == "" {
		return ChatResponse{}, errors.New("LLM 响应内容为空")
	}

	return ChatResponse{
		Content: reply,
		Raw:     strings.TrimSpace(raw.String()),
	}, nil
}

// handleOpenAIStreamLine 解析 OpenAI-compatible 流式响应行。
func handleOpenAIStreamLine(line string, content *strings.Builder, raw *strings.Builder, onDelta func(ChatStreamDelta) error) (bool, error) {
	if line == "" || strings.HasPrefix(line, ":") {
		return false, nil
	}
	if !strings.HasPrefix(line, "data:") {
		return false, nil
	}

	data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if data == "[DONE]" {
		return true, nil
	}
	if data == "" {
		return false, nil
	}

	var chunk openAIChatStreamChunk
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return false, err
	}
	raw.WriteString(data)
	raw.WriteByte('\n')
	for _, choice := range chunk.Choices {
		if choice.Delta.Content == "" {
			continue
		}
		content.WriteString(choice.Delta.Content)
		if err := onDelta(ChatStreamDelta{
			Content: choice.Delta.Content,
			Raw:     data,
		}); err != nil {
			return false, err
		}
	}

	return false, nil
}

// openAIChatRequest 是 OpenAI-compatible 接口的内部请求结构。
type openAIChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// openAIChatResponse 是 OpenAI-compatible 接口的内部响应结构。
type openAIChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// openAIChatStreamChunk 是 OpenAI-compatible 流式响应块。
type openAIChatStreamChunk struct {
	Choices []struct {
		Delta Message `json:"delta"`
	} `json:"choices"`
}
