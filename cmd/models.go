package cmd

import (
	"fmt"
	"suggest/internal/config"

	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List or update available models",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		update, _ := cmd.Flags().GetBool("update")
		if update || len(cfg.Models.OpenAI) == 0 && len(cfg.Models.Groq) == 0 {
			fmt.Println("Updating models list...")
			err = config.UpdateModels(cfg, config.ProviderAll)
			if err != nil {
				fmt.Println("Error updating models:", err)
				return
			}
			cfg, err = config.LoadConfig()
			if err != nil {
				fmt.Println("Error reloading config:", err)
				return
			}
		}

		if len(cfg.Models.OpenAI) == 0 && len(cfg.Models.Groq) == 0 {
			fmt.Println("No models available. Please set an API key and run 'suggest models --update'")
			return
		}

		fmt.Println("Available models:")
		fmt.Println("\nOpenAI models:")
		for _, model := range cfg.Models.OpenAI {
			printModelWithAliases(model, cfg.ModelAliases)
		}

		fmt.Println("\nGroq models:")
		for _, model := range cfg.Models.Groq {
			printModelWithAliases(model, cfg.ModelAliases)
		}
	},
}

func init() {
	modelsCmd.Flags().BoolP("update", "u", false, "Update the models list")
	rootCmd.AddCommand(modelsCmd)
}

func printModelWithAliases(model string, aliases map[string]string) {
	modelAliases := []string{}
	for alias, m := range aliases {
		if m == model {
			modelAliases = append(modelAliases, alias)
		}
	}

	if len(modelAliases) > 0 {
		fmt.Printf("  %s (aliases: %v)\n", model, modelAliases)
	} else {
		fmt.Printf("  %s\n", model)
	}
} 