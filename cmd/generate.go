package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"suggest/internal/config"

	"github.com/spf13/cobra"
)

var generateConfigCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a template configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error finding home directory:", err)
			return
		}

		configPath := filepath.Join(homeDir, ".suggest", "config.yaml")
		err = os.MkdirAll(filepath.Dir(configPath), os.ModePerm)
		if err != nil {
			fmt.Println("Error creating config directory:", err)
			return
		}

		defaultPrompts := []config.SystemPrompt{
			{
				Title:   "Helpful Assistant",
				Content: "You are a helpful assistant. You provide clear, accurate, and concise responses. When discussing code, you use markdown formatting and include helpful comments.",
			},
			{
				Title:   "Programming Assistant",
				Content: "You are a programming assistant. You help write, explain, and debug code.",
			},
			{
				Title:   "Technical Writer",
				Content: "You are a technical writer. You help create clear documentation and explanations.",
			},
		}

		defaultTemplates := []config.Template{
			{
				Title:   "Code Function",
				Content: "Write a [language] function that [task]",
			},
			{
				Title:   "Code Review",
				Content: "Review this [language] code:\n[code]",
			},
		}

		cfg := config.Config{
			OpenAIAPIKey:   "your-openai-api-key",
			GroqAPIKey:     "your-groq-api-key",
			SystemPrompt:   defaultPrompts[0].Content,
			SystemPrompts:  defaultPrompts,
			Model:         "llama-3.1-70b-versatile",
			Models: config.ModelsConfig{
				OpenAI: []string{},
				Groq:   []string{},
			},
			ModelAliases: make(map[string]string),
			Templates:    defaultTemplates,
		}

		err = config.SaveConfig(&cfg)
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Println("Configuration file generated at", configPath)
		fmt.Println("\nDefault system prompt set to:")
		fmt.Printf("  %s\n", defaultPrompts[0].Content)
		fmt.Println("\nUse 'suggest system list' to see all available system prompts")
		fmt.Println("Use 'suggest system set \"your prompt\"' to change the active system prompt")
	},
}

func init() {
	rootCmd.AddCommand(generateConfigCmd)
}
