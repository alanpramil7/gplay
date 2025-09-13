package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alanpramil7/gplay/internal/yt"
	"google.golang.org/api/youtube/v3"
)

// PlaylistService interface for YouTube playlist operations
type PlaylistService interface {
	GetPlaylistItems(playlistID string, max int64) ([]yt.SearchResult, error)
}

type playlistService struct {
	ytClient *yt.Client
}

// NewPlaylistService creates a new playlist service instance
func NewPlaylistService(ytClient *yt.Client) PlaylistService {
	return &playlistService{
		ytClient: ytClient,
	}
}

// GetPlaylistItems retrieves all videos (songs) from a playlist
func (p *playlistService) GetPlaylistItems(playlistID string, max int64) ([]yt.SearchResult, error) {
	service := p.ytClient.GetService()

	results := []yt.SearchResult{}
	nextPageToken := ""

	for {
		call := service.PlaylistItems.List([]string{"id", "snippet", "contentDetails"}).
			PlaylistId(playlistID).
			MaxResults(max).
			PageToken(nextPageToken)

		response, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("error fetching playlist items: %w", err)
		}

		videoIDs := make([]string, 0, len(response.Items))

		// First pass: collect video IDs and snippet info
		for _, item := range response.Items {
			if item.Snippet != nil && item.ContentDetails != nil {
				videoIDs = append(videoIDs, item.ContentDetails.VideoId)

				publishedAt, _ := time.Parse(time.RFC3339, item.Snippet.PublishedAt)

				result := yt.SearchResult{
					VideoID:      item.ContentDetails.VideoId,
					Title:        item.Snippet.Title,
					Description:  item.Snippet.Description,
					ChannelTitle: item.Snippet.ChannelTitle,
					ChannelID:    item.Snippet.ChannelId,
					PublishedAt:  publishedAt,
					ThumbnailURL: getBestThumbnail(item.Snippet.Thumbnails),
					URL:          fmt.Sprintf("https://www.youtube.com/watch?v=%s", item.ContentDetails.VideoId),
				}
				results = append(results, result)
			}
		}

		// Second pass: get extra details
		if len(videoIDs) > 0 {
			details, err := p.getVideoDetails(service, videoIDs)
			if err != nil {
				log.Printf("Warning: failed to get video details: %v", err)
			} else {
				for i, result := range results {
					if detail, exists := details[result.VideoID]; exists {
						results[i].Duration = detail.Duration
						results[i].ViewCount = detail.ViewCount
						results[i].LikeCount = detail.LikeCount
					}
				}
			}
		}

		// Handle pagination
		if response.NextPageToken == "" {
			break
		}
		nextPageToken = response.NextPageToken
	}

	return results, nil
}

// getVideoDetails retrieves detailed information for a list of video IDs
func (p *playlistService) getVideoDetails(service *youtube.Service, videoIDs []string) (map[string]VideoDetails, error) {
	call := service.Videos.List([]string{"statistics", "contentDetails"}).
		Id(strings.Join(videoIDs, ","))

	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error getting video details: %w", err)
	}

	details := make(map[string]VideoDetails)
	for _, video := range response.Items {
		if video.Statistics != nil && video.ContentDetails != nil {
			details[video.Id] = VideoDetails{
				Duration:  video.ContentDetails.Duration,
				ViewCount: video.Statistics.ViewCount,
				LikeCount: video.Statistics.LikeCount,
			}
		}
	}

	return details, nil
}
