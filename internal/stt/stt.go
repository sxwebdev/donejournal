package stt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tkcrm/mx/logger"
)

const (
	defaultModelName = "ggml-small.bin"
	modelDownloadURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.bin"
)

// Service provides speech-to-text transcription using whisper-cli subprocess.
type Service struct {
	logger    logger.Logger
	modelPath string
}

// New creates a new STT service. If modelPath is empty, the model is stored in
// {dataDir}/models/ggml-small.bin and downloaded automatically on first start.
func New(ctx context.Context, l logger.Logger, dataDir string, modelPath string) (*Service, error) {
	if modelPath == "" {
		modelPath = filepath.Join(dataDir, "models", defaultModelName)
	}

	if err := ensureModel(ctx, l, modelPath); err != nil {
		return nil, fmt.Errorf("failed to ensure whisper model: %w", err)
	}

	return &Service{logger: l, modelPath: modelPath}, nil
}

// Transcribe converts OGG/Opus audio bytes to text.
// Requires ffmpeg and whisper-cli to be available in PATH.
func (s *Service) Transcribe(ctx context.Context, oggData []byte) (string, error) {
	// Write OGG to temp file
	tmpOgg, err := os.CreateTemp("", "voice-*.ogg")
	if err != nil {
		return "", fmt.Errorf("failed to create temp ogg file: %w", err)
	}
	defer os.Remove(tmpOgg.Name())

	if _, err := tmpOgg.Write(oggData); err != nil {
		tmpOgg.Close()
		return "", fmt.Errorf("failed to write ogg data: %w", err)
	}
	tmpOgg.Close()

	// Convert OGG/Opus → WAV 16kHz mono using ffmpeg
	tmpWav, err := os.CreateTemp("", "voice-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp wav file: %w", err)
	}
	wavPath := tmpWav.Name()
	tmpWav.Close()
	defer os.Remove(wavPath)

	var stderr bytes.Buffer
	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", tmpOgg.Name(),
		"-ar", "16000",
		"-ac", "1",
		"-y", wavPath,
	)
	ffmpegCmd.Stderr = &stderr
	if err := ffmpegCmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %w (stderr: %s)", err, stderr.String())
	}

	// Run whisper-cli for transcription
	// Output is written to wavPath + ".txt" when using -otxt flag
	stderr.Reset()
	whisperCmd := exec.CommandContext(ctx, "whisper-cli",
		"-m", s.modelPath,
		"-l", "ru",
		"--no-timestamps",
		"-otxt",
		"-of", wavPath,
		"-f", wavPath,
	)
	whisperCmd.Stderr = &stderr
	if err := whisperCmd.Run(); err != nil {
		return "", fmt.Errorf("whisper-cli failed: %w (stderr: %s)", err, stderr.String())
	}

	// Read the output text file produced by whisper-cli
	txtPath := wavPath + ".txt"
	defer os.Remove(txtPath)

	raw, err := os.ReadFile(txtPath)
	if err != nil {
		return "", fmt.Errorf("failed to read whisper output: %w", err)
	}

	return strings.TrimSpace(string(raw)), nil
}

// ensureModel downloads the whisper model if it doesn't exist at modelPath.
func ensureModel(ctx context.Context, l logger.Logger, modelPath string) error {
	if _, err := os.Stat(modelPath); err == nil {
		return nil // already exists
	}

	if err := os.MkdirAll(filepath.Dir(modelPath), 0o700); err != nil {
		return fmt.Errorf("failed to create models directory: %w", err)
	}

	l.Infof("whisper model not found, downloading ggml-small.bin (~466MB) to %s", modelPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelDownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status when downloading model: %d", resp.StatusCode)
	}

	// Write to a temp file first, then rename atomically
	tmpPath := modelPath + ".download"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp model file: %w", err)
	}

	written, err := io.Copy(f, resp.Body)
	f.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write model: %w", err)
	}

	if err := os.Rename(tmpPath, modelPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to finalize model file: %w", err)
	}

	l.Infof("whisper model downloaded successfully (%.1f MB)", float64(written)/1024/1024)
	return nil
}
