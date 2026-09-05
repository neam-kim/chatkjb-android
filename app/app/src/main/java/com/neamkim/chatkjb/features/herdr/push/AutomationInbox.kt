package com.neamkim.chatkjb.features.herdr.push

import android.content.Context
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

@Serializable
data class AutomationInboxItem(
    val kind: String,
    val title: String,
    val body: String,
    val updatedAt: Long,
)

object AutomationInbox {
    private const val PREFS = "automation_inbox"
    private const val SENTINEL = "sentinel"
    private const val SKILL = "skill"
    private val json = Json { ignoreUnknownKeys = true }

    fun update(context: Context, payload: HerdrPushPayload) {
        when {
            payload.kind == "sentinel-clear" -> remove(context, SENTINEL)
            payload.kind == "skill-clear" -> remove(context, SKILL)
            payload.kind.startsWith("sentinel-") -> put(context, SENTINEL, payload)
            payload.kind.startsWith("skill-") -> put(context, SKILL, payload)
        }
    }

    fun sentinel(context: Context): AutomationInboxItem? = get(context, SENTINEL)

    fun skill(context: Context): AutomationInboxItem? = get(context, SKILL)

    fun preferences(context: Context) = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)

    private fun put(context: Context, key: String, payload: HerdrPushPayload) {
        val item = AutomationInboxItem(
            kind = payload.kind,
            title = payload.title,
            body = payload.body,
            updatedAt = System.currentTimeMillis(),
        )
        preferences(context).edit().putString(key, json.encodeToString(item)).apply()
    }

    private fun remove(context: Context, key: String) {
        preferences(context).edit().remove(key).apply()
    }

    private fun get(context: Context, key: String): AutomationInboxItem? =
        preferences(context).getString(key, null)?.let { raw ->
            runCatching { json.decodeFromString<AutomationInboxItem>(raw) }.getOrNull()
        }
}
