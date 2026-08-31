import Foundation

#if canImport(Speech)
import AVFoundation
import Speech

public enum SpeechSTTError: Error { case unavailable, permissionDenied, onDeviceUnavailable, recognitionFailed(String) }

/// macOS 26's file-based SpeechAnalyzer path works in a LaunchAgent and uses
/// system-managed local assets. It avoids the callback/run-loop dependency of
/// legacy SFSpeechURLRecognitionRequest in a headless companion process.
public final class SpeechSTTAdapter {
    public init() {}

    public func transcribe(fileURL: URL, localeIdentifier: String = "ko-KR",
                           completion: @escaping (Result<String, Error>) -> Void) {
        Task {
            do {
                completion(.success(try await transcribe(fileURL: fileURL, localeIdentifier: localeIdentifier)))
            } catch {
                completion(.failure(error))
            }
        }
    }

    @available(macOS 26.0, *)
    public func transcribe(fileURL: URL, localeIdentifier: String = "ko-KR") async throws -> String {
        guard SpeechTranscriber.isAvailable,
              let locale = await SpeechTranscriber.supportedLocale(equivalentTo: Locale(identifier: localeIdentifier)) else {
            throw SpeechSTTError.onDeviceUnavailable
        }
        let transcriber = SpeechTranscriber(locale: locale, preset: .transcription)
        let modules: [any SpeechModule] = [transcriber]
        if await AssetInventory.status(forModules: modules) != .installed,
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
        let trimmed = resultText.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else {
            throw SpeechSTTError.recognitionFailed("empty transcription")
        }
        return trimmed
    }
}
#else
/// Non-macOS builds retain a type-checkable placeholder; production deployment
/// compiles this file on macOS with Speech.framework linked.
public final class SpeechSTTAdapter { public init() {} }
#endif
