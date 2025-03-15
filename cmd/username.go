package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tedfulk/suggest/internal/config"
)

var usernameCmd = &cobra.Command{
	Use:   "username",
	Short: "Set your username for chat sessions",
	Long: `Set your username that will be displayed during chat sessions.
If no username is set, it will display as "User" by default.

Example:
  suggest username  # Will prompt for username input`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		fmt.Print("Enter your username: ")
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			fmt.Println("Error reading input")
			return
		}

		username := strings.TrimSpace(scanner.Text())
		if username == "" {
			fmt.Println("Username cannot be empty")
			return
		}

		cfg.Username = username
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Username set to: %s\n", username)
	},
}

func init() {
	rootCmd.AddCommand(usernameCmd)
} 