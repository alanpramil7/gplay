package yt

import "time"

// SearchResult represents a single search result from YouTube
type SearchResult struct {
	VideoID      string    `json:"video_id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	ChannelTitle string    `json:"channel_title"`
	ChannelID    string    `json:"channel_id"`
	PublishedAt  time.Time `json:"published_at"`
	Duration     string    `json:"duration"`
	ViewCount    uint64    `json:"view_count"`
	LikeCount    uint64    `json:"like_count"`
	ThumbnailURL string    `json:"thumbnail_url"`
	URL          string    `json:"url"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Results       []SearchResult `json:"results"`
	TotalResults  int64          `json:"total_results"`
	Query         string         `json:"query"`
	NextPageToken string         `json:"next_page_token,omitempty"`
}

// SearchConfig holds configuration for search operations
type SearchConfig struct {
	MaxResults    int64  `json:"max_results"`
	Order         string `json:"order"`          // relevance, date, rating, viewCount, title
	SafeSearch    string `json:"safe_search"`    // none, moderate, strict
	VideoDuration string `json:"video_duration"` // any, short, medium, long
	VideoType     string `json:"video_type"`     // any, episode, movie
}
