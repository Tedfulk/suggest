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
	SystemPrompt   string            `yaml:"system_prompt"`
	SystemPrompts  []SystemPrompt    `yaml:"system_prompts"`
	Model          string            `yaml:"model"`
	Models         ModelsConfig      `yaml:"models"`
	ModelAliases   map[string]string `yaml:"model_aliases"`
	Templates      []Template        `yaml:"templates"`
}

type ModelsConfig struct {
	OpenAI []string `yaml:"openai"`
	Groq   []string `yaml:"groq"`
	Gemini []string `yaml:"gemini"`
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

	if cfg.Models.OpenAI == nil {
		cfg.Models.OpenAI = []string{}
	}
	if cfg.Models.Groq == nil {
		cfg.Models.Groq = []string{}
	}
	if cfg.Models.Gemini == nil {
		cfg.Models.Gemini = []string{}
	}

	if cfg.Templates == nil {
		cfg.Templates = []Template{}
	}

	if oldTemplates, ok := interface{}(cfg.Templates).(map[string]string); ok {
		var newTemplates []Template
		for name, content := range oldTemplates {
			newTemplates = append(newTemplates, Template{
				Title:   name,
				Content: content,
			})
		}
		cfg.Templates = newTemplates
	}
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
				Models: ModelsConfig{
					OpenAI: []string{},
					Groq:   []string{},
					Gemini: []string{},
				},
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

func UpdateModels(config *Config, provider Provider) error {
	switch provider {
	case ProviderOpenAI:
		if config.OpenAIAPIKey != "" {
			openAIModels, err := fetchModels("https://api.openai.com/v1/models", config.OpenAIAPIKey)
			if err != nil {
				return fmt.Errorf("error fetching OpenAI models: %w", err)
			}
			config.Models.OpenAI = openAIModels
		}
	case ProviderGroq:
		if config.GroqAPIKey != "" {
			groqModels, err := fetchModels("https://api.groq.com/openai/v1/models", config.GroqAPIKey)
			if err != nil {
				return fmt.Errorf("error fetching Groq models: %w", err)
			}
			config.Models.Groq = groqModels
		}
	case ProviderGemini:
		if config.GeminiAPIKey != "" {
			url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", config.GeminiAPIKey)
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("error fetching Gemini models: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to fetch Gemini models: %s", string(body))
			}

			var geminiResp struct {
				Models []struct {
					Name                      string   `json:"name"`
					SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
				} `json:"models"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
				return fmt.Errorf("error parsing Gemini models: %w", err)
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
			config.Models.Gemini = models
		}
	case ProviderAll:
		if config.OpenAIAPIKey != "" {
			openAIModels, err := fetchModels("https://api.openai.com/v1/models", config.OpenAIAPIKey)
			if err != nil {
				return fmt.Errorf("error fetching OpenAI models: %w", err)
			}
			config.Models.OpenAI = openAIModels
		}
		if config.GroqAPIKey != "" {
			groqModels, err := fetchModels("https://api.groq.com/openai/v1/models", config.GroqAPIKey)
			if err != nil {
				return fmt.Errorf("error fetching Groq models: %w", err)
			}
			config.Models.Groq = groqModels
		}
		if config.GeminiAPIKey != "" {
			url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", config.GeminiAPIKey)
			resp, err := http.Get(url)
			if err != nil {
				return fmt.Errorf("error fetching Gemini models: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("failed to fetch Gemini models: %s", string(body))
			}

			var geminiResp struct {
				Models []struct {
					Name                      string   `json:"name"`
					SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
				} `json:"models"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
				return fmt.Errorf("error parsing Gemini models: %w", err)
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
			config.Models.Gemini = models
		}
	}

	return SaveConfig(config)
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

func DetermineModelProvider(model string, config *Config) string {
	for _, m := range config.Models.OpenAI {
		if m == model {
			return "openai"
		}
	}
	for _, m := range config.Models.Groq {
		if m == model {
			return "groq"
		}
	}
	for _, m := range config.Models.Gemini {
		if m == model {
			return "gemini"
		}
	}

	if strings.HasPrefix(model, "gpt-") {
		return "openai"
	}
	if strings.HasPrefix(model, "mixtral-") || strings.HasPrefix(model, "llama-") {
		return "groq"
	}
	if strings.HasPrefix(model, "gemini-") {
		return "gemini"
	}

	return ""
}
