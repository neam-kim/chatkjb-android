package dev.herdr.mobile.features.qservant

import dev.herdr.mobile.features.chat.net.QServantModel
import dev.herdr.mobile.features.chat.net.ServerFrame
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import java.util.Base64

typealias QServantCatalog = ServerFrame.QServantCatalogResult

fun compatibleEfforts(models: List<QServantModel>, modelId: String?): List<String> = models.firstOrNull { it.id == modelId }?.efforts.orEmpty()
fun compatibleEfforts(catalog: ServerFrame.QServantCatalogResult, modelId: String?): List<String> = compatibleEfforts(catalog.models, modelId)

object QServantClientFrame {
    fun catalog(reqId: String) = buildJsonObject { put("t", "qservant_catalog"); put("reqId", reqId) }.toString()
    fun submit(reqId: String, model: String, effort: String?, audioBytes: ByteArray): String {
        require(audioBytes.size <= 10 * 1024 * 1024) { "audio exceeds 10 MiB" }
        return buildJsonObject {
            put("t", "qservant_submit"); put("reqId", reqId); put("model", model)
            // The v8 contract keeps `effort` present; an empty string asks the
            // provider to use its default when the selected model exposes none.
            put("effort", effort ?: "")
            put("audioMime", "audio/mp4"); put("audioBase64", Base64.getEncoder().encodeToString(audioBytes))
        }.toString()
    }
    fun status(reqId: String, jobId: String) = buildJsonObject { put("t", "qservant_status"); put("reqId", reqId); put("jobId", jobId) }.toString()
    fun cancel(reqId: String, jobId: String) = buildJsonObject { put("t", "qservant_cancel"); put("reqId", reqId); put("jobId", jobId) }.toString()
}

class QServantAudioException(message: String) : IllegalArgumentException(message)
fun decodeQServantAudio(mime: String, base64: String): ByteArray {
    if (mime.trim().lowercase() != "audio/mp4") throw QServantAudioException("audio MIME must be audio/mp4")
    val bytes = runCatching { Base64.getDecoder().decode(base64) }.getOrElse { throw QServantAudioException("invalid audio base64") }
    if (bytes.size > 10 * 1024 * 1024) throw QServantAudioException("audio exceeds 10 MiB")
    return bytes
}
fun parseQServantFrame(text: String): ServerFrame = dev.herdr.mobile.features.chat.net.parseServerFrame(text)
