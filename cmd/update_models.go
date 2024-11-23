package cmd

import (
	"fmt"
	"github.com/tedfulk/suggest/internal/config"

	"github.com/spf13/cobra"
)

var updateModelsCmd = &cobra.Command{
	Use:   "update-models",
	Short: "Update the models list in the configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		err = config.UpdateModels(cfg, config.ProviderAll)
		if err != nil {
			fmt.Println("Error updating models:", err)
			return
		}

		fmt.Println("Models updated in configuration file.")
	},
}

func init() {
	rootCmd.AddCommand(updateModelsCmd)
} 