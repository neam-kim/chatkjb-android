package dev.herdr.mobile.features.qservant

import android.media.MediaRecorder
import java.io.File

const val Q_SERVANT_MAX_AUDIO_BYTES: Long = 10L * 1024L * 1024L

enum class RecordingFailure { PermissionDenied, StartFailed, StopFailed, EncodingFailed, TooLarge }
data class RecordingResult(val file: File? = null, val failure: RecordingFailure? = null, val message: String? = null) {
    val ok get() = file != null && failure == null
}

/** Purely testable mapping of recorder/permission outcomes to user-visible state. */
fun recordingFailureState(permissionGranted: Boolean, startError: Throwable? = null, stopError: Throwable? = null, fileBytes: Long? = null): RecordingFailure? = when {
    !permissionGranted -> RecordingFailure.PermissionDenied
    startError != null -> RecordingFailure.StartFailed
    stopError != null -> RecordingFailure.StopFailed
    fileBytes != null && fileBytes > Q_SERVANT_MAX_AUDIO_BYTES -> RecordingFailure.TooLarge
    fileBytes != null && fileBytes <= 0L -> RecordingFailure.EncodingFailed
    else -> null
}

/** Thin Android MediaRecorder wrapper. The caller owns runtime permission requests. */
class QServantRecorder(private val maxBytes: Long = Q_SERVANT_MAX_AUDIO_BYTES) {
    private var recorder: MediaRecorder? = null
    private var output: File? = null

    fun start(file: File, permissionGranted: Boolean): RecordingResult {
        if (!permissionGranted) return RecordingResult(failure = RecordingFailure.PermissionDenied, message = "Microphone permission denied")
        return try {
            file.parentFile?.mkdirs()
            val r = MediaRecorder()
            r.setAudioSource(MediaRecorder.AudioSource.MIC)
            r.setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
            r.setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
            r.setAudioEncodingBitRate(128_000)
            r.setAudioSamplingRate(44_100)
            r.setOutputFile(file.absolutePath)
            r.prepare()
            r.start()
            recorder = r
            output = file
            RecordingResult(file = file)
        } catch (t: Throwable) {
            runCatching { recorder?.release() }
            recorder = null
            runCatching { file.delete() }
            RecordingResult(failure = RecordingFailure.StartFailed, message = t.message)
        }
    }

    fun stop(): RecordingResult {
        val r = recorder ?: return RecordingResult(failure = RecordingFailure.StopFailed, message = "Not recording")
        val file = output
        recorder = null
        output = null
        return try {
            r.stop()
            r.reset(); r.release()
            if (file == null || !file.exists() || file.length() <= 0L) {
                file?.delete()
                RecordingResult(failure = RecordingFailure.EncodingFailed, message = "No encoded audio was produced")
            } else if (file.length() > maxBytes) {
                file.delete()
                RecordingResult(failure = RecordingFailure.TooLarge, message = "Audio exceeds 10 MiB")
            } else RecordingResult(file = file)
        } catch (t: Throwable) {
            runCatching { r.reset(); r.release() }
            file?.delete()
            RecordingResult(failure = RecordingFailure.StopFailed, message = t.message)
        }
    }

    fun cancel() {
        val r = recorder
        recorder = null
        output?.delete()
        output = null
        runCatching { r?.reset(); r?.release() }
    }
}
