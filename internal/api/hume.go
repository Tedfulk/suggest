package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const HumeTTSEndpoint = "https://api.hume.ai/v0/tts"

type HumeClient struct {
	APIKey string
	client *http.Client
}

type HumeTTSRequest struct {
	Utterances []HumeUtterance `json:"utterances"`
	Context    *HumeContext    `json:"context,omitempty"`
	Format     HumeFormat      `json:"format"`
	NumGenerations int         `json:"num_generations"`
}

type HumeUtterance struct {
	Text        string `json:"text"`
	Description string `json:"description"`
}

type HumeContext struct {
	Utterances []HumeUtterance `json:"utterances"`
}

type HumeFormat struct {
	Type string `json:"type"`
}

type HumeTTSResponse struct {
	RequestID   string           `json:"request_id"`
	Generations []HumeGeneration `json:"generations"`
}

type HumeGeneration struct {
	GenerationID string         `json:"generation_id"`
	Duration     float64        `json:"duration"`
	FileSize     int            `json:"file_size"`
	Encoding     HumeEncoding   `json:"encoding"`
	Audio        string         `json:"audio"`
}

type HumeEncoding struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sample_rate"`
}

func NewHumeClient(apiKey string) *HumeClient {
	return &HumeClient{
		APIKey:  apiKey,
		client:  &http.Client{},
	}
}

// CreateTTS generates speech from text using Hume TTS
func (c *HumeClient) CreateTTS(text, voiceDescription string) ([]byte, error) {
	// Create the request payload
	ttsReq := HumeTTSRequest{
		Utterances: []HumeUtterance{
			{
				Text:        text,
				Description: voiceDescription,
			},
		},
		Format: HumeFormat{
			Type: "mp3",
		},
		NumGenerations: 1,
	}

	jsonData, err := json.Marshal(ttsReq)
	if err != nil {
		return nil, fmt.Errorf("error marshaling TTS request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", HumeTTSEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating TTS request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Hume-Api-Key", c.APIKey)

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

	// Parse the response to get the audio data
	var ttsResp HumeTTSResponse
	if err := json.Unmarshal(body, &ttsResp); err != nil {
		return nil, fmt.Errorf("error parsing TTS response: %w", err)
	}

	if len(ttsResp.Generations) == 0 {
		return nil, fmt.Errorf("no audio generation returned from Hume TTS")
	}

	// The audio field contains base64 encoded audio data
	audioData := ttsResp.Generations[0].Audio
	if audioData == "" {
		return nil, fmt.Errorf("no audio data in response")
	}

	// Decode base64 audio data
	decodedAudio, err := base64.StdEncoding.DecodeString(audioData)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 audio data: %w", err)
	}

	return decodedAudio, nil
} 