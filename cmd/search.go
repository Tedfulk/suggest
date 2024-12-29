package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"github.com/tedfulk/suggest/internal/api"
	"github.com/tedfulk/suggest/internal/config"
)

var (
	topic          string
	daysBack       int
	maxResults     int
	includeAnswer  bool
	includeDomains []string
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search the web using Tavily API",
	Long: `Search the web using Tavily API.
	
Example:
  suggest search "What is quantum computing"
  suggest search --topic general "Open source alternatives for google search"
  suggest search --days 7 --topic news "latest AI developments"
  suggest search --max-results 10 "golang tutorials"
  suggest search --include-domains github.com,golang.org "go modules"`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Please provide a search query")
			return
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		if cfg.TavilyAPIKey == "" {
			fmt.Println("Tavily API key not set. Use 'suggest keys tavily' to set it.")
			return
		}

		query := strings.Join(args, " ")
		client := api.NewTavilyClient(cfg.TavilyAPIKey)

		req := api.TavilySearchRequest{
			Query:          query,
			APIKey:         cfg.TavilyAPIKey,
			SearchDepth:    "basic",
			MaxResults:     maxResults,
			Topic:          topic,
			DaysBack:       daysBack,
			IncludeAnswer:  includeAnswer,
			IncludeDomains: includeDomains,
		}

		resp, err := client.SearchWithOptions(req)
		if err != nil {
			fmt.Printf("Search failed: %v\n", err)
			return
		}

		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(100),
		)

		if resp.Answer != "" {
			fmt.Println("\nAnswer:")
			answer, _ := r.Render(resp.Answer)
			fmt.Print(answer)
		}

		fmt.Println("\nResults:")
		for i, result := range resp.Results {
			output := fmt.Sprintf("### %d. %s\n%s\n\nURL: %s\n\n",
				i+1,
				result.Title,
				result.Content,
				result.URL,
			)
			rendered, _ := r.Render(output)
			fmt.Print(rendered)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().StringVarP(&topic, "topic", "t", "general", "Search topic (general or news)")
	searchCmd.Flags().IntVarP(&daysBack, "days", "d", 3, "Number of days back for news search")
	searchCmd.Flags().IntVarP(&maxResults, "max-results", "m", 5, "Maximum number of results to return")
	searchCmd.Flags().BoolVarP(&includeAnswer, "answer", "a", true, "Include AI-generated answer")
	searchCmd.Flags().StringSliceVarP(&includeDomains, "include-domains", "i", []string{}, "Comma-separated list of domains to include")
} 