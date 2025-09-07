package yt

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Client wraps the YouTube API service
type Client struct {
	service *youtube.Service
}

// NewClient creates a new YouTube API client
func NewClient() (*Client, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing youtube api key: set GOOGLE_API_KEY environment variable")
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error creating youtube service: %w", err)
	}

	return &Client{
		service: service,
	}, nil
}

// GetService returns the underlying YouTube service for API calls
func (c *Client) GetService() *youtube.Service {
	return c.service
}
