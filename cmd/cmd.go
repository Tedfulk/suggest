package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tedfulk/suggest/internal/api"
	"github.com/tedfulk/suggest/internal/config"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var cmdCmd = &cobra.Command{
	Use:   "cmd [command description]",
	Short: "Get command suggestions with interactive options",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide a command description")
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		var messages []api.ChatMessage
		systemPrompt := "Return only the command to be executed as a raw string, no string delimiters wrapping it, no talking, no markdown, no fenced code blocks, what you return will be passed to subprocess.check_output() directly."
		
		messages = append(messages, api.ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})

		// Keep track of conversation history
		var conversationHistory []api.ChatMessage
		conversationHistory = append(conversationHistory, messages[0])

		for {
			currentMessage := strings.Join(args, " ")
			messages = append(messages, api.ChatMessage{
				Role:    "user",
				Content: currentMessage,
			})
			conversationHistory = append(conversationHistory, messages[len(messages)-1])

			req := &api.ChatCompletionRequest{
				Model:    cfg.Model,
				Messages: messages,
			}

			provider := config.DetermineModelProvider(cfg.Model, cfg)
			resp, apiErr := getResponse(provider, cfg, req)
			if apiErr != nil {
				fmt.Printf("Error: %v\n", apiErr)
				return
			}

			command := resp.Choices[0].Message.Content
			conversationHistory = append(conversationHistory, api.ChatMessage{
				Role:    "assistant",
				Content: command,
			})

			fmt.Printf("\nCommand: %s\n\n", command)

			prompt := promptui.Select{
				Label: "Choose an action",
				Items: []string{"Run", "Explain", "Follow up", "Exit"},
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}",
					Active:   "\U0001F449 {{ . | cyan }}", // 👉
					Inactive: "  {{ . | white }}",
					Selected: "\U00002705 {{ . | green }}", // ✅
				},
			}

			_, result, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			switch result {
			case "Run":
				cmd := exec.Command("sh", "-c", command)
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("Error executing command: %v\n", err)
				} else {
					fmt.Printf("%s\n", output)
				}
				return

			case "Explain":
				messages = make([]api.ChatMessage, len(conversationHistory))
				copy(messages, conversationHistory)
				messages = append(messages, api.ChatMessage{
					Role:    "user",
					Content: "Explain this command in detail and why it's relevant to my situation",
				})
				args = []string{"Explain this command in detail and why it's relevant to my situation"}
				continue

			case "Follow up":
				fmt.Print("\nEnter your follow-up question: ")
				reader := bufio.NewReader(os.Stdin)
				followUp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Printf("Error reading input: %v\n", err)
					return
				}
				followUp = strings.TrimSpace(followUp)
				
				messages = make([]api.ChatMessage, len(conversationHistory))
				copy(messages, conversationHistory)
				args = strings.Fields(followUp)
				continue

			case "Exit":
				return
			}
		}
	},
}

// Helper function to get response from the appropriate API
func getResponse(provider string, cfg *config.Config, req *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error) {
	switch provider {
	case "groq":
		if cfg.GroqAPIKey == "" {
			return nil, fmt.Errorf("groq API key not set")
		}
		return api.NewGroqClient(cfg.GroqAPIKey).CreateChatCompletion(req)
	case "openai":
		if cfg.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("openai API key not set")
		}
		return api.NewOpenAIClient(cfg.OpenAIAPIKey).CreateChatCompletion(req)
	case "gemini":
		if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("gemini API key not set")
		}
		return api.NewGeminiClient(cfg.GeminiAPIKey).CreateChatCompletion(req)
	case "ollama":
		return api.NewOllamaClient(cfg.OllamaHost).CreateChatCompletion(req)
	default:
		return nil, fmt.Errorf("model '%s' not supported", cfg.Model)
	}
}

func init() {
	rootCmd.AddCommand(cmdCmd)
} 