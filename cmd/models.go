package cmd

import (
	"fmt"
	"sort"

	"github.com/tedfulk/suggest/internal/config"

	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available models",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		// Fetch models from each configured provider
		if cfg.OpenAIAPIKey != "" {
			fmt.Println("\nOpenAI models:")
			models, err := config.FetchModels(config.ProviderOpenAI, cfg)
			if err != nil {
				fmt.Printf("Error fetching OpenAI models: %v\n", err)
			} else {
				sort.Strings(models)
				for _, model := range models {
					printModelWithAliases(model, cfg.ModelAliases)
				}
			}
		}

		if cfg.GroqAPIKey != "" {
			fmt.Println("\nGroq models:")
			models, err := config.FetchModels(config.ProviderGroq, cfg)
			if err != nil {
				fmt.Printf("Error fetching Groq models: %v\n", err)
			} else {
				sort.Strings(models)
				for _, model := range models {
					printModelWithAliases(model, cfg.ModelAliases)
				}
			}
		}

		if cfg.GeminiAPIKey != "" {
			fmt.Println("\nGemini models:")
			models, err := config.FetchModels(config.ProviderGemini, cfg)
			if err != nil {
				fmt.Printf("Error fetching Gemini models: %v\n", err)
			} else {
				sort.Strings(models)
				for _, model := range models {
					printModelWithAliases(model, cfg.ModelAliases)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(modelsCmd)
}

func printModelWithAliases(model string, aliases map[string]string) {
	modelAliases := []string{}
	for alias, m := range aliases {
		if m == model {
			modelAliases = append(modelAliases, alias)
		}
	}

	if len(modelAliases) > 0 {
		fmt.Printf("  %s (aliases: %v)\n", model, modelAliases)
	} else {
		fmt.Printf("  %s\n", model)
	}
} 