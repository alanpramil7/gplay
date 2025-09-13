package cmd

import (
	"fmt"
	"log"

	"github.com/alanpramil7/gplay/internal/yt"
	"github.com/alanpramil7/gplay/internal/yt/services"
	"github.com/spf13/cobra"
)

var (
	maxResults    int64
	order         string
	safeSearch    string
	videoDuration string
	videoType     string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for YouTube videos",
	Long: `Search for YouTube videos using the YouTube Data API.

Examples:
  gplay search "golang tutorial"
  gplay search "music" --max 10 --order viewCount
  gplay search "cooking" --duration short --format json`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		searchQuery := args[0]

		// Create YouTube client
		client, err := yt.NewClient()
		if err != nil {
			log.Fatalf("Error creating YouTube client: %v", err)
		}

		// Create search service
		service := services.NewSearchService(client, maxResults)

		// Create search configuration
		config := &yt.SearchConfig{
			MaxResults:    maxResults,
			Order:         order,
			SafeSearch:    safeSearch,
			VideoDuration: videoDuration,
			VideoType:     videoType,
		}

		// Perform search
		results, err := service.SearchWithConfig(searchQuery, config)
		if err != nil {
			log.Fatalf("Error performing search: %v", err)
		}

		for _, result := range results.Results {
			fmt.Printf("%s\n", result.URL)
			fmt.Println(result.ThumbnailURL)
		}

	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Search parameters
	searchCmd.Flags().Int64VarP(&maxResults, "max", "m", 5, "Maximum number of results to return (1-50)")
	searchCmd.Flags().StringVarP(&order, "order", "o", "relevance", "Order of results (relevance, date, rating, viewCount, title)")
	searchCmd.Flags().StringVarP(&safeSearch, "safe", "s", "moderate", "Safe search level (none, moderate, strict)")
	searchCmd.Flags().StringVarP(&videoDuration, "duration", "d", "any", "Video duration (any, short, medium, long)")
	searchCmd.Flags().StringVarP(&videoType, "type", "t", "any", "Video type (any, episode, movie)")
}

