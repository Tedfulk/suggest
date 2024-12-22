package cmd

import (
	"fmt"
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
)

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
	Short: "Chat with AI models using Groq or OpenAI",
	Long: `A CLI tool for interacting with various AI models through Groq and OpenAI APIs.
Simply type your message after 'suggest' to start chatting.

Example:
  suggest Tell me a joke about programming
  suggest --model gpt-4 What is the meaning of life?
  suggest -m mixtral-8x7b-32768 Tell me a story
  suggest -t "Code Function" --vars "language=Python,task=sort a list"
  suggest -s "Programming Assistant" Write a function`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide a message. Use --help for more information.")
			return
		}

		message := strings.Join(args, " ")
		
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
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
		
		default:
			fmt.Printf("Model '%s' not supported. Please use a Groq, OpenAI, or Gemini model.\n", model)
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

	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetUsageTemplate(usageTemplate)

	for _, cmd := range rootCmd.Commands() {
		cmd.SetHelpTemplate(helpTemplate)
		cmd.SetUsageTemplate(usageTemplate)
	}
}


