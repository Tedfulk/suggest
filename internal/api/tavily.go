package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const TavilyAPIEndpoint = "https://api.tavily.com/search"

type TavilyClient struct {
    APIKey string
    client *http.Client
}

type TavilySearchRequest struct {
    Query             string   `json:"query"`
    APIKey            string   `json:"api_key"`
    SearchDepth      string   `json:"search_depth,omitempty"`
    MaxResults       int      `json:"max_results,omitempty"`
	Topic            string   `json:"topic,omitempty"`
	DaysBack         int      `json:"days,omitempty"`
    IncludeAnswer    bool     `json:"include_answer,omitempty"`
    IncludeImages    bool     `json:"include_images,omitempty"`
    IncludeDomains   []string `json:"include_domains,omitempty"`
    ExcludeDomains   []string `json:"exclude_domains,omitempty"`
}

type TavilySearchResponse struct {
    Query         string         `json:"query"`
    Answer        string         `json:"answer"`
    ResponseTime  float64        `json:"response_time"`
    Results       []TavilyResult `json:"results"`
}

type TavilyResult struct {
    Title   string  `json:"title"`
    URL     string  `json:"url"`
    Content string  `json:"content"`
    Score   float64 `json:"score"`
}

func NewTavilyClient(apiKey string) *TavilyClient {
    return &TavilyClient{
        APIKey: apiKey,
        client: &http.Client{},
    }
}

func (c *TavilyClient) Search(query string) (*TavilySearchResponse, error) {
    req := TavilySearchRequest{
        Query:          query,
        APIKey:         c.APIKey,
        SearchDepth:    "basic",
        MaxResults:     5,
        IncludeAnswer:  true,
    }

    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("error marshaling request: %w", err)
    }

    httpReq, err := http.NewRequest("POST", TavilyAPIEndpoint, bytes.NewBuffer(jsonData))
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

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
    }

    var result TavilySearchResponse
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, fmt.Errorf("error parsing response: %w", err)
    }

    return &result, nil
}

func (c *TavilyClient) SearchWithOptions(req TavilySearchRequest) (*TavilySearchResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", TavilyAPIEndpoint, bytes.NewBuffer(jsonData))
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result TavilySearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &result, nil
} 