package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/alanpramil7/gplay/internal/yt"
	"github.com/alanpramil7/gplay/internal/yt/services"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultPlaylistID = "PLdavpelzZMWVhADtPAMJWzGrT0OKVpDAp"
	defaultMaxResults = int64(100)
	searchCharLimit   = 100
	searchWidth       = 50
)

// UI color constants
const (
	colorPrimary   = "#00D9FF"
	colorSecondary = "#BD93F9"
	colorText      = "#F8F8F2"
	colorMuted     = "#6272A4"
	colorBorder    = "#3C3C3C"
	colorError     = "#FF5555"
	colorWarning   = "#FFB86C"
	colorSuccess   = "#50FA7B"
	colorPaused    = "#F1FA8C"
	colorHelp      = "#626262"
)

// Styles for the UI with modern transparent design
var (
	leftPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 1)

	rightPanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorBorder)).
			Padding(0, 1)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(colorPrimary)).
			Padding(1, 3).
			Margin(1, 0)

	modalTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorHelp))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true).
			MarginBottom(1).
			PaddingLeft(1)

	emptyStateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorHelp)).
			Italic(true).
			Align(lipgloss.Center).
			MarginTop(2)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorError)).
			Bold(true)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWarning)).
			Bold(true)
)

// NewApp creates a new TUI application instance
func NewApp() *AppModel {
	// Initialize text input
	searchInput := textinput.New()
	searchInput.Placeholder = "Enter search query..."
	searchInput.Focus()
	searchInput.CharLimit = searchCharLimit
	searchInput.Width = searchWidth
	searchInput.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorPrimary))
	searchInput.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(colorText))

	// Initialize viewport
	resultsViewport := viewport.New(0, 0)
	resultsViewport.MouseWheelEnabled = true

	// Initialize YouTube client
	client, err := yt.NewClient()
	if err != nil {
		log.Fatalf("Failed to create YouTube client: %v", err)
	}

	// Initialize services
	audioService := services.NewAudioService()
	playlistService := services.NewPlaylistService(client)

	// Load initial playlist
	initialResults, err := playlistService.GetPlaylistItems(defaultPlaylistID, defaultMaxResults)
	if err != nil {
		log.Fatalf("Failed to load initial playlist: %v", err)
	}

	app := &AppModel{
		state:         StateNormal,
		client:        client,
		searchInput:   searchInput,
		results:       resultsViewport,
		searchResults: initialResults,
		selected:      0,
		isLoadingSong: false,

		AudioService:    audioService,
		PlaylistService: &playlistService,
	}

	// Set up completion callback (kept for compatibility)
	audioService.SetOnComplete(func() {
		// This will be called when a song completes naturally
		// We'll handle this in the Update method
	})

	return app
}

func (m *AppModel) Init() tea.Cmd {
	return m.listenForSongCompletion()
}

