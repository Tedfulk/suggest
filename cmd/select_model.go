package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/tedfulk/suggest/internal/config"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var selectModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Interactively select a model to use",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		// First, select the provider
		providers := []string{"OpenAI", "Groq", "Gemini", "Ollama", "Exit"}

		providerPrompt := promptui.Select{
			Label: "Select Provider",
			Items: providers,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "\U0001F449 {{ . | cyan }}", 
				Inactive: "  {{ . | white }}",
				Selected: "\U00002705 {{ . | green }}", 
			},
		}

		_, provider, err := providerPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		if provider == "Exit" {
			return
		}

		// Check if API key is set for the selected provider
		var apiKey string
		var providerType config.Provider
		switch provider {
		case "OpenAI":
			apiKey = cfg.OpenAIAPIKey
			providerType = config.ProviderOpenAI
		case "Groq":
			apiKey = cfg.GroqAPIKey
			providerType = config.ProviderGroq
		case "Gemini":
			apiKey = cfg.GeminiAPIKey
			providerType = config.ProviderGemini
		case "Ollama":
			// No API key needed for Ollama, just check the host
			if cfg.OllamaHost == "" {
				cfg.OllamaHost = "http://localhost:11434" // Set default if not configured
			}
			providerType = config.ProviderOllama
		}

		// If no API key is set and it's not Ollama, prompt the user to enter one
		if apiKey == "" && provider != "Ollama" {
			fmt.Printf("\nNo API key set for %s. Please enter your API key: ", provider)
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				apiKey = scanner.Text()
			}
			if apiKey == "" {
				fmt.Println("API key is required to list models")
				return
			}

			// Update the config with the new API key
			switch provider {
			case "OpenAI":
				cfg.OpenAIAPIKey = apiKey
			case "Groq":
				cfg.GroqAPIKey = apiKey
			case "Gemini":
				cfg.GeminiAPIKey = apiKey
			}

			err = config.SaveConfig(cfg)
			if err != nil {
				fmt.Printf("Error saving config: %v\n", err)
				return
			}
			fmt.Printf("%s API key updated\n", provider)
		}

		fmt.Printf("\nFetching %s models...\n", provider)
		models, err := config.FetchModels(providerType, cfg)
		if err != nil {
			fmt.Printf("Error fetching models: %v\n", err)
			return
		}

		if len(models) == 0 {
			if provider == "Ollama" {
				fmt.Println("No Ollama models found. Please pull some models first using 'ollama pull <model>'")
			} else {
				fmt.Printf("No models available for %s\n", provider)
			}
			return
		}

		// Add aliases that match the selected provider
		var modelList []string
		modelList = append(modelList, models...)
		
		var aliasList []string
		for alias, model := range cfg.ModelAliases {
			if config.DetermineModelProvider(model, cfg) == string(providerType) {
				aliasList = append(aliasList, fmt.Sprintf("%s (alias for %s)", alias, model))
			}
		}
		sort.Strings(aliasList)
		modelList = append(modelList, aliasList...)

		cyan := color.New(color.FgCyan).SprintFunc()
		white := color.New(color.FgWhite).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		faint := color.New(color.Faint).SprintFunc()

		funcMap := template.FuncMap{
			"hasPrefix": strings.HasPrefix,
			"cyan":     cyan,
			"white":    white,
			"green":    green,
			"faint":    faint,
		}

		modelPrompt := promptui.Select{
			Label: "Select Model",
			Items: modelList,
			Size:  20,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "\U0001F449 {{ . | cyan }}", 
				Inactive: "  {{ . | white }}",
				Selected: "\U00002705 {{ . | green }}", 
				FuncMap:  funcMap,
				Details: `
{{ "Provider:" | faint }}	{{ if hasPrefix . "gpt-" }}OpenAI{{ else if or (hasPrefix . "mixtral-") (hasPrefix . "llama-") }}Groq{{ else if hasPrefix . "gemini-" }}Gemini{{ else }}Ollama{{ end }}
{{ "Current:" | faint }}	{{ if eq . $.Model }}Yes{{ else }}No{{ end }}`,
			},
		}

		_, result, err := modelPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		model := result
		if strings.HasSuffix(result, ")") {
			parts := strings.Split(result, " (alias for ")
			if len(parts) == 2 {
				model = strings.TrimSuffix(parts[1], ")")
			}
		}

		cfg.Model = model
		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Model set to: %s\n", model)
	},
}

func init() {
	rootCmd.AddCommand(selectModelCmd)
} 