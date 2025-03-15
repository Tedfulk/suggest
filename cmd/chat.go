package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tedfulk/suggest/internal/api"
	"github.com/tedfulk/suggest/internal/config"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)


var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with the AI model",
	Long: `Start an interactive chat session with the AI model. The conversation
will continue until you type "bye", "stop", "end", or press Ctrl+C.

Example:
  suggest chat
  suggest chat --model gpt-4
  suggest chat -m llama3.3-70b-versatile
  suggest chat -s "Programming Assistant"`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		model := cfg.Model
		if modelFlag != "" {
			model = modelFlag
		}

		if actualModel, exists := cfg.ModelAliases[model]; exists {
			model = actualModel
		}

		systemPrompt := cfg.SystemPrompt
		if systemFlag != "" {
			var found bool
			for _, p := range cfg.SystemPrompts {
				if p.Title == systemFlag {
					systemPrompt = p.Content
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("System prompt '%s' not found\n", systemFlag)
				return
			}
		}

		messages := []api.ChatMessage{}
		if systemPrompt != "" {
			messages = append(messages, api.ChatMessage{
				Role:    "system",
				Content: systemPrompt,
			})
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		fmt.Printf("\nStarting chat session with %s\n", cyan(model))
		blue := color.New(color.FgBlue).SprintFunc()
		fmt.Printf("Type %s, %s, or %s to exit the conversation\n", blue("'bye'"), blue("'stop'"), blue("'end'"))
		fmt.Printf("Press %s to force quit\n\n", blue("Ctrl+C"))

		scanner := bufio.NewScanner(os.Stdin)
		displayName := "User"
		if cfg.Username != "" {
			displayName = cfg.Username
		}

		for {
			fmt.Print(cyan(displayName + ": "))
			if !scanner.Scan() {
				break
			}

			input := strings.TrimSpace(scanner.Text())
			if input == "" {
				continue
			}

			if input == "bye" || input == "stop" || input == "end" {
				fmt.Println("\nEnding chat session. Goodbye!")
				break
			}

			messages = append(messages, api.ChatMessage{
				Role:    "user",
				Content: input,
			})

			req := &api.ChatCompletionRequest{
				Model:       model,
				Messages:    messages,
				Temperature: 0.1,
			}

			var resp *api.ChatCompletionResponse
			var apiErr error

			provider := config.DetermineModelProvider(model, cfg)
			if provider == "" {
				fmt.Printf("Model '%s' not supported. Please use a Groq, OpenAI, Gemini, or Ollama model.\n", model)
				return
			}

			switch provider {
			case "groq":
				if cfg.GroqAPIKey == "" {
					fmt.Println("Groq API key not set. Please set it in your config file.")
					return
				}
				client := api.NewGroqClient(cfg.GroqAPIKey)
				resp, apiErr = client.CreateChatCompletion(req)
			
			case "openai":
				if cfg.OpenAIAPIKey == "" {
					fmt.Println("OpenAI API key not set. Please set it in your config file.")
					return
				}
				client := api.NewOpenAIClient(cfg.OpenAIAPIKey)
				resp, apiErr = client.CreateChatCompletion(req)
			
			case "gemini":
				if cfg.GeminiAPIKey == "" {
					fmt.Println("Gemini API key not set. Please set it in your config file.")
					return
				}
				client := api.NewGeminiClient(cfg.GeminiAPIKey)
				resp, apiErr = client.CreateChatCompletion(req)
			
			case "ollama":
				client := api.NewOllamaClient(cfg.OllamaHost)
				resp, apiErr = client.CreateChatCompletion(req)
			}

			if apiErr != nil {
				fmt.Printf("Error: %v\n", apiErr)
				continue
			}

			if len(resp.Choices) > 0 {
				output := resp.Choices[0].Message.Content
				messages = append(messages, api.ChatMessage{
					Role:    "assistant",
					Content: output,
				})

				// Render markdown using Glamour
				r, _ := glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(100),
				)
				doc, err := r.Render(output)
				if err != nil {
					fmt.Printf("\n%s: %s\n\n", cyan(model), output)
				} else {
					fmt.Printf("\n%s:\n%s\n", cyan(model), doc)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(chatCmd)
} 