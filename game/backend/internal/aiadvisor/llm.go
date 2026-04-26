package aiadvisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LLMClient — интерфейс для обращения к языковым моделям.
type LLMClient interface {
	Ask(ctx context.Context, model, systemPrompt, userMessage string, maxTokens int) (string, error)
}

// ModelID → display name.
var KnownModels = map[string]int{
	"claude-haiku-4-5-20251001": 5,
	"claude-sonnet-4-6":         20,
	"claude-opus-4-7":           80,
}

// ---- Anthropic -------------------------------------------------------

type anthropicClient struct {
	apiKey   string
	proxyURL string
	http     *http.Client
}

// NewAnthropicClient создаёт клиент для Anthropic Messages API.
// proxyURL опционален; если задан — используется вместо прямого эндпоинта.
func NewAnthropicClient(apiKey, proxyURL string) LLMClient {
	return &anthropicClient{
		apiKey:   apiKey,
		proxyURL: proxyURL,
		http:     &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *anthropicClient) Ask(ctx context.Context, model, systemPrompt, userMessage string, maxTokens int) (string, error) {
	endpoint := "https://api.anthropic.com/v1/messages"
	if c.proxyURL != "" {
		endpoint = c.proxyURL + "/v1/messages"
	}

	body, _ := json.Marshal(map[string]any{
		"model":      model,
		"max_tokens": maxTokens,
		"system":     systemPrompt,
		"messages":   []map[string]string{{"role": "user", "content": userMessage}},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("aiadvisor: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("aiadvisor: anthropic request: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("aiadvisor: anthropic status %d: %s", resp.StatusCode, data)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("aiadvisor: parse response: %w", err)
	}
	if len(result.Content) == 0 {
		return "", fmt.Errorf("aiadvisor: empty response")
	}
	return result.Content[0].Text, nil
}

// ---- Ollama ----------------------------------------------------------

type ollamaClient struct {
	baseURL string
	http    *http.Client
}

// NewOllamaClient создаёт клиент для Ollama HTTP API.
func NewOllamaClient(baseURL string) LLMClient {
	return &ollamaClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *ollamaClient) Ask(ctx context.Context, model, systemPrompt, userMessage string, maxTokens int) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":  model,
		"prompt": systemPrompt + "\n\n" + userMessage,
		"stream": false,
		"options": map[string]int{
			"num_predict": maxTokens,
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("aiadvisor: build ollama request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("aiadvisor: ollama request: %w", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("aiadvisor: ollama status %d: %s", resp.StatusCode, data)
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("aiadvisor: parse ollama response: %w", err)
	}
	return result.Response, nil
}
