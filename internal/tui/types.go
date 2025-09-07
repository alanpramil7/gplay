package tui

import (
	"github.com/alanpramil7/gplay/internal/yt"
	// "github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

// AppState represents the current state of the application
type AppState int

const (
	StateNormal AppState = iota
	StateSearchInput
	StateLoading
)

// AppModel represents the app state
type AppModel struct {
	state         AppState
	searchInput   textinput.Model
	results       viewport.Model
	searchResults []yt.SearchResult
	selected      int
	width, height int
	err           error
}

// Custom messages for async operations
type searchStartMsg string
type searchCompleteMsg *yt.SearchResponse
type searchErrorMsg error

// ListItem implements list.Item for search results
type ListItem struct {
	result yt.SearchResult
}

func (i ListItem) Title() string       { return i.result.Title }
func (i ListItem) Description() string { return i.result.ChannelTitle }
func (i ListItem) FilterValue() string { return i.result.Title }