// listenForSongCompletion returns a command that listens for song completion
func (m *AppModel) listenForSongCompletion() tea.Cmd {
	return func() tea.Msg {
		<-m.AudioService.GetSongCompleteChannel()
		return songCompleteMsg{}
	}
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		leftWidth := int(float64(msg.Width)*0.2) - 6
		panelHeight := msg.Height - 6

		m.results = viewport.New(leftWidth-4, panelHeight-2)
		m.results.MouseWheelEnabled = true
		m.updateResultsViewport()

	case tea.KeyMsg:
		switch m.state {
		case StateNormal:
			return m.handleNormalKeys(msg)
		case StateSearchInput:
			return m.handleSearchInputKeys(msg)
		case StateLoading:
			return m.handleLoadingKeys(msg)
		}

	case searchCompleteMsg:
		m.state = StateNormal
		m.searchResults = msg.Videos
		m.selected = 0
		m.updateResultsViewport()

	case searchErrorMsg:
		m.state = StateNormal
		m.err = msg

	case songLoadCompleteMsg:
		m.isLoadingSong = false
		// Continue listening for song completion
		return m, m.listenForSongCompletion()

	case songLoadErrorMsg:
		m.isLoadingSong = false
		m.err = msg.error

	case songCompleteMsg:
		// Song completed naturally, play next song
		if m.selected < len(m.searchResults)-1 {
			m.selected++
			m.selectedItem = &m.searchResults[m.selected]
			m.updateResultsViewport()
			m.isLoadingSong = true
			return m, m.playSelectedSong()
		}
		// No more songs, continue listening for completion
		return m, m.listenForSongCompletion()

	default:
		var cmd tea.Cmd
		m.results, cmd = m.results.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *AppModel) handleNormalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		// Stop audio before quitting
		m.AudioService.Stop()
		return m, tea.Quit
	case "/", "s":
		m.state = StateSearchInput
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		return m, textinput.Blink
	case "up", "k":
		if m.selected > 0 {
			m.selected--
			m.updateResultsViewport()
		}
	case "down", "j":
		if m.selected < len(m.searchResults)-1 {
			m.selected++
			m.updateResultsViewport()
		}
	case "enter":
		if len(m.searchResults) > 0 && m.selected >= 0 && m.selected < len(m.searchResults) {
			m.selectedItem = &m.searchResults[m.selected]
			m.isLoadingSong = true
			return m, m.playSelectedSong()
		}

	case " ":
		if m.AudioService.IsPlaying() {
			m.AudioService.Pause()
		} else {
			if m.selectedItem != nil {
				if m.AudioService.GetCurrentSong() == m.selectedItem.URL {
					m.AudioService.Play()
				} else {
					m.isLoadingSong = true
					return m, m.playSelectedSong()
				}
			} else if len(m.searchResults) > 0 && m.selected >= 0 && m.selected < len(m.searchResults) {
				m.selectedItem = &m.searchResults[m.selected]
				m.isLoadingSong = true
				return m, m.playSelectedSong()
			}
		}
	case "x":
		m.AudioService.Stop()
	}
	return m, nil
}

func (m *AppModel) handleSearchInputKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = StateNormal
		m.searchInput.Blur()
		return m, nil
	case "enter":
		query := m.searchInput.Value()
		if strings.TrimSpace(query) == "" {
			m.state = StateNormal
			m.searchInput.Blur()
			return m, nil
		}
		m.state = StateLoading
		m.searchInput.Blur()
		return m, m.performSearch(query)
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}
}

func (m *AppModel) handleLoadingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	return m, nil
}

type songLoadCompleteMsg struct{}
type songLoadErrorMsg struct {
	error error
}

func (m *AppModel) playSelectedSong() tea.Cmd {
	return func() tea.Msg {
		if m.selectedItem == nil {
			return songLoadErrorMsg{fmt.Errorf("no song selected")}
		}

		err := m.AudioService.PlayStream(m.selectedItem.URL)
		if err != nil {
			return songLoadErrorMsg{err}
		}

		return songLoadCompleteMsg{}
	}
}

func (m *AppModel) performSearch(query string) tea.Cmd {
	return func() tea.Msg {
		service := services.NewSearchService(m.client, 10)
		results, err := service.Search(query)
		if err != nil {
			return searchErrorMsg(fmt.Errorf("search failed: %w", err))
		}
		return searchCompleteMsg(results)
	}
}

func (m *AppModel) updateResultsViewport() {
	var b strings.Builder
	for i, r := range m.searchResults {
		if i == m.selected {
			indicator := lipgloss.NewStyle().Foreground(lipgloss.Color(colorPrimary)).Render("▶ ")
			title := lipgloss.NewStyle().Foreground(lipgloss.Color(colorPrimary)).Bold(true).
				Render(truncate(r.Title, 40))
			channel := lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary)).Italic(true).
				Render(r.ChannelTitle)
			fmt.Fprintf(&b, "%s%s\n  %s\n", indicator, title, channel)
		} else {
			title := lipgloss.NewStyle().Foreground(lipgloss.Color(colorText)).
				Render(truncate(r.Title, 40))
			channel := lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).
				Render(r.ChannelTitle)
			fmt.Fprintf(&b, "  %s\n  %s\n", title, channel)
		}
	}
	m.results.SetContent(b.String())

	// keep selected visible
	linesPerItem := 2
	start := m.selected * linesPerItem
	end := start + linesPerItem - 1
	visible := m.results.VisibleLineCount()

	if start < m.results.YOffset {
		m.results.SetYOffset(start)
	} else if end >= m.results.YOffset+visible {
		m.results.SetYOffset(end - visible + 1)
	}
}

