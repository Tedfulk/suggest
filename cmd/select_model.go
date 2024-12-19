package cmd

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/tedfulk/suggest/internal/config"

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

		// First, select the provider
		providers := []string{}
		if len(cfg.Models.OpenAI) > 0 {
			providers = append(providers, "OpenAI")
		}
		if len(cfg.Models.Groq) > 0 {
			providers = append(providers, "Groq")
		}
		if len(cfg.Models.Gemini) > 0 {
			providers = append(providers, "Gemini")
		}

		if len(providers) == 0 {
			fmt.Println("No models available. Please set API keys and run 'suggest models --update'")
			return
		}

		providerPrompt := promptui.Select{
			Label: "Select Provider",
			Items: providers,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "\U0001F449 {{ . | cyan }}", 
				Inactive: "  {{ . | white }}",
				Selected: "\U00002705 {{ . | green }}", 
			},
		}

		_, provider, err := providerPrompt.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		// Then, select the model based on the provider
		var modelList []string
		switch provider {
		case "OpenAI":
			modelList = append(modelList, cfg.Models.OpenAI...)
		case "Groq":
			modelList = append(modelList, cfg.Models.Groq...)
		case "Gemini":
			modelList = append(modelList, cfg.Models.Gemini...)
		}

		// Sort the base model list
		sort.Strings(modelList)

		// Add aliases that match the selected provider
		var aliasList []string
		for alias, model := range cfg.ModelAliases {
			isOpenAI := strings.HasPrefix(model, "gpt-")
			isGroq := strings.HasPrefix(model, "mixtral-") || strings.HasPrefix(model, "llama-")
			isGemini := strings.HasPrefix(model, "gemini-")
			
			if (provider == "OpenAI" && isOpenAI) ||
				(provider == "Groq" && isGroq) ||
				(provider == "Gemini" && isGemini) {
				aliasList = append(aliasList, fmt.Sprintf("%s (alias for %s)", alias, model))
			}
		}
		sort.Strings(aliasList)
		modelList = append(modelList, aliasList...)

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

		modelPrompt := promptui.Select{
			Label: "Select Model",
			Items: modelList,
			Size:  20,
			Templates: &promptui.SelectTemplates{
				Label:    "{{ . }}",
				Active:   "\U0001F449 {{ . | cyan }}", 
				Inactive: "  {{ . | white }}",
				Selected: "\U00002705 {{ . | green }}", 
				FuncMap:  funcMap,
				Details: `
{{ "Provider:" | faint }}	{{ if hasPrefix . "gpt-" }}OpenAI{{ else if or (hasPrefix . "mixtral-") (hasPrefix . "llama-") }}Groq{{ else }}Unknown{{ end }}
{{ "Current:" | faint }}	{{ if eq . $.Model }}Yes{{ else }}No{{ end }}`,
			},
		}

		_, result, err := modelPrompt.Run()
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