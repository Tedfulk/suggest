package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	GeminiAPIEndpoint     = "https://generativelanguage.googleapis.com/v1beta/models"
	GeminiChatEndpoint    = "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent"
)

type GeminiClient struct {
	APIKey string
	client *http.Client
}

type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Role  string       `json:"role"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		APIKey:  apiKey,
		client:  &http.Client{},
	}
}

func (c *GeminiClient) CreateChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Convert from generic request to Gemini-specific format
	geminiReq := GeminiRequest{
		Contents: make([]GeminiContent, len(req.Messages)),
	}

	// Map roles from OpenAI format to Gemini format
	roleMap := map[string]string{
		"system":    "user",      // Gemini doesn't have system, use user
		"user":      "user",
		"assistant": "model",
	}

	for i, msg := range req.Messages {
		role := roleMap[msg.Role]
		if role == "" {
			role = "user" // default to user if unknown role
		}

		geminiReq.Contents[i] = GeminiContent{
			Role:  role,
			Parts: []GeminiPart{{Text: msg.Content}},
		}
	}

	jsonData, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", GeminiChatEndpoint, c.APIKey)
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}


	// Update the response structure to match Gemini's format
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Status  string `json:"status"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w, body: %s", err, string(body))
	}

	// Check for API error
	if geminiResp.Error.Message != "" {
		return nil, fmt.Errorf("API error: %s", geminiResp.Error.Message)
	}

	result := &ChatCompletionResponse{
		Choices: []struct {
			Index   int         `json:"index"`
			Message ChatMessage `json:"message"`
		}{},
	}

	for _, candidate := range geminiResp.Candidates {
		if len(candidate.Content.Parts) > 0 {
			result.Choices = append(result.Choices, struct {
				Index   int         `json:"index"`
				Message ChatMessage `json:"message"`
			}{
				Index: len(result.Choices),
				Message: ChatMessage{
					Role:    "assistant",
					Content: candidate.Content.Parts[0].Text,
				},
			})
		}
	}

	return result, nil
}
