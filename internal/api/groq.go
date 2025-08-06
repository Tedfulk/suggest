package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const GroqAPIEndpoint = "https://api.groq.com/openai/v1/chat/completions"
const GroqTTSEndpoint = "https://api.groq.com/openai/v1/audio/speech"

type GroqClient struct {
	APIKey string
	client *http.Client
}

// GroqTTSRequest represents the request structure for Groq TTS
type GroqTTSRequest struct {
	Model          string `json:"model"`
	Input          string `json:"input"`
	Voice          string `json:"voice"`
	ResponseFormat string `json:"response_format,omitempty"`
}

func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{
		APIKey:  apiKey,
		client:  &http.Client{},
	}
}

func (c *GroqClient) CreateChatCompletion(req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", GroqAPIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &result, nil
}

// CreateTTS generates speech from text using Groq TTS
func (c *GroqClient) CreateTTS(text, voiceName string) ([]byte, error) {
	// Create TTS request
	ttsReq := GroqTTSRequest{
		Model:          "playai-tts",
		Input:          text,
		Voice:          voiceName,
		ResponseFormat: "wav",
	}

	jsonData, err := json.Marshal(ttsReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling TTS request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", GroqTTSEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating TTS request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making TTS request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading TTS response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TTS API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// The response is directly the audio data (WAV format)
	return body, nil
}
