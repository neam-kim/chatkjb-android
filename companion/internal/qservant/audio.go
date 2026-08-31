package qservant

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const MaxAudioBytes = 10 << 20

var ErrAudioVersion = errors.New("unsupported audio input version")
var ErrAudioType = errors.New("audio must be AAC/M4A")
var ErrAudioTooLarge = errors.New("audio exceeds 10 MiB")

type AudioInput struct {
	Version     int    `json:"v"`
	MIMEType    string `json:"mimeType"`
	Data        string `json:"data"`
	DebugRetain bool   `json:"debugRetain,omitempty"`
}

func DecodeAudioJSON(b []byte) (AudioInput, error) {
	var in AudioInput
	if err := json.Unmarshal(b, &in); err != nil {
		return in, err
	}
	// Accept the spelling used by older Android clients while retaining v1 as
	// the sole wire version.
	if in.Version == 0 || in.MIMEType == "" || in.Data == "" {
		var legacy struct {
			Version int    `json:"version"`
			MIME    string `json:"mime"`
			Audio   string `json:"audio"`
			Base64  string `json:"base64"`
		}
		if json.Unmarshal(b, &legacy) == nil {
			if in.Version == 0 {
				in.Version = legacy.Version
			}
			if in.MIMEType == "" {
				in.MIMEType = legacy.MIME
			}
			if in.Data == "" {
				in.Data = legacy.Audio
				if in.Data == "" {
					in.Data = legacy.Base64
				}
			}
		}
	}
	if in.Version != 1 {
		return in, ErrAudioVersion
	}
	if !allowedMIME(in.MIMEType) {
		return in, ErrAudioType
	}
	return in, nil
}

// TranscribeAudio materializes, recognizes in ko-KR, and removes the temporary
// file even when the recognizer returns an error.
func TranscribeAudio(ctx context.Context, in AudioInput, stt STT) (string, error) {
	if stt == nil {
		return "", ErrSTTUnavailable
	}
	f, err := MaterializeAudio(in)
	if err != nil {
		return "", err
	}
	defer f.Cleanup()
	return stt.Transcribe(ctx, f.Path, "ko-KR")
}

type SpeechRecognizer = STT

func allowedMIME(m string) bool {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case "audio/mp4", "audio/m4a", "audio/x-m4a", "audio/aac", "audio/x-aac", "audio/aac-adts":
		return true
	}
	return false
}

// MaterializeAudio decodes input into a private temporary file. Caller must call Cleanup.
func MaterializeAudio(in AudioInput) (*AudioFile, error) {
	raw, err := base64.StdEncoding.DecodeString(in.Data)
	if err != nil {
		return nil, fmt.Errorf("decode audio: %w", err)
	}
	if len(raw) > MaxAudioBytes {
		return nil, ErrAudioTooLarge
	}
	f, err := os.CreateTemp("", "qservant-audio-*.m4a")
	if err != nil {
		return nil, err
	}
	name := f.Name()
	if err := f.Chmod(0600); err != nil {
		f.Close()
		os.Remove(name)
		return nil, err
	}
	if _, err = f.Write(raw); err != nil {
		f.Close()
		os.Remove(name)
		return nil, err
	}
	f.Close()
	return &AudioFile{Path: name, retain: in.DebugRetain}, nil
}

type AudioFile struct {
	Path    string
	retain  bool
	cleaned bool
}

func (f *AudioFile) Cleanup() {
	if f == nil || f.cleaned {
		return
	}
	f.cleaned = true
	if !f.retain {
		_ = os.Remove(filepath.Clean(f.Path))
	}
}

type STT interface {
	Transcribe(context.Context, string, string) (string, error)
}

var ErrSTTUnavailable = errors.New("speech recognition unavailable")
var ErrSTTPermission = errors.New("speech recognition permission denied")
var ErrSTTOnDevice = errors.New("on-device speech recognition unavailable")

type FakeSTT struct {
	Text     string
	Err      error
	LastPath string
}

func (f *FakeSTT) Transcribe(ctx context.Context, path, locale string) (string, error) {
	f.LastPath = path
	if f.Err != nil {
		return "", f.Err
	}
	return f.Text, nil
}
