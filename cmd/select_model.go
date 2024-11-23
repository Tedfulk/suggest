package cmd

import (
	"fmt"
	"strings"
	"github.com/tedfulk/suggest/internal/config"
	"text/template"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var selectModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Interactively select a model to use",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		var allModels []string
		if len(cfg.Models.OpenAI) > 0 {
			allModels = append(allModels, cfg.Models.OpenAI...)
		}
		if len(cfg.Models.Groq) > 0 {
			allModels = append(allModels, cfg.Models.Groq...)
		}

		if len(allModels) == 0 {
			fmt.Println("No models available. Please set API keys and run 'suggest models --update'")
			return
		}

		for alias, model := range cfg.ModelAliases {
			allModels = append(allModels, fmt.Sprintf("%s (alias for %s)", alias, model))
		}

		cyan := color.New(color.FgCyan).SprintFunc()
		white := color.New(color.FgWhite).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		faint := color.New(color.Faint).SprintFunc()

		funcMap := template.FuncMap{
			"hasPrefix": strings.HasPrefix,
			"cyan":     cyan,
			"white":    white,
			"green":    green,
			"faint":    faint,
		}

		prompt := promptui.Select{
			Label: "Select Model",
			Items: allModels,
			Size:  20,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "\U0001F449 {{ . | cyan }}",  // 👉 pointer emoji
				Inactive: "  {{ . | white }}",
				Selected: "\U00002705 {{ . | green }}",  // ✅ checkmark emoji
				FuncMap:  funcMap,
				Details: `
{{ "Provider:" | faint }}	{{ if hasPrefix . "gpt-" }}OpenAI{{ else if or (hasPrefix . "mixtral-") (hasPrefix . "llama-") }}Groq{{ else }}Unknown{{ end }}
{{ "Current:" | faint }}	{{ if eq . $.Model }}Yes{{ else }}No{{ end }}`,
			},
		}

		_, result, err := prompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		model := result
		if strings.HasSuffix(result, ")") {
			parts := strings.Split(result, " (alias for ")
			if len(parts) == 2 {
				model = strings.TrimSuffix(parts[1], ")")
			}
		}

		cfg.Model = model
		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Model set to: %s\n", model)
	},
}

func init() {
	rootCmd.AddCommand(selectModelCmd)
} 