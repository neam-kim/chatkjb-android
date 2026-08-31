package dev.herdr.mobile.features.qservant

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import dev.herdr.mobile.features.chat.net.CompanionClient
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.filter
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import java.io.File

val LocalQServantClient = staticCompositionLocalOf<CompanionClient?> { null }

private val QBgTop = Color(0xFF0B1626)
private val QBgBottom = Color(0xFF07101B)
private val QText = Color(0xFFEEF5FF)
private val QMuted = Color(0xFF9FB0C8)
private val QAccent = Color(0xFF79A7FF)
private val QGreen = Color(0xFF62D59B)
private val QRed = Color(0xFFFF7F8B)

/** Deliberately small Q Servant surface; navigation only needs to provide onBack. */
@Composable
fun QServantScreen(onBack: () -> Unit) {
    val client = LocalQServantClient.current ?: error("QServantScreen requires the shared CompanionClient")
    QServantScreen(client = client, onBack = onBack)
}

/** Injection overload for hosts that already own a connected CompanionClient. */
@Composable
fun QServantScreen(client: CompanionClient, onBack: () -> Unit) {
    val context = LocalContext.current
    val controller = remember { QServantController(client) }
    val state by controller.state.collectAsState()
    val recorder = remember { QServantRecorder() }
	DisposableEffect(recorder) {
		onDispose { recorder.cancel() }
	}
    var pendingPermission by remember { mutableStateOf(false) }
    var recordingMessage by remember { mutableStateOf<String?>(null) }
    var reportText by remember { mutableStateOf("") }
    LaunchedEffect(client) {
        launch {
            client.connected.collect { controller.connection(it) }
        }
        launch { client.connected.filter { it }.first(); controller.sendCatalog() }
        client.frames.collect { frame -> controller.onFrame(frame) }
    }
    val permissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { granted ->
        if (granted && pendingPermission) {
            val file = File(context.cacheDir, "qservant-${System.currentTimeMillis()}.m4a")
            val result = recorder.start(file, true)
            controller.setRecording(result.ok)
            recordingMessage = result.failure?.name
        } else if (!granted) {
            controller.setRecording(false)
            recordingMessage = RecordingFailure.PermissionDenied.name
        }
        pendingPermission = false
    }

	BackHandler(onBack = onBack)
	val model = state.selectedModel
	val catalog = state.catalog
	val efforts = compatibleEfforts(catalog?.models.orEmpty(), model)
	val selectedModelInfo = catalog?.models?.firstOrNull { it.id == model }
	val reportSummary = buildString {
		state.transcript?.let { append("REQUEST / TRANSCRIPT\n$it\n\n") }
		state.report?.let {
			append("PERFORMED WORK\n${it.work}\n\n")
			append("VERIFICATION / TESTS\n${it.verification}\n\n")
			append("CHANGED FILES\n${it.changes.joinToString("\n")}\n\n")
			it.commit?.takeIf { value -> value.isNotBlank() }?.let { value -> append("COMMIT / PR\n$value\n\n") }
			append("FINAL RESULT\n${it.result}\nSuccess: ${it.success}")
		}
	}
	val visibleReport = when {
		!state.error.isNullOrBlank() -> state.error
		reportSummary.isNotBlank() -> reportSummary
		reportText.isNotBlank() -> reportText
		else -> "아직 작업 보고가 없습니다.\n\n음성 명령을 보내면 요청, 수행 내용, 검증, 변경 파일과 최종 결과만 여기에 표시됩니다."
	}.orEmpty()
	val onMicClick: () -> Unit = {
		if (state.recording) {
			val result = recorder.stop()
			controller.setRecording(false)
			if (result.ok && result.file != null && !model.isNullOrBlank()) {
				val bytes = runCatching { result.file.readBytes() }.getOrNull()
				if (bytes == null) reportText = "Recording failed: encoding could not be read"
				else runCatching { controller.sendSubmit(bytes) }
					.onSuccess { sent -> reportText = if (sent) "Submitted ${bytes.size} bytes" else "Submit unavailable" }
					.onFailure { reportText = "Submit failed: ${it.message}" }
			} else reportText = result.file?.let { "Select a model before submitting" } ?: "Recording failed: ${result.failure}"
			result.file?.delete()
		} else {
			val granted = ContextCompat.checkSelfPermission(context, Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED
			if (granted) {
				val file = File(context.cacheDir, "qservant-${System.currentTimeMillis()}.m4a")
				val result = recorder.start(file, true)
				controller.setRecording(result.ok)
				recordingMessage = result.failure?.name
			} else {
				pendingPermission = true
				permissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
			}
		}
	}
	var modelMenu by remember { mutableStateOf(false) }
	var effortMenu by remember { mutableStateOf(false) }
	Box(
		Modifier.fillMaxSize().background(Brush.verticalGradient(listOf(QBgTop, QBgBottom))),
	) {
		Column(Modifier.fillMaxSize().safeDrawingPadding().padding(horizontal = 16.dp, vertical = 12.dp)) {
			Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
				TextButton(onClick = onBack, contentPadding = PaddingValues(0.dp), modifier = Modifier.size(36.dp)) {
					Text("‹", color = QAccent, style = MaterialTheme.typography.headlineMedium)
				}
				Box(Modifier.size(8.dp).background(if (state.connection == QServantConnection.Connected) QGreen else QRed, CircleShape))
				Spacer(Modifier.width(7.dp))
				Text(
					if (state.connection == QServantConnection.Connected) "Mac mini connected" else "Mac mini disconnected",
					color = QMuted,
					style = MaterialTheme.typography.labelMedium,
				)
				Spacer(Modifier.weight(1f))
				Text("Q SERVANT", color = QMuted, style = MaterialTheme.typography.labelSmall, fontWeight = FontWeight.Bold)
			}
			Spacer(Modifier.height(10.dp))
			Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(9.dp)) {
				Box(Modifier.weight(1f)) {
					SelectorCard(
						label = "MODEL",
						value = selectedModelInfo?.label?.ifBlank { selectedModelInfo.id } ?: "Select model",
						detail = selectedModelInfo?.quota?.label,
						enabled = catalog?.models?.isNotEmpty() == true,
						onClick = { modelMenu = true },
					)
					DropdownMenu(expanded = modelMenu, onDismissRequest = { modelMenu = false }) {
						catalog?.models.orEmpty().forEach { item ->
							val itemLabel = listOfNotNull(item.label.ifBlank { item.id }, item.quota?.label?.takeIf(String::isNotBlank)).joinToString(" · ")
							DropdownMenuItem(text = { Text(itemLabel) }, onClick = { controller.selectModel(item.id); modelMenu = false })
						}
					}
				}
				Box(Modifier.weight(1f)) {
					SelectorCard(
						label = "EFFORT",
						value = state.selectedEffort?.replaceFirstChar { it.uppercase() } ?: "Provider default",
						detail = if (efforts.isEmpty()) "Model managed" else "${efforts.size} options",
						enabled = efforts.isNotEmpty(),
						onClick = { effortMenu = true },
					)
					DropdownMenu(expanded = effortMenu, onDismissRequest = { effortMenu = false }) {
						efforts.forEach { effort ->
							DropdownMenuItem(text = { Text(effort.replaceFirstChar { it.uppercase() }) }, onClick = { controller.selectEffort(effort); effortMenu = false })
						}
					}
				}
			}
			Spacer(Modifier.height(12.dp))
			Column(
				Modifier.fillMaxWidth().height(218.dp).background(
					Brush.verticalGradient(listOf(Color(0xFF12243E), Color(0xFF0D1829))),
					RoundedCornerShape(18.dp),
				).border(1.dp, Color(0xFF365A86), RoundedCornerShape(18.dp)),
				horizontalAlignment = Alignment.CenterHorizontally,
				verticalArrangement = Arrangement.Center,
			) {
				Button(
					onClick = onMicClick,
					shape = CircleShape,
					colors = ButtonDefaults.buttonColors(containerColor = if (state.recording) Color(0xFF6B2634) else Color(0xFF1B3153)),
					border = BorderStroke(1.dp, if (state.recording) QRed else Color(0xFF5E8FD0)),
					contentPadding = PaddingValues(0.dp),
					modifier = Modifier.size(96.dp),
				) { Text(if (state.recording) "■" else "🎙", color = QText, style = MaterialTheme.typography.headlineLarge) }
				Spacer(Modifier.height(13.dp))
				Text(if (state.recording) "눌러서 전송" else "누르고 말하기", color = QText, fontWeight = FontWeight.Bold)
				Text(
					if (state.recording) "녹음 중 · 다시 누르면 Mac mini로 전송" else "음성은 Mac mini에서 로컬 처리됩니다",
					color = QMuted,
					style = MaterialTheme.typography.labelSmall,
				)
			}
			recordingMessage?.let {
				Text("Recording: $it", color = QRed, style = MaterialTheme.typography.labelSmall, modifier = Modifier.padding(top = 4.dp))
			}
			Spacer(Modifier.height(12.dp))
			Column(
				Modifier.fillMaxWidth().weight(1f).background(Color(0xFF09121E), RoundedCornerShape(16.dp))
					.border(1.dp, Color(0xFF2B425F), RoundedCornerShape(16.dp)).padding(14.dp),
			) {
				Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
					Text("작업 보고", color = QText, fontWeight = FontWeight.Bold)
					Spacer(Modifier.weight(1f))
					JobBadge(state.jobState)
				}
				Spacer(Modifier.height(10.dp))
				Text(
					visibleReport,
					color = if (state.error.isNullOrBlank()) Color(0xFFD9E4F2) else QRed,
					style = MaterialTheme.typography.bodySmall,
					modifier = Modifier.fillMaxWidth().weight(1f).verticalScroll(rememberScrollState()),
				)
				Text("전체 PTY 로그는 Mac mini에만 유지됩니다", color = QMuted, style = MaterialTheme.typography.labelSmall)
			}
		}
	}
}

@Composable
private fun SelectorCard(label: String, value: String, detail: String?, enabled: Boolean, onClick: () -> Unit) {
	Column(
		Modifier.fillMaxWidth().height(78.dp).background(Color(0xFF0E1621), RoundedCornerShape(12.dp))
			.border(1.dp, Color(0xFF334760), RoundedCornerShape(12.dp))
			.clickable(enabled = enabled, onClick = onClick).padding(horizontal = 11.dp, vertical = 9.dp),
	) {
		Text(label, color = QMuted, style = MaterialTheme.typography.labelSmall)
		Row(verticalAlignment = Alignment.CenterVertically) {
			Text(value, color = if (enabled) QText else QMuted, fontWeight = FontWeight.Bold, maxLines = 1, overflow = TextOverflow.Ellipsis, modifier = Modifier.weight(1f))
			if (enabled) Text("⌄", color = QAccent)
		}
		if (!detail.isNullOrBlank()) Text(detail, color = QMuted, style = MaterialTheme.typography.labelSmall, maxLines = 1, overflow = TextOverflow.Ellipsis)
	}
}

@Composable
private fun JobBadge(state: QServantJobState) {
	val terminal = state == QServantJobState.Completed || state == QServantJobState.Failed || state == QServantJobState.Cancelled
	val color = when (state) {
		QServantJobState.Failed, QServantJobState.Cancelled -> QRed
		QServantJobState.Completed -> QGreen
		else -> QAccent
	}
	Box(Modifier.background(color.copy(alpha = .14f), CircleShape).border(1.dp, color.copy(alpha = .45f), CircleShape).padding(horizontal = 9.dp, vertical = 5.dp)) {
		Text(if (terminal) state.name.uppercase() else if (state == QServantJobState.Idle) "READY" else state.name.uppercase(), color = color, style = MaterialTheme.typography.labelSmall)
	}
}
