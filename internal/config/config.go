package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type SystemPrompt struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

type Template struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

type Config struct {
	OpenAIAPIKey   string            `yaml:"openai_api_key"`
	GroqAPIKey     string            `yaml:"groq_api_key"`
	GeminiAPIKey   string            `yaml:"gemini_api_key"`
	TavilyAPIKey   string            `yaml:"tavily_api_key"`
	OllamaHost     string            `yaml:"ollama_host"`
	SystemPrompt   string            `yaml:"system_prompt"`
	SystemPrompts  []SystemPrompt    `yaml:"system_prompts"`
	Model          string            `yaml:"model"`
	ModelAliases   map[string]string `yaml:"model_aliases"`
	Templates      []Template        `yaml:"templates"`
	Username       string            `yaml:"username"`
}

type ModelResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderGroq   Provider = "groq"
	ProviderGemini Provider = "gemini"
	ProviderOllama Provider = "ollama"
	ProviderAll    Provider = "all"
)

func migrateConfig(cfg *Config) {
	if cfg.Templates == nil {
		cfg.Templates = []Template{}
	}

	if cfg.ModelAliases == nil {
			cfg.ModelAliases = make(map[string]string)
	}

	if cfg.SystemPrompts == nil {
		cfg.SystemPrompts = []SystemPrompt{}
	}

	var newPrompts []SystemPrompt
	for _, prompt := range cfg.SystemPrompts {
		if s, ok := interface{}(prompt).(string); ok {
			newPrompts = append(newPrompts, SystemPrompt{
				Title:   "Legacy Prompt",
				Content: s,
			})
		} else {
			newPrompts = append(newPrompts, prompt)
		}
	}
	cfg.SystemPrompts = newPrompts
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".suggest", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Templates:    []Template{},
				ModelAliases: make(map[string]string),
				SystemPrompts: []SystemPrompt{},
			}, nil
		}
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	migrateConfig(&cfg)

	return &cfg, nil
}

func SaveConfig(config *Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".suggest", "config.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func DetermineModelProvider(model string, config *Config) string {
	if strings.HasPrefix(model, "gpt-") {
		return "openai"
	}
	if strings.HasPrefix(model, "mixtral-") || strings.HasPrefix(model, "llama-") {
		return "groq"
	}
	if strings.HasPrefix(model, "gemini-") {
		return "gemini"
	}
	if strings.Contains(model, ":") ||
		strings.HasPrefix(model, "llama2") ||
		strings.HasPrefix(model, "codellama") ||
		strings.HasPrefix(model, "mistral") {
		return "ollama"
	}
	return ""
}

// FetchModels fetches available models from a provider
func FetchModels(provider Provider, cfg *Config) ([]string, error) {
	switch provider {
	case ProviderOpenAI:
		if cfg.OpenAIAPIKey != "" {
			return fetchModels("https://api.openai.com/v1/models", cfg.OpenAIAPIKey)
		}
	case ProviderGroq:
		if cfg.GroqAPIKey != "" {
			return fetchModels("https://api.groq.com/openai/v1/models", cfg.GroqAPIKey)
		}
	case ProviderGemini:
		if cfg.GeminiAPIKey != "" {
			url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", cfg.GeminiAPIKey)
			resp, err := http.Get(url)
			if err != nil {
				return nil, fmt.Errorf("error fetching Gemini models: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return nil, fmt.Errorf("failed to fetch Gemini models: %s", string(body))
			}

			var geminiResp struct {
				Models []struct {
					Name                      string   `json:"name"`
					SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
				} `json:"models"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
				return nil, fmt.Errorf("error parsing Gemini models: %w", err)
			}

			var models []string
			for _, model := range geminiResp.Models {
				for _, method := range model.SupportedGenerationMethods {
					if method == "generateContent" {
						name := strings.TrimPrefix(model.Name, "models/")
						models = append(models, name)
						break
					}
				}
			}
			return models, nil
		}
	case ProviderOllama:
		host := cfg.OllamaHost
		if host == "" {
			host = "http://localhost:11434"
		}
		
		resp, err := http.Get(fmt.Sprintf("%s/api/tags", host))
		if err != nil {
			return nil, fmt.Errorf("error fetching Ollama models: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to fetch Ollama models: %s", string(body))
		}

		var ollamaResp struct {
			Models []struct {
				Name string `json:"name"`
			} `json:"models"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return nil, fmt.Errorf("error parsing Ollama models: %w", err)
		}

		var models []string
		for _, model := range ollamaResp.Models {
			models = append(models, model.Name)
		}
		return models, nil
	}
	return nil, fmt.Errorf("no API key set for provider %s", provider)
}

func fetchModels(url, apiKey string) ([]string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch models: %s", resp.Status)
	}

	var modelResponse ModelResponse
	err = json.NewDecoder(resp.Body).Decode(&modelResponse)
	if err != nil {
		return nil, err
	}

	var models []string
	for _, model := range modelResponse.Data {
		models = append(models, model.ID)
	}

	return models, nil
}
