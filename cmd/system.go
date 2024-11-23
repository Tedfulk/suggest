package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"suggest/internal/config"
	"suggest/internal/utils"
	"text/template"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:   "system",
	Short: "Manage system prompts",
	Long: `Manage system prompts for AI interactions.
	
Example:
  suggest system add "title" "You are a helpful programming assistant"
  suggest system set "title"
  suggest system remove "title"
  suggest system list`,
}

var systemAddCmd = &cobra.Command{
	Use:   "add [title] [prompt]",
	Short: "Add a new system prompt",
	Long: `Add a new system prompt. Can be used in two ways:
	
1. Interactive mode:
   suggest system add

2. Direct mode:
   suggest system add "title" "prompt content"`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		var title, content string

		if len(args) == 0 {
			fmt.Print("Enter prompt title: ")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				title = scanner.Text()
			}

			fmt.Println("Enter prompt content (Shift+Enter for new line, Enter to finish):")
			var lines []string
			reader := bufio.NewReader(os.Stdin)
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading input:", err)
					return
				}

				line = strings.TrimRight(line, "\n")

				if strings.HasSuffix(line, "\r") {
					line = strings.TrimRight(line, "\r")
					lines = append(lines, line)
				} else {
					if line != "" {
						lines = append(lines, line)
					}
					break
				}
			}
			content = strings.Join(lines, "\n")
		} else if len(args) == 2 {
			title = args[0]
			content = args[1]
		} else {
			fmt.Println("Invalid number of arguments. Use either no arguments for interactive mode or provide both title and content.")
			return
		}

		if title == "" || content == "" {
			fmt.Println("Both title and content are required")
			return
		}

		for _, p := range cfg.SystemPrompts {
			if p.Title == title {
				fmt.Println("Prompt with this title already exists")
				return
			}
		}

		newPrompt := config.SystemPrompt{
			Title:   title,
			Content: content,
		}

		cfg.SystemPrompts = append(cfg.SystemPrompts, newPrompt)
		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Printf("System prompt '%s' added\n", title)
	},
}

var systemRemoveCmd = &cobra.Command{
	Use:   "remove [title]",
	Short: "Remove a system prompt",
	Long: `Remove a system prompt. Can be used in two ways:
	
1. Interactive mode:
   suggest system remove

2. Direct mode:
   suggest system remove "title"`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		var title string

		if len(args) == 0 {
			if len(cfg.SystemPrompts) == 0 {
				fmt.Println("No system prompts available to remove")
				return
			}

			cyan := color.New(color.FgCyan).SprintFunc()
			white := color.New(color.FgWhite).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			faint := color.New(color.Faint).SprintFunc()

			funcMap := template.FuncMap{
				"cyan":     cyan,
				"white":    white,
				"green":    green,
				"red":      red,
				"faint":    faint,
				"truncate": utils.TruncateText,
			}

			prompt := promptui.Select{
				Label: "Select System Prompt to Remove",
					Items: cfg.SystemPrompts,
					Size:  20,
					Templates: &promptui.SelectTemplates{
						Label:    "{{ . }}",
						Active:   "\U0001F449 {{ .Title | cyan }}", // 👉
						Inactive: "  {{ .Title | white }}",
						Selected: "\U0001F5D1 {{ .Title | red }}", // 🗑 
						Details: `
{{ "Title:" | faint }}	{{ .Title }}
{{ "Preview:" | faint }}	{{ .Content | truncate 100 | faint }}
{{ "Current:" | faint }}	{{ if eq .Content $.SystemPrompt }}Yes{{ else }}No{{ end }}`,
						FuncMap: funcMap,
					},
			}

			idx, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			title = cfg.SystemPrompts[idx].Title

			confirmPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Are you sure you want to remove '%s'", title),
				IsConfirm: true,
			}

			result, err := confirmPrompt.Run()
			if err != nil || strings.ToLower(result) != "y" {
				fmt.Println("Removal cancelled")
				return
			}
		} else {
			title = args[0]
		}

		found := false
		var newPrompts []config.SystemPrompt
		for _, p := range cfg.SystemPrompts {
			if p.Title != title {
				newPrompts = append(newPrompts, p)
			} else {
				found = true
			}
		}

		if !found {
			fmt.Println("Prompt not found")
			return
		}

		cfg.SystemPrompts = newPrompts

		for _, p := range cfg.SystemPrompts {
			if p.Content == cfg.SystemPrompt {
				cfg.SystemPrompt = ""
				fmt.Println("Note: Removed prompt was the active system prompt. No system prompt is now active.")
				break
			}
		}

		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Printf("System prompt '%s' removed\n", title)
	},
}

var systemListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all system prompts",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if len(cfg.SystemPrompts) == 0 {
			fmt.Println("No system prompts configured")
			return
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		fmt.Println("System prompts:")
		for _, prompt := range cfg.SystemPrompts {
			if prompt.Content == cfg.SystemPrompt {
				fmt.Printf("%s %s: %s %s\n", 
					green("*"),
					cyan(prompt.Title),
					prompt.Content,
					green("(active)"))
			} else {
				fmt.Printf("  %s: %s\n", prompt.Title, prompt.Content)
			}
		}
	},
}

var systemSelectCmd = &cobra.Command{
	Use:   "select [title]",
	Short: "Interactively select a system prompt",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if len(cfg.SystemPrompts) == 0 {
			fmt.Println("No system prompts configured")
			return
		}

		var selectedPrompt *config.SystemPrompt

		if len(args) == 0 {
			cyan := color.New(color.FgCyan).SprintFunc()
			white := color.New(color.FgWhite).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			faint := color.New(color.Faint).SprintFunc()

			var items []config.SystemPrompt
			for _, p := range cfg.SystemPrompts {
				items = append(items, p)
			}

			funcMap := template.FuncMap{
				"cyan":     cyan,
				"white":    white,
				"green":    green,
				"faint":    faint,
				"truncate": utils.TruncateText,
			}

			prompt := promptui.Select{
				Label: "Select System Prompt",
				Items: items,
				Size:  20,
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}",
					Active:   "\U0001F449 {{ .Title | cyan }}", // 👉
					Inactive: "  {{ .Title | white }}",
					Selected: "\U00002705 {{ .Title | green }}", // ✅
					Details: `
{{ "Title:" | faint }}	{{ .Title }}
{{ "Content:" | faint }}	{{ .Content | truncate 100 | faint }}
{{ "Current:" | faint }}	{{ if eq .Content $.SystemPrompt }}Yes{{ else }}No{{ end }}`,
					FuncMap: funcMap,
				},
			}

			idx, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			selectedPrompt = &items[idx]
		} else {
			title := args[0]
			for _, p := range cfg.SystemPrompts {
				if p.Title == title {
					selectedPrompt = &p
					break
				}
			}
			if selectedPrompt == nil {
				fmt.Printf("System prompt '%s' not found\n", title)
				return
			}
		}

		cfg.SystemPrompt = selectedPrompt.Content
		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("System prompt set to: %s\n", selectedPrompt.Title)
	},
}

func init() {
	systemCmd.AddCommand(systemAddCmd)
	systemCmd.AddCommand(systemRemoveCmd)
	systemCmd.AddCommand(systemListCmd)
	systemCmd.AddCommand(systemSelectCmd)
	rootCmd.AddCommand(systemCmd)
}