func (m *AppModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	leftWidth := int(float64(m.width)*0.2) - 1
	rightWidth := m.width - leftWidth - 4
	panelHeight := m.height - 4

	leftContent := ""
	if len(m.searchResults) == 0 {
		emptyMsg := `
    Press '/' or 's' to search
    Press 'q' to quit`
		leftContent = emptyStateStyle.
			Width(leftWidth - 4).
			Height(panelHeight - 4).
			Render(emptyMsg)
	} else {
		title := titleStyle.Render("Search Results")
		leftContent = title + "\n" + m.results.View()
	}
	leftPanel := leftPanelStyle.
		Width(leftWidth).
		Height(panelHeight).
		Render(leftContent)

	rightTitle := titleStyle.Render("Player")
	var rightContent string

	var statusLine string
	if m.isLoadingSong {
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorWarning)).
			Bold(true).
			Render("⏳ LOADING...")
	} else if m.AudioService.IsPlaying() {
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorSuccess)).
			Bold(true).
			Render("▶ NOW PLAYING")
	} else if m.selectedItem != nil && m.AudioService.GetCurrentSong() == m.selectedItem.URL {
		// Song is loaded but paused
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPaused)).
			Bold(true).
			Render("⏸ PAUSED")
	} else {
		statusLine = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted)).
			Render("⏹ STOPPED")
	}

	if m.selectedItem != nil {
		rightContent = fmt.Sprintf(
			"%s\n\n%s\n\nChannel: %s\n\nVideo ID: %s\n\nDescription: %s\n\nDuration: %s\n\nThumbnail URL: %s\n\nURL: %s",
			statusLine,
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colorPrimary)).Render(m.selectedItem.Title),
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorSecondary)).Italic(true).Render(m.selectedItem.ChannelTitle),
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render(m.selectedItem.ID),
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render(truncate(m.selectedItem.Description, 100)),
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render(m.selectedItem.Duration),
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render(m.selectedItem.ThumbnailURL),
			lipgloss.NewStyle().Foreground(lipgloss.Color(colorMuted)).Render(m.selectedItem.URL),
		)
	} else {
		rightContent = statusLine + "\n\n" + emptyStateStyle.Render("No video selected")
	}

	rightPanel := rightPanelStyle.
		Width(rightWidth).
		Height(panelHeight).
		Render(rightTitle + "\n" + rightContent)

	mainView := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Updated help text with consistent loading states
	helpText := ""
	switch m.state {
	case StateNormal:
		if m.isLoadingSong {
			helpText = loadingStyle.Render("Loading song...")
		} else if len(m.searchResults) > 0 {
			if m.AudioService.IsPlaying() {
				helpText = "'/' search  •  ↑↓ navigate  •  ↵ play  •  space pause  •  x stop  •  q quit"
			} else if m.selectedItem != nil && m.AudioService.GetCurrentSong() == m.selectedItem.URL {
				helpText = "'/' search  •  ↑↓ navigate  •  ↵ play  •  space resume  •  x stop  •  q quit"
			} else {
				helpText = "'/' search  •  ↑↓ navigate  •  ↵ play  •  space toggle  •  x stop  •  q quit"
			}
		} else {
			helpText = "Press '/' or 's' to search  •  Press 'q' to quit"
		}
	case StateLoading:
		helpText = loadingStyle.Render("Searching YouTube...")
	}

	if m.err != nil {
		helpText = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
		m.err = nil
	}
	help := helpStyle.Render(helpText)

	if m.state == StateSearchInput {
		title := modalTitleStyle.Render("Search YouTube")
		input := m.searchInput.View()
		helperText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true).
			Render("↵ Enter to search  •  ESC to cancel")

		modalContent := fmt.Sprintf("%s\n\n%s\n\n%s", title, input, helperText)
		modal := modalStyle.Render(modalContent)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal,
			lipgloss.WithWhitespaceBackground(lipgloss.NoColor{}))
	}

	return mainView + "\n" + help
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
