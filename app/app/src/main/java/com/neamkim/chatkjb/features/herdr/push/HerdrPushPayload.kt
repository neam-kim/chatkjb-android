package com.neamkim.chatkjb.features.herdr.push

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class HerdrPushPayload(
    val kind: String,
    val paneId: String = "",
    val workspaceId: String = "",
    val title: String = "",
    val body: String = "",
)

private val pushJson = Json { ignoreUnknownKeys = true }

fun parseHerdrPush(bytes: ByteArray): HerdrPushPayload? = runCatching {
    pushJson.decodeFromString<HerdrPushPayload>(bytes.decodeToString())
}.getOrNull()
