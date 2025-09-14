package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alanpramil7/gplay/internal/yt"
	"google.golang.org/api/youtube/v3"
)

const (
	defaultOrder         = "relevance"
	defaultSafeSearch    = "moderate"
	defaultVideoDuration = "any"
	defaultVideoType     = "any"
)

// SearchService interface for YouTube search operations
type SearchService interface {
	Search(query string) (*yt.SearchResponse, error)
	SearchWithConfig(query string, config *yt.SearchConfig) (*yt.SearchResponse, error)
}

type searchService struct {
	client *yt.Client
	config *yt.SearchConfig
}

// NewSearchService creates a new search service instance
func NewSearchService(client *yt.Client, maxResults int64) SearchService {
	config := &yt.SearchConfig{
		MaxResults:    maxResults,
		Order:         defaultOrder,
		SafeSearch:    defaultSafeSearch,
		VideoDuration: defaultVideoDuration,
		VideoType:     defaultVideoType,
	}
	return &searchService{
		client: client,
		config: config,
	}
}

// Search performs a YouTube search with default configuration
func (s *searchService) Search(query string) (*yt.SearchResponse, error) {
	return s.SearchWithConfig(query, s.config)
}

// SearchWithConfig performs a YouTube search with custom configuration
func (s *searchService) SearchWithConfig(query string, config *yt.SearchConfig) (*yt.SearchResponse, error) {
	service := s.client.Service()

	// Build the search call
	call := service.Search.List([]string{"id", "snippet"}).
		Q(query).
		MaxResults(config.MaxResults).
		Order(config.Order).
		SafeSearch(config.SafeSearch).
		VideoDuration(config.VideoDuration).
		VideoType(config.VideoType).
		Type("video")

	// Execute the search
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}

	// Convert to our internal format
	results := make([]yt.SearchResult, 0, len(response.Items))
	videoIDs := make([]string, 0, len(response.Items))

	// First pass: collect video IDs and basic info
	for _, item := range response.Items {
		if item.Id != nil && item.Snippet != nil {
			videoIDs = append(videoIDs, item.Id.VideoId)

			publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

			result := yt.SearchResult{
				ID:           item.Id.VideoId,
				Title:        item.Snippet.Title,
				Description:  item.Snippet.Description,
				ChannelTitle: item.Snippet.ChannelTitle,
				ChannelID:    item.Snippet.ChannelId,
				PublishedAt:  publishedAt,
				ThumbnailURL: getBestThumbnail(item.Snippet.Thumbnails),
				URL:          fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.Id.VideoId),
			}
			results = append(results, result)
		}
	}

	// Second pass: get detailed video information
	if len(videoIDs) > 0 {
		details, err := s.getVideoDetails(service, videoIDs)
		if err != nil {
			log.Printf("Warning: failed to get video details: %v", err)
		} else {
			// Merge details with basic info
			for i, result := range results {
				if detail, exists := details[result.ID]; exists {
					results[i].Duration = detail.Duration
					results[i].ViewCount = detail.ViewCount
					results[i].LikeCount = detail.LikeCount
				}
			}
		}
	}

	return &yt.SearchResponse{
		Videos:        results,
		TotalResults:  response.PageInfo.TotalResults,
		Query:         query,
		NextPageToken: response.NextPageToken,
	}, nil
}

// VideoDetails holds additional video information
type VideoDetails struct {
	Duration  string `json:"duration"`
	ViewCount uint64 `json:"view_count"`
	LikeCount uint64 `json:"like_count"`
}

// getVideoDetails retrieves detailed information for a list of video IDs
func (s *searchService) getVideoDetails(service *youtube.Service, videoIDs []string) (map[string]VideoDetails, error) {
	call := service.Videos.List([]string{"statistics", "contentDetails"}).
		Id(strings.Join(videoIDs, ","))

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error getting video details: %w", err)
	}

	details := make(map[string]VideoDetails)
	for _, video := range response.Items {
		if video.Statistics != nil && video.ContentDetails != nil {
			viewCount := video.Statistics.ViewCount
			likeCount := video.Statistics.LikeCount

			details[video.Id] = VideoDetails{
				Duration:  video.ContentDetails.Duration,
				ViewCount: viewCount,
				LikeCount: likeCount,
			}
		}
	}

	return details, nil
}

// getBestThumbnail returns the URL of the best available thumbnail
func getBestThumbnail(thumbnails *youtube.ThumbnailDetails) string {
	if thumbnails == nil {
		return ""
	}

	// Prefer high quality thumbnails
	if thumbnails.Maxres != nil {
		return thumbnails.Maxres.Url
	}
	if thumbnails.High != nil {
		return thumbnails.High.Url
	}
	if thumbnails.Medium != nil {
		return thumbnails.Medium.Url
	}
	if thumbnails.Default != nil {
		return thumbnails.Default.Url
	}

	return ""
}
