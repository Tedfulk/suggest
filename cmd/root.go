package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/tedfulk/suggest/internal/api"
	"github.com/tedfulk/suggest/internal/config"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	modelFlag    string
	templateFlag string
	systemFlag   string
	enhanceFlag  bool
)

func getLatestTag() string {
	cmd := exec.Command("git", "ls-remote", "--tags", "--sort=-v:refname", "https://github.com/tedfulk/suggest.git")
	out, err := cmd.Output()
	if err != nil {
		return "dev"
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return "dev"
	}

	tag := strings.Fields(lines[0])[1]
	tag = strings.TrimPrefix(tag, "refs/tags/")
	tag = strings.TrimPrefix(tag, "v")
	tag = strings.TrimSuffix(tag, "^{}")

	return tag
}

var version = getLatestTag()  // This will be the latest git tag or "dev" if no tags exist

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	white  = color.New(color.FgWhite).SprintFunc()
)

var helpTemplate = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}
Current Configuration:
  Username: ` + cyan("{{Username}}") + `
  Version: ` + cyan("{{Version}}") + `
  Model: ` + cyan("{{Model}}") + `
  System Prompt: ` + green("{{SystemPrompt | wrap}}") + `

{{.UsageString}}{{end}}`

var usageTemplate = `Usage:{{if .Runnable}}
  ` + cyan("{{.UseLine}}") + `{{end}}{{if .HasAvailableSubCommands}}
  ` + cyan("{{.CommandPath}}") + ` ` + yellow("[command]") + `{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  ` + green("{{rpad .Name .NamePadding }}") + ` {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
` + blue("{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}") + `{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
` + blue("{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}") + `{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} ` + yellow("[command]") + ` ` + blue("--help") + `" for more information about a command.{{end}}
`

var rootCmd = &cobra.Command{
	Use:   "suggest [message]",
	Short: "Chat with AI models using Groq, OpenAI, Gemini, or Ollama",
	Long: `A CLI tool for interacting with various AI models through Groq, OpenAI, Gemini, and Ollama APIs.
Simply type your message after 'suggest' to start chatting or pipe content into it.

Example:
  suggest Tell me a joke about programming
  suggest --model gpt-4 What is the meaning of life?
  suggest -m llama3.3-70b-versatile Tell me a story
  suggest -t "Code Function" --vars "language=Python,task=sort a list"
  suggest -s "Programming Assistant" Write a function
  suggest -e "What are design patterns?"
  suggest chat  # Start an interactive chat session
  cat file.txt | suggest "Summarize this file"
  cat code.py | suggest -s "Programming Assistant" "Review this Python code"`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		var message string
		var pipedContent string
		var argMessage string

		// Check if data is being piped
		stat, _ := os.Stdin.Stat()
		isPiped := (stat.Mode() & os.ModeCharDevice) == 0

		if isPiped {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Println("Error reading from stdin:", err)
				return
			}
			pipedContent = strings.TrimSpace(string(bytes))
		}

		if len(args) > 0 {
			argMessage = strings.Join(args, " ")
		}

		// Determine the final message based on input sources
		if pipedContent != "" && argMessage != "" {
			// Combine piped content and arguments
			message = fmt.Sprintf("Context provided via pipe:\n---\n%s\n---\n\nUser query based on arguments:\n%s", pipedContent, argMessage)
		} else if pipedContent != "" {
			// Use only piped content
			message = pipedContent
		} else if argMessage != "" {
			// Use only arguments
			message = argMessage
		} else {
			// No input provided
			fmt.Println("Please provide a message via arguments or pipe content. Use --help for more information.")
			return
		}

		// If message is effectively empty after processing, exit.
		if strings.TrimSpace(message) == "" {
			fmt.Println("Received empty or whitespace-only input.")
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if enhanceFlag {
			if cfg.GroqAPIKey == "" {
				fmt.Println("Groq API key not set. Please set it in your config file.")
				return
			}

			enhancedPrompt, err := enhancePrompt(message, cfg)
			if err != nil {
				fmt.Printf("Error enhancing prompt: %v\n", err)
				return
			}

			// Use Glamour to render the enhanced prompt
			r, _ := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(100),
			)
			renderedPrompt, err := r.Render(enhancedPrompt)
			if err != nil {
				fmt.Printf("\nEnhanced prompt:\n%s\n\nProcessing enhanced prompt...\n\n", enhancedPrompt)
			} else {
				fmt.Printf("\nEnhanced prompt:\n%s\nProcessing enhanced prompt...\n\n", renderedPrompt)
			}
			message = enhancedPrompt
		}

		if templateFlag != "" {
			var selectedTemplate *config.Template
			for _, t := range cfg.Templates {
				if t.Title == templateFlag {
					selectedTemplate = &t
					break
				}
			}
			if selectedTemplate == nil {
				fmt.Printf("Template '%s' not found\n", templateFlag)
				return
			}
			message = selectedTemplate.Content
			if vars, _ := cmd.Flags().GetString("vars"); vars != "" {
				varMap := parseTemplateVars(vars)
				for key, value := range varMap {
					message = strings.ReplaceAll(message, "["+key+"]", value)
				}
			}
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

		model := cfg.Model
		if modelFlag != "" {
			model = modelFlag
		}

		if actualModel, exists := cfg.ModelAliases[model]; exists {
			model = actualModel
		}

		messages := []api.ChatMessage{}
		
		if systemPrompt != "" {
			messages = append(messages, api.ChatMessage{
				Role:    "system",
				Content: systemPrompt,
			})
		}
		
		messages = append(messages, api.ChatMessage{
			Role:    "user",
			Content: message,
		})

		req := &api.ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: 0.7,
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
		
		default:
			fmt.Printf("Model '%s' not supported. Please use a Groq, OpenAI, Gemini, or Ollama model.\n", model)
			return
		}

		if apiErr != nil {
			fmt.Printf("Error: %v\n", apiErr)
			return
		}

		if len(resp.Choices) > 0 {
			output := resp.Choices[0].Message.Content
			
			// Render markdown using Glamour
			r, _ := glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(100),
			)
			doc, err := r.Render(output)
			if err != nil {
				fmt.Printf("Error rendering markdown: %v\n", err)
				fmt.Println(output)
				return
			}
			fmt.Print(doc)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&modelFlag, "model", "m", "", "Specify the model to use")
	rootCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "Use a template (format: template-name)")
	rootCmd.Flags().StringVarP(&systemFlag, "system", "s", "", "Use a specific system prompt by title")
	rootCmd.Flags().BoolVarP(&enhanceFlag, "enhance", "e", false, "Enhance the prompt before processing")
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")

	cobra.AddTemplateFunc("cyan", cyan)
	cobra.AddTemplateFunc("yellow", yellow)
	cobra.AddTemplateFunc("blue", blue)
	cobra.AddTemplateFunc("green", green)

	// Define wrap function first
	wrap := func(s string) string {
		if len(s) > 60 {
			return s[:57] + "..."
		}
		return s
	}

	cfg, _ := config.LoadConfig()
	cobra.AddTemplateFunc("Model", func() string {
		return cfg.Model
	})
	cobra.AddTemplateFunc("SystemPrompt", func() string {
		return wrap(cfg.SystemPrompt)
	})
	cobra.AddTemplateFunc("wrap", wrap)
	cobra.AddTemplateFunc("Version", func() string {
		return version
	})
	cobra.AddTemplateFunc("Username", func() string {
		if cfg.Username == "" {
			return "User"
		}
		return cfg.Username
	})

	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	for _, cmd := range rootCmd.Commands() {
		cmd.SetHelpTemplate(helpTemplate)
		cmd.SetUsageTemplate(usageTemplate)
	}

	// Add version flag handling
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Printf("suggest version %s\n", version)
			os.Exit(0)
		}
	}
}


