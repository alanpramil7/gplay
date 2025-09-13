package cmd

import (
	"fmt"
	"log"

	"github.com/alanpramil7/gplay/internal/yt"
	"github.com/alanpramil7/gplay/internal/yt/services"
	"github.com/spf13/cobra"
)

// playlistCmd represents the playlist command
var playlistCmd = &cobra.Command{
	Use:   "playlist [playlistId]",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		playlistId := args[0]

		client, err := yt.NewClient()
		if err != nil {
			log.Fatalf("Error creating YouTube client: %v", err)
		}

		playlistService := services.NewPlaylistService(client)
		res, err := playlistService.GetPlaylistItems(playlistId, 100)
		if err != nil {
			log.Fatalf("Error getting playlist details: %v", err)
		}
		for _, r := range res {
			fmt.Println(r.URL)
		}
	},
}

func init() {
	rootCmd.AddCommand(playlistCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// playlistCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// playlistCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
