package cmd

import (
	"fmt"

	"github.com/alanpramil7/gplay/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

const (
	appName        = "gplay"
	appDescription = "A YouTube video search and player CLI tool"
	appLongDesc    = `GPlay is a command-line tool for searching and interacting with YouTube videos.

Features:
  • Search YouTube videos with advanced filters
  • Interactive TUI interface
  • Multiple output formats (table, JSON)
  • Configurable search parameters

Examples:
  gplay search "golang tutorial"
  gplay search "music" --max 10 --order viewCount
  gplay  # Launch interactive TUI`
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: appDescription,
	Long:  appLongDesc,
	RunE:  runTUI,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// runTUI launches the interactive TUI interface
func runTUI(cmd *cobra.Command, args []string) error {
	app := tui.NewApp()
	program := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("failed to run TUI application: %w", err)
	}

	return nil
}

func init() {
	// Add any global flags here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gplay.yaml)")
}
