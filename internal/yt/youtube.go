package yt

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	envAPIKey = "GOOGLE_API_KEY"
)

// Client wraps the YouTube API service
type Client struct {
	service *youtube.Service
}

// NewClient creates a new YouTube API client
func NewClient() (*Client, error) {
	apiKey := os.Getenv(envAPIKey)
	if apiKey == "" {
		return nil, fmt.Errorf("missing YouTube API key: set %s environment variable", envAPIKey)
	}

	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	return &Client{
		service: service,
	}, nil
}

// Service returns the underlying YouTube service for API calls
func (c *Client) Service() *youtube.Service {
	return c.service
}
