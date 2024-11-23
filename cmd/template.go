package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"github.com/tedfulk/suggest/internal/config"
	"text/template"

	"github.com/tedfulk/suggest/internal/utils"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage message templates",
	Long: `Manage message templates for common patterns.
	
Example:
  suggest template add code "Write a [language] function that [task]"
  suggest template use code --vars "language=Python,task=sorts a list"
  suggest template remove code
  suggest template list`,
}

var templateAddCmd = &cobra.Command{
	Use:   "add [title] [template]",
	Short: "Add a new template",
	Long: `Add a new template. Can be used in two ways:
	
1. Interactive mode:
   suggest template add

2. Direct mode:
   suggest template add "Code Function" "Write a [language] function that [task]"`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		var title, content string

		if len(args) == 0 {
			fmt.Print("Enter template title: ")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				title = scanner.Text()
			}

			fmt.Println("Enter template content (use [variable] for placeholders, Shift+Enter for new line, Enter to finish):")
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

		for _, t := range cfg.Templates {
			if t.Title == title {
				fmt.Printf("Template '%s' already exists. Use a different title or remove it first.\n", title)
				return
			}
		}

		cfg.Templates = append(cfg.Templates, config.Template{
			Title:   title,
			Content: content,
		})

		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Printf("Template '%s' added\n", title)

		vars := utils.ExtractVariables(content)
		if len(vars) > 0 {
			fmt.Println("\nTemplate variables:")
			for _, v := range vars {
				fmt.Printf("  %s\n", v)
			}
			fmt.Printf("\nUse with: suggest template use %s --vars \"", title)
			for i, v := range vars {
				if i > 0 {
					fmt.Print(",")
				}
				fmt.Printf("%s=value", v)
			}
			fmt.Println("\"")
		}
	},
}

var templateRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a template",
	Long: `Remove a template. Can be used in two ways:
	
1. Interactive mode:
   suggest template remove

2. Direct mode:
   suggest template remove "name"`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if len(cfg.Templates) == 0 {
			fmt.Println("No templates configured")
			return
		}

		var name string

		if len(args) == 0 {
			cyan := color.New(color.FgCyan).SprintFunc()
			white := color.New(color.FgWhite).SprintFunc()
			red := color.New(color.FgRed).SprintFunc()
			faint := color.New(color.Faint).SprintFunc()

			var items []config.Template
			for _, t := range cfg.Templates {
				items = append(items, t)
			}

			funcMap := template.FuncMap{
				"cyan":     cyan,
				"white":    white,
				"red":      red,
				"faint":    faint,
				"truncate": utils.TruncateText,
			}

			prompt := promptui.Select{
				Label: "Select Template to Remove",
					Items: items,
					Size:  20,
					Templates: &promptui.SelectTemplates{
						Label:    "{{ . }}",
						Active:   "\U0001F449 {{ .Title | cyan }}", // 👉
						Inactive: "  {{ .Title | white }}",
						Selected: "\U0001F5D1 {{ .Title | red }}", // 🗑
						Details: `
{{ "Title:" | faint }}	{{ .Title }}
{{ "Content:" | faint }}	{{ .Content | truncate 100 | faint }}`,
						FuncMap: funcMap,
					},
				}

				idx, _, err := prompt.Run()
				if err != nil {
					fmt.Printf("Prompt failed %v\n", err)
					return
				}

				name = items[idx].Title

				confirmPrompt := promptui.Prompt{
					Label:     fmt.Sprintf("Are you sure you want to remove template '%s'", name),
					IsConfirm: true,
				}

				result, err := confirmPrompt.Run()
				if err != nil || strings.ToLower(result) != "y" {
					fmt.Println("Removal cancelled")
					return
				}
		} else {
			name = args[0]
		}

		found := false
		var newTemplates []config.Template
		for _, t := range cfg.Templates {
			if t.Title != name {
				newTemplates = append(newTemplates, t)
			} else {
				found = true
			}
		}

		if !found {
			fmt.Printf("Template '%s' not found\n", name)
			return
		}

		cfg.Templates = newTemplates
		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Printf("Template '%s' removed\n", name)
	},
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all templates",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if len(cfg.Templates) == 0 {
			fmt.Println("No templates configured")
			return
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		faint := color.New(color.Faint).SprintFunc()

		fmt.Println("Available templates:")
		for _, template := range cfg.Templates {
			vars := utils.ExtractVariables(template.Content)
			content := template.Content
			for _, v := range vars {
				content = strings.ReplaceAll(content, "["+v+"]", yellow("["+v+"]"))
			}

			fmt.Printf("  %s: %s\n", cyan(template.Title), content)

			if len(vars) > 0 {
				fmt.Printf("    %s %s\n", faint("Variables:"), yellow(strings.Join(vars, ", ")))
			}
		}
	},
}

var (
	templateVars string
)

var templateSelectCmd = &cobra.Command{
	Use:   "select [title]",
	Short: "Interactively select and use a template",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if len(cfg.Templates) == 0 {
			fmt.Println("No templates configured")
			return
		}

		var selectedTemplate *config.Template

		if len(args) == 0 {
			cyan := color.New(color.FgCyan).SprintFunc()
			white := color.New(color.FgWhite).SprintFunc()
			green := color.New(color.FgGreen).SprintFunc()
			faint := color.New(color.Faint).SprintFunc()
			yellow := color.New(color.FgYellow).SprintFunc()

			var items []config.Template
			for _, t := range cfg.Templates {
				items = append(items, t)
			}

			funcMap := template.FuncMap{
				"cyan":     cyan,
				"white":    white,
				"green":    green,
				"faint":    faint,
				"yellow":   yellow,
				"truncate": utils.TruncateText,
				"extractVariables": utils.ExtractVariables,
				"join":     strings.Join,
			}

			prompt := promptui.Select{
				Label: "Select Template",
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
{{ "Variables:" | faint }}	{{ extractVariables .Content | join ", " | yellow }}`,
					FuncMap: funcMap,
				},
			}

			idx, _, err := prompt.Run()
			if err != nil {
				fmt.Printf("Prompt failed %v\n", err)
				return
			}

			selectedTemplate = &items[idx]
		} else {
			title := args[0]
			for _, t := range cfg.Templates {
				if t.Title == title {
					selectedTemplate = &t
					break
				}
			}
			if selectedTemplate == nil {
				fmt.Printf("Template '%s' not found\n", title)
				return
			}
		}

		message := selectedTemplate.Content
		if templateVars != "" {
			vars := parseTemplateVars(templateVars)
			for key, value := range vars {
				message = strings.ReplaceAll(message, "["+key+"]", value)
			}
		}

		rootCmd.SetArgs(strings.Fields(message))
		rootCmd.Execute()
	},
}

func parseTemplateVars(vars string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(vars, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}

func init() {
	templateSelectCmd.Flags().StringVarP(&templateVars, "vars", "v", "", "Variables for template (format: key1=value1,key2=value2)")
	
	templateCmd.AddCommand(templateAddCmd)
	templateCmd.AddCommand(templateRemoveCmd)
	templateCmd.AddCommand(templateListCmd)
	templateCmd.AddCommand(templateSelectCmd)
	rootCmd.AddCommand(templateCmd)
} 