package services

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/hajimehoshi/oto/v2"
)

type AudioService struct {
	mu          sync.Mutex
	context     *oto.Context
	player      oto.Player
	isPlaying   bool
	currentSong string
	cancelFunc  context.CancelFunc
	cmd         *exec.Cmd
}

func NewAudioService() *AudioService {
	return &AudioService{}
}

func (s *AudioService) PlayStream(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// stop any previous playback first
	s.stopInternal()

	if s.context == nil {
		audioContext, ready, err := oto.NewContext(48000, 2, 2)
		if err != nil {
			return fmt.Errorf("failed to create audio context: %w", err)
		}
		<-ready
		s.context = audioContext
	}

	streamUrl, err := s.GetStreamUrl(url)
	if err != nil {
		return fmt.Errorf("error getting stream url: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelFunc = cancel

	s.cmd = exec.CommandContext(ctx, "ffmpeg",
		"-i", streamUrl,
		"-f", "s16le", "-ar", "48000", "-ac", "2",
		"-loglevel", "warning", "pipe:1",
	)

	stdout, err := s.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	s.player = s.context.NewPlayer(stdout)
	s.isPlaying = true
	s.currentSong = url
	s.player.Play()

	go func(song string) {
		_ = s.cmd.Wait()
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.currentSong != song {
			return
		}

		if s.player != nil {
			s.player.Close()
			s.player = nil
		}

		if s.currentSong == song {
			s.isPlaying = false
			s.currentSong = ""
		}
	}(url)

	return nil
}

func (s *AudioService) GetStreamUrl(url string) (string, error) {
	cmd := exec.Command("yt-dlp", "--get-url", "-f", "bestaudio", url)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error get stream url: %v", err)
	}
	streamUrl := strings.TrimSpace(string(output))
	if streamUrl == "" {
		return "", fmt.Errorf("empty stream URL returned from yt-dlp")
	}
	return streamUrl, nil
}

func (s *AudioService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.player != nil {
		s.stopInternal()
		s.isPlaying = false
	}
}

func (s *AudioService) Pause() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.player != nil {
		s.player.Pause()
		s.isPlaying = false
	}
}

func (s *AudioService) Play() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.player != nil {
		s.player.Play()
		s.isPlaying = true
	}
}

func (s *AudioService) stopInternal() {
	if s.cancelFunc != nil {
		s.cancelFunc()
		s.cancelFunc = nil
	}

	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		s.cmd = nil
	}

	if s.player != nil {
		s.player.Close()
		s.player = nil
	}


	s.isPlaying = false
	s.currentSong = ""
}

func (s *AudioService) IsPlaying() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.isPlaying
}

func (s *AudioService) GetCurrentSong() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.currentSong
}
