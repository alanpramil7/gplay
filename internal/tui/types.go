package tui

import (
	"github.com/alanpramil7/gplay/internal/yt"
	"github.com/alanpramil7/gplay/internal/yt/services"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// State represents the current state of the application
type State int

const (
	StateNormal State = iota
	StateSearchInput
	StateLoading
)

// Model represents the TUI application state
type Model struct {
	state         State
	client        *yt.Client
	searchInput   textinput.Model
	results       viewport.Model
	searchResults []yt.SearchResult
	searchMode    SearchMode
	selected      int
	selectedItem  *yt.SearchResult
	isLoadingSong bool
	width, height int
	err           error

	AudioService    *services.AudioService
	PlaylistService services.PlaylistService
}

// Custom messages for async operations
type searchStartMsg string
type searchCompleteMsg []yt.SearchResult
type searchErrorMsg error
type songCompleteMsg struct{}

// AppModel is an alias for Model for backward compatibility
type AppModel = Model

// Search mode for playlist or song
type SearchMode int

const (
	SearchModeQuery SearchMode = iota
	SearchModePlaylist
)
