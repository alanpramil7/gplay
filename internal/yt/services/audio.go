package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/oto/v2"
)

const (
	defaultSampleRate = 48000
	defaultChannels   = 2
	defaultBufferSize = "64k"
	defaultLogLevel   = "warning"
)

// AudioService handles audio playback operations
type AudioService struct {
	mu              sync.Mutex
	context         *oto.Context
	player          oto.Player
	isPlaying       bool
	isPaused        bool
	currentSong     string
	cancelFunc      context.CancelFunc
	cmd             *exec.Cmd
	streamDone      chan bool
	onComplete      func()
	manuallyStopped bool
	songComplete    chan bool
}

func NewAudioService() *AudioService {
	return &AudioService{
		streamDone:   make(chan bool, 1),
		songComplete: make(chan bool, 1),
	}
}

// SetOnComplete sets the callback function to be called when a song completes
func (s *AudioService) SetOnComplete(callback func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onComplete = callback
}

// GetSongCompleteChannel returns the channel that signals when a song completes
func (s *AudioService) GetSongCompleteChannel() <-chan bool {
	return s.songComplete
}

func (s *AudioService) PlayStream(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// stop any previous playback first
	s.stopInternal()

	// Reset manually stopped flag for new song
	s.manuallyStopped = false

	if s.context == nil {
		audioContext, ready, err := oto.NewContext(defaultSampleRate, defaultChannels, defaultChannels)
		if err != nil {
			return fmt.Errorf("failed to create audio context: %w", err)
		}
		<-ready
		s.context = audioContext
	}

	streamURL, err := s.GetStreamURL(url)
	if err != nil {
		return fmt.Errorf("error getting stream url: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	// Use better FFmpeg options for streaming
	s.cmd = exec.CommandContext(ctx, "ffmpeg",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "5",
		"-i", streamURL,
		"-f", "s16le",
		"-ar", fmt.Sprintf("%d", defaultSampleRate),
		"-ac", fmt.Sprintf("%d", defaultChannels),
		"-acodec", "pcm_s16le",
		"-bufsize", defaultBufferSize,
		"-loglevel", defaultLogLevel,
		"pipe:1",
	)

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	// Create a buffered reader to help with streaming
	s.player = s.context.NewPlayer(stdout)
	s.isPlaying = true
	s.isPaused = false
	s.currentSong = url

	s.player.Play()

	// Monitor the stream in a separate goroutine
	go s.monitorStream(url)

	return nil
}

func (s *AudioService) monitorStream(songUrl string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Stream monitor recovered from panic: %v\n", r)
		}
	}()

	// Wait for the FFmpeg process to complete
	err := s.cmd.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Only clean up if we're still playing the same song
	if s.currentSong != songUrl {
		return
	}

	// Give the player a moment to finish playing buffered audio
	if s.player != nil {
		// Don't immediately close - let buffered audio finish
		time.Sleep(100 * time.Millisecond)
	}

	// Check if FFmpeg exited with an error
	if err != nil && s.cancelFunc != nil {
		fmt.Printf("FFmpeg process ended with error: %v\n", err)
	}

	// Clean up resources
	if s.player != nil {
		s.player.Close()
		s.player = nil
	}

	// Only reset state if we're still the current song
	if s.currentSong == songUrl {
		s.isPlaying = false
		s.isPaused = false
		s.currentSong = ""

		// Signal song completion if it finished naturally (not manually stopped)
		if !s.manuallyStopped {
			select {
			case s.songComplete <- true:
			default:
			}
		}
	}

	// Signal that stream is done
	select {
	case s.streamDone <- true:
	default:
	}
}

// GetStreamURL retrieves the direct stream URL for a YouTube video
func (s *AudioService) GetStreamURL(url string) (string, error) {
	// Use better format selection to avoid issues
	cmd := exec.Command("yt-dlp",
		"--get-url",
		"-f", "bestaudio[ext=m4a]/bestaudio[ext=webm]/bestaudio",
		"--no-playlist",
		url)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error getting stream url: %v", err)
	}

	streamURL := strings.TrimSpace(string(output))
	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from yt-dlp")
	}

	return streamURL, nil
}

func (s *AudioService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopInternal()
}

func (s *AudioService) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.player != nil && s.isPlaying {
		s.player.Pause()
		s.isPlaying = false
		s.isPaused = true
	}
}

func (s *AudioService) Play() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.player != nil && s.isPaused {
		s.player.Play()
		s.isPlaying = true
		s.isPaused = false
	}
}

func (s *AudioService) stopInternal() {
	s.manuallyStopped = true

	// Cancel the context first to stop FFmpeg gracefully
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	// Give FFmpeg a moment to shut down gracefully
	if s.cmd != nil && s.cmd.Process != nil {
		// Wait a short time for graceful shutdown
		time.Sleep(50 * time.Millisecond)

		// Force kill if still running
		_ = s.cmd.Process.Kill()
		s.cmd = nil
	}

	// Close the player
	if s.player != nil {
		s.player.Close()
		s.player = nil
	}

	s.isPlaying = false
	s.isPaused = false
	s.currentSong = ""
}

func (s *AudioService) IsPlaying() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isPlaying
}

func (s *AudioService) IsPaused() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isPaused
}

// GetCurrentSong returns the URL of the currently loaded song
func (s *AudioService) GetCurrentSong() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentSong
}
