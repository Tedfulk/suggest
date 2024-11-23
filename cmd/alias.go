package cmd

import (
	"fmt"
	"suggest/internal/config"

	"github.com/spf13/cobra"
)

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage model aliases",
	Long: `Manage model aliases. Usage:
  suggest alias add <alias> <model>    - Add or update an alias
  suggest alias remove <alias>         - Remove an alias
  suggest alias list                   - List all aliases`,
}

var aliasAddCmd = &cobra.Command{
	Use:   "add <alias> <model>",
	Short: "Add or update a model alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]
		model := args[1]

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		modelExists := false
		for _, m := range cfg.Models.OpenAI {
			if m == model {
				modelExists = true
				break
			}
		}
		if !modelExists {
			for _, m := range cfg.Models.Groq {
				if m == model {
					modelExists = true
					break
				}
			}
		}

		if !modelExists {
			fmt.Printf("Warning: Model '%s' is not in the list of available models\n", model)
		}

		cfg.ModelAliases[alias] = model
		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Printf("Alias '%s' set to model '%s'\n", alias, model)
	},
}

var aliasRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Remove a model alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if _, exists := cfg.ModelAliases[alias]; !exists {
			fmt.Printf("Alias '%s' does not exist\n", alias)
			return
		}

		delete(cfg.ModelAliases, alias)
		err = config.SaveConfig(cfg)
		if err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Printf("Alias '%s' removed\n", alias)
	},
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all model aliases",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if len(cfg.ModelAliases) == 0 {
			fmt.Println("No aliases configured")
			return
		}

		fmt.Println("Configured aliases:")
		for alias, model := range cfg.ModelAliases {
			fmt.Printf("  %s -> %s\n", alias, model)
		}
	},
}

func init() {
	aliasCmd.AddCommand(aliasAddCmd)
	aliasCmd.AddCommand(aliasRemoveCmd)
	aliasCmd.AddCommand(aliasListCmd)
	rootCmd.AddCommand(aliasCmd)
} 