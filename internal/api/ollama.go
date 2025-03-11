package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const DefaultOllamaEndpoint = "http://localhost:11434"

type OllamaClient struct {
	Host   string
	client *http.Client
}

type OllamaRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

func NewOllamaClient(host string) *OllamaClient {
	if host == "" {
		host = DefaultOllamaEndpoint
	}
	return &OllamaClient{
		Host:   host,
		client: &http.Client{},
	}
}

func (c *OllamaClient) CreateChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	ollamaReq := OllamaRequest{
		Model:    req.Model,
		Messages: req.Messages,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", c.Host)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Accumulate the full response
	var fullMessage strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var streamResp struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			Done bool `json:"done"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &streamResp); err != nil {
			continue // Skip malformed lines
		}
		fullMessage.WriteString(streamResp.Message.Content)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %w", err)
	}

	result := &ChatCompletionResponse{
		Choices: []struct {
			Index   int         `json:"index"`
			Message ChatMessage `json:"message"`
		}{
			{
				Index: 0,
				Message: ChatMessage{
					Role:    "assistant",
					Content: fullMessage.String(),
				},
			},
		},
	}

	return result, nil
} 