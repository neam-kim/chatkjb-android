package engine

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/qservant"
)

type SwiftSTT struct {
	runner qservant.CommandRunner
}

func NewSwiftSTT() *SwiftSTT {
	return NewSwiftSTTWithRunner(qservant.ExecCommandRunner{})
}

func NewSwiftSTTWithRunner(r qservant.CommandRunner) *SwiftSTT {
	if r == nil {
		r = qservant.ExecCommandRunner{}
	}
	return &SwiftSTT{runner: r}
}

func (s *SwiftSTT) Transcribe(ctx context.Context, audioPath, locale string) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", qservant.ErrSTTUnavailable
	}
	if locale == "" {
		locale = "ko-KR"
	}
	script := fmt.Sprintf(`
import Foundation
import Speech
import AVFoundation

let fileURL = URL(fileURLWithPath: %q)
let requestedLocale = Locale(identifier: %q)

Task {
    do {
        guard #available(macOS 26.0, *) else {
            print("speech recognition unavailable: macOS 26 required")
            exit(1)
        }
        guard SpeechTranscriber.isAvailable,
              let locale = await SpeechTranscriber.supportedLocale(equivalentTo: requestedLocale) else {
            print("on-device speech recognition unavailable for locale")
            exit(1)
        }
        let transcriber = SpeechTranscriber(locale: locale, preset: .transcription)
        let modules: [any SpeechModule] = [transcriber]
        let assetStatus = await AssetInventory.status(forModules: modules)
        if assetStatus != .installed,
           let installation = try await AssetInventory.assetInstallationRequest(supporting: modules) {
            try await installation.downloadAndInstall()
        }
        let audioFile = try AVAudioFile(forReading: fileURL)
        let analyzer = SpeechAnalyzer(modules: modules)
        try await analyzer.start(inputAudioFile: audioFile, finishAfterFile: true)
        var resultText = ""
        for try await result in transcriber.results where result.isFinal {
            resultText += String(result.text.characters)
        }
        resultText = resultText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !resultText.isEmpty else {
            print("speech recognition failed: empty transcription")
            exit(1)
        }
        print(resultText)
        exit(0)
    } catch {
        print("speech recognition failed: \(error.localizedDescription)")
        exit(1)
    }
}
dispatchMain()
`, audioPath, locale)

	out, err := s.runner.Run(ctx, "/usr/bin/swift", "-e", script)
	if err != nil {
		message := strings.ToLower(strings.TrimSpace(string(out)))
		switch {
		case strings.Contains(message, "permission"):
			return "", qservant.ErrSTTPermission
		case strings.Contains(message, "on-device"), strings.Contains(message, "asset"):
			return "", qservant.ErrSTTOnDevice
		case strings.Contains(message, "unavailable"):
			return "", qservant.ErrSTTUnavailable
		}
		return "", fmt.Errorf("speech recognition failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
