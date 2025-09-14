package cmd

import (
	"fmt"

	"github.com/alanpramil7/gplay/internal/yt"
	"github.com/alanpramil7/gplay/internal/yt/services"
	"github.com/spf13/cobra"
)

const (
	defaultMaxResults    = int64(5)
	defaultOrder         = "relevance"
	defaultSafeSearch    = "moderate"
	defaultVideoDuration = "any"
	defaultVideoType     = "any"
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
	RunE: runSearch,
}

// runSearch executes the search command
func runSearch(cmd *cobra.Command, args []string) error {
	searchQuery := args[0]

	// Create YouTube client
	client, err := yt.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create YouTube client: %w", err)
	}

	// Create search service
	searchService := services.NewSearchService(client, maxResults)

	// Create search configuration
	config := &yt.SearchConfig{
		MaxResults:    maxResults,
		Order:         order,
		SafeSearch:    safeSearch,
		VideoDuration: videoDuration,
		VideoType:     videoType,
	}

	// Perform search
	results, err := searchService.SearchWithConfig(searchQuery, config)
	if err != nil {
		return fmt.Errorf("failed to perform search: %w", err)
	}

	// Display results
	for _, result := range results.Videos {
		fmt.Printf("Title: %s\n", result.Title)
		fmt.Printf("Channel: %s\n", result.ChannelTitle)
		fmt.Printf("URL: %s\n", result.URL)
		fmt.Printf("Thumbnail: %s\n", result.ThumbnailURL)
		fmt.Println("---")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Search parameters
	searchCmd.Flags().Int64VarP(&maxResults, "max", "m", defaultMaxResults, "Maximum number of results to return (1-50)")
	searchCmd.Flags().StringVarP(&order, "order", "o", defaultOrder, "Order of results (relevance, date, rating, viewCount, title)")
	searchCmd.Flags().StringVarP(&safeSearch, "safe", "s", defaultSafeSearch, "Safe search level (none, moderate, strict)")
	searchCmd.Flags().StringVarP(&videoDuration, "duration", "d", defaultVideoDuration, "Video duration (any, short, medium, long)")
	searchCmd.Flags().StringVarP(&videoType, "type", "t", defaultVideoType, "Video type (any, episode, movie)")
}
