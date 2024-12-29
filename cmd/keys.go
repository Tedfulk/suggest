package cmd

import (
	"fmt"

	"github.com/tedfulk/suggest/internal/config"

	"github.com/spf13/cobra"
)

var (
	openaiKey string
	groqKey   string
)

func maskKey(key string) string {
	if key == "" {
		return "not set"
	}
	if len(key) < 8 {
		return "****"
	}
	return fmt.Sprintf("%s****%s", 
		key[:4], 
		key[len(key)-4:])
}

var keysCmd = &cobra.Command{
	Use:   "keys [provider]",
	Short: "Manage API keys for OpenAI, Groq, Gemini, and Tavily",
	Long: `Manage API keys for various AI services.
	
Example:
  suggest keys openai     - Set OpenAI API key
  suggest keys groq       - Set Groq API key
  suggest keys gemini     - Set Gemini API key
  suggest keys tavily     - Set Tavily API key
  suggest keys            - Show current keys`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if len(args) == 0 {
			fmt.Printf("OpenAI API key: %s\n", maskKey(cfg.OpenAIAPIKey))
			fmt.Printf("Groq API key: %s\n", maskKey(cfg.GroqAPIKey))
			fmt.Printf("Gemini API key: %s\n", maskKey(cfg.GeminiAPIKey))
			fmt.Printf("Tavily API key: %s\n", maskKey(cfg.TavilyAPIKey))
			return
		}

		provider := args[0]
		switch provider {
		case "openai":
			fmt.Print("OpenAI API Key: ")
			var key string
			fmt.Scanln(&key)
			if key != "" {
				cfg.OpenAIAPIKey = key
				err = config.SaveConfig(cfg)
				if err != nil {
					fmt.Println("Error saving config:", err)
					return
				}
				fmt.Println("OpenAI API key updated")
				fmt.Println("Updating available models...")
				err = config.UpdateModels(cfg, config.ProviderOpenAI)
				if err != nil {
					fmt.Println("Error updating models:", err)
					return
				}
				fmt.Println("Models list updated")
			}

		case "groq":
			fmt.Print("Groq API Key: ")
			var key string
			fmt.Scanln(&key)
			if key != "" {
				cfg.GroqAPIKey = key
				err = config.SaveConfig(cfg)
				if err != nil {
					fmt.Println("Error saving config:", err)
					return
				}
				fmt.Println("Groq API key updated")
				fmt.Println("Updating available models...")
				err = config.UpdateModels(cfg, config.ProviderGroq)
				if err != nil {
					fmt.Println("Error updating models:", err)
					return
				}
				fmt.Println("Models list updated")
			}

		case "gemini":
			fmt.Print("Gemini API Key: ")
			var key string
			fmt.Scanln(&key)
			if key != "" {
				cfg.GeminiAPIKey = key
				err = config.SaveConfig(cfg)
				if err != nil {
					fmt.Println("Error saving config:", err)
					return
				}
				fmt.Println("Gemini API key updated")
				fmt.Println("Updating available models...")
				err = config.UpdateModels(cfg, config.ProviderGemini)
				if err != nil {
					fmt.Println("Error updating models:", err)
					return
				}
				fmt.Println("Models list updated")
			}

		case "tavily":
			fmt.Print("Tavily API Key: ")
			var key string
			fmt.Scanln(&key)
			if key != "" {
				cfg.TavilyAPIKey = key
				err = config.SaveConfig(cfg)
				if err != nil {
					fmt.Println("Error saving config:", err)
					return
				}
				fmt.Println("Tavily API key updated")
			}

		default:
			fmt.Printf("Unknown provider '%s'. Use 'openai', 'groq', 'gemini', or 'tavily'\n", provider)
		}
	},
}

func init() {
	rootCmd.AddCommand(keysCmd)
} 