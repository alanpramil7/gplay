package cmd

import (
	"log"
	"os"

	"github.com/alanpramil7/gplay/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gplay",
	Short: "A YouTube video search and player CLI tool",
	Long: `GPlay is a command-line tool for searching and interacting with YouTube videos.

Features:
  • Search YouTube videos with advanced filters
  • Interactive TUI interface
  • Multiple output formats (table, JSON)
  • Configurable search parameters

Examples:
  gplay search "golang tutorial"
  gplay search "music" --max 10 --order viewCount
  gplay  # Launch interactive TUI`,
	Run: func(cmd *cobra.Command, args []string) {
		// Launch interactive TUI
		app := tui.NewApp()
		program := tea.NewProgram(app, tea.WithAltScreen())
		if _, err := program.Run(); err != nil {
			log.Fatalf("Error running application: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gplay.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
