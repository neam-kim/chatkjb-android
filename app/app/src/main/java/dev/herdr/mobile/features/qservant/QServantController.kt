package dev.herdr.mobile.features.qservant

import dev.herdr.mobile.features.chat.net.CompanionClient
import dev.herdr.mobile.features.chat.net.QServantReport
import dev.herdr.mobile.features.chat.net.ServerFrame
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import java.util.concurrent.atomic.AtomicInteger

enum class QServantConnection { Disconnected, Connecting, Connected }
enum class QServantJobState { Idle, Recorded, Uploading, Transcribing, Queued, Running, Completed, Failed, Cancelled }
data class QServantState(
    val connection: QServantConnection = QServantConnection.Disconnected,
    val catalog: ServerFrame.QServantCatalogResult? = null,
    val selectedModel: String? = null,
    val selectedEffort: String? = null,
    val recording: Boolean = false,
    val jobId: String? = null,
    val transcript: String? = null,
    val jobState: QServantJobState = QServantJobState.Idle,
    val report: QServantReport? = null,
    val error: String? = null,
)

fun reduceQServant(state: QServantState, frame: ServerFrame): QServantState = when (frame) {
    is ServerFrame.QServantCatalogResult -> {
        val model = state.selectedModel?.takeIf { id -> frame.models.any { it.id == id } }
            ?: frame.defaultModel?.takeIf { id -> frame.models.any { it.id == id } }
            ?: frame.models.firstOrNull()?.id
        val modelInfo = frame.models.firstOrNull { it.id == model }
        val efforts = modelInfo?.efforts.orEmpty()
        val effort = state.selectedEffort?.takeIf { it in efforts }
            ?: frame.defaultEffort?.takeIf { model == frame.defaultModel && it in efforts }
            ?: modelInfo?.defaultEffort?.takeIf { it in efforts }
            ?: efforts.firstOrNull()
        state.copy(catalog = frame, selectedModel = model, selectedEffort = effort)
    }
    is ServerFrame.QServantJobFrame -> state.copy(jobId = frame.job.jobId.takeIf { it.isNotBlank() } ?: state.jobId, jobState = frame.job.state.toJobState(), transcript = frame.job.transcript ?: state.transcript, report = frame.job.report ?: state.report, error = frame.job.error)
    is ServerFrame.QServantError -> state.copy(error = frame.message, jobState = if (state.jobId != null || frame.jobId != null) QServantJobState.Failed else state.jobState)
    else -> state
}

private fun String.toJobState() = when (lowercase()) {
    "recorded" -> QServantJobState.Recorded; "uploading" -> QServantJobState.Uploading; "transcribing" -> QServantJobState.Transcribing; "queued" -> QServantJobState.Queued; "running" -> QServantJobState.Running; "completed", "complete", "success" -> QServantJobState.Completed; "cancelled", "canceled" -> QServantJobState.Cancelled; "failed", "error" -> QServantJobState.Failed; else -> QServantJobState.Idle
}

class QServantController(private val client: CompanionClient? = null) {
    private val reqSeq = AtomicInteger()
    private val _state = MutableStateFlow(QServantState())
    val state: StateFlow<QServantState> = _state.asStateFlow()
    fun connection(connected: Boolean) { _state.update { it.copy(connection = if (connected) QServantConnection.Connected else QServantConnection.Disconnected) } }
    fun setRecording(value: Boolean) { _state.update { it.copy(recording = value, jobState = if (value) QServantJobState.Idle else it.jobState, error = null) } }
    fun selectModel(model: String?) { _state.update {
        val modelInfo = it.catalog?.models?.firstOrNull { entry -> entry.id == model }
        val efforts = modelInfo?.efforts.orEmpty()
        val effort = it.selectedEffort?.takeIf { e -> e in efforts }
            ?: modelInfo?.defaultEffort?.takeIf { e -> e in efforts }
            ?: efforts.firstOrNull()
        it.copy(selectedModel = model, selectedEffort = effort)
    } }
    fun selectEffort(effort: String?) { _state.update { it.copy(selectedEffort = effort?.takeIf { e -> e in compatibleEfforts(it.catalog?.models.orEmpty(), it.selectedModel) }) } }
    fun onFrame(frame: ServerFrame) { _state.update { reduceQServant(it, frame) } }
    fun sendCatalog() { client?.send(QServantClientFrame.catalog("q${reqSeq.incrementAndGet()}")) }
    fun sendSubmit(audioBytes: ByteArray): Boolean { val s = state.value; val model = s.selectedModel ?: return false; client?.send(QServantClientFrame.submit("q${reqSeq.incrementAndGet()}", model, s.selectedEffort, audioBytes)); _state.update { it.copy(jobState = QServantJobState.Uploading, error = null) }; return client != null }
    fun sendStatus(jobId: String) { client?.send(QServantClientFrame.status("q${reqSeq.incrementAndGet()}", jobId)) }
    fun sendCancel(jobId: String) { client?.send(QServantClientFrame.cancel("q${reqSeq.incrementAndGet()}", jobId)) }
}
