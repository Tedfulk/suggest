package cmd

import (
	"fmt"
	"strings"

	"github.com/tedfulk/suggest/internal/api"
	"github.com/tedfulk/suggest/internal/config"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

const enhanceProgrammingPrompt = `You are an expert in refining vague coding-related prompts. Your task is to take an input prompt and transform it into a clearer, more detailed, and structured version that improves specificity and relevance. Follow these steps:

1. **Identify missing details**: Determine what key information is lacking, such as programming language, frameworks, performance constraints, or specific goals.
2. **Enhance clarity**: Ensure the refined prompt is structured and unambiguous.
3. **Add specificity**: Include relevant details like libraries, performance considerations, or real-world use cases.
4. **Maintain original intent**: Ensure the improved prompt aligns with the users initial question.
5. **Provide an improved version**: Output a refined prompt that is more effective for generating high-quality responses.

### **Examples:**

**Input:** "How do I use generics in TypeScript?"
**Refined Output:** "What are the best practices for using generics in TypeScript to create reusable and type-safe functions, classes, and interfaces? Explain concepts such as generic constraints (extends), default generic types, and key utility types like Partial<T> and Record<K, T>. Provide real-world examples of applying generics in APIs and component-based architectures."

**Input:** "How do I manage state in JavaScript?"
**Refined Output:** "What are the best state management techniques in JavaScript for modern web applications? Compare approaches such as React Context API, Redux, Zustand, and using built-in browser storage (localStorage, sessionStorage). Discuss their use cases, performance considerations, and best practices for managing global and local state efficiently."

**Input:** "What are some system design patterns?"
**Refined Output:** "What are the most commonly used system design patterns for building scalable and resilient distributed systems? Focus on patterns such as event-driven architecture, microservices, CQRS (Command Query Responsibility Segregation), and database sharding. Discuss their use cases, advantages, and trade-offs in large-scale applications."

Now, refine the following prompt:

%s`

func enhancePrompt(prompt string, cfg *config.Config) (string, error) {
	messages := []api.ChatMessage{
		{
			Role:    "system",
			Content: fmt.Sprintf(enhanceProgrammingPrompt, prompt),
		},
	}

	req := &api.ChatCompletionRequest{
		Model:       "llama-3.3-70b-versatile",
		Messages:    messages,
		Temperature: 0.7,
		Stream:      false,
	}

	client := api.NewGroqClient(cfg.GroqAPIKey)
	resp, err := client.CreateChatCompletion(req)
	if err != nil {
		return "", fmt.Errorf("error enhancing prompt: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response received from enhancement model")
	}

	enhancedPrompt := resp.Choices[0].Message.Content
	enhancedPrompt = strings.TrimSpace(enhancedPrompt)

	if strings.Contains(enhancedPrompt, "Refined Output:") {
		parts := strings.SplitN(enhancedPrompt, "Refined Output:", 2)
		if len(parts) > 1 {
			enhancedPrompt = strings.TrimSpace(parts[1])
		}
	}

	return enhancedPrompt, nil
}

func renderWithGlamour(text string) string {
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	doc, err := r.Render(text)
	if err != nil {
		return text
	}
	return doc
}

var enhanceCmd = &cobra.Command{
	Use:   "enhance [message]",
	Short: "Enhance a coding-related prompt before processing",
	Long: `Enhance a coding-related prompt by adding more specificity, clarity, and structure.
The enhanced prompt will be processed by Groq's llama-3.3-70b-versatile model before being sent to your default model.

Example:
  suggest enhance "How do I use generics?"
  suggest enhance "What are design patterns?"`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide a message to enhance. Use --help for more information.")
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if cfg.GroqAPIKey == "" {
			fmt.Println("Groq API key not set. Please set it in your config file.")
			return
		}

		message := strings.Join(args, " ")
		enhancedPrompt, err := enhancePrompt(message, cfg)
		if err != nil {
			fmt.Printf("Error enhancing prompt: %v\n", err)
			return
		}

		fmt.Printf("\nEnhanced prompt:\n%s\n\nProcessing enhanced prompt...\n\n", 
			renderWithGlamour(enhancedPrompt))

		// Now process the enhanced prompt with the default model
		messages := []api.ChatMessage{}
		if cfg.SystemPrompt != "" {
			messages = append(messages, api.ChatMessage{
				Role:    "system",
				Content: cfg.SystemPrompt,
			})
		}

		messages = append(messages, api.ChatMessage{
			Role:    "user",
			Content: enhancedPrompt,
		})

		req := &api.ChatCompletionRequest{
			Model:       cfg.Model,
			Messages:    messages,
			Temperature: 0.7,
		}

		provider := config.DetermineModelProvider(cfg.Model, cfg)
		resp, apiErr := getResponse(provider, cfg, req)
		if apiErr != nil {
			fmt.Printf("Error: %v\n", apiErr)
			return
		}

		if len(resp.Choices) > 0 {
			output := resp.Choices[0].Message.Content
			fmt.Print(renderWithGlamour(output))
		}
	},
}

func init() {
	rootCmd.AddCommand(enhanceCmd)
} 