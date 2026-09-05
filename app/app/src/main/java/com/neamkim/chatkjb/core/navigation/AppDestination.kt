package com.neamkim.chatkjb.core.navigation

import android.content.Intent

/**
 * Top-level destinations owned by the unified ChatKJB app.
 *
 * [HOMEPAGE] and [FINANCE] are reachable only from the launcher,
 * never from the deep-link contract.
 */
enum class AppDestination {
    HOME,
    HOMEPAGE,
    FINANCE,
    EMAIL,
    CHAT_KJB,
    CONSOLE_SETTINGS,
    AUTOBOT,
    SERVER,
    MOONLIGHT,
}

/**
 * Resolve the app's own deep-link contract.
 *
 * `kimjb://open/{home,email,chat,notifications}` is the current contract. `kjbmail://open` is the legacy
 * entry point kimjb.com still emits; it predates this app and always meant "open mail".
 */
private fun resolveDestination(scheme: String?, host: String?, path: String?): AppDestination? {
    if (!host.equals("open", ignoreCase = true)) return null

    if (scheme.equals("kjbmail", ignoreCase = true)) return AppDestination.EMAIL
    if (!scheme.equals("kimjb", ignoreCase = true)) return null

    return when (path?.trim('/')?.lowercase()) {
        "home" -> AppDestination.HOME
        "email" -> AppDestination.EMAIL
        "chat" -> AppDestination.CHAT_KJB
        "notifications" -> AppDestination.CONSOLE_SETTINGS
        "server" -> AppDestination.MOONLIGHT
        else -> null
    }
}

/** Parse only the app's own optional deep-link contract. */
fun parseDestinationIntent(intent: Intent?): AppDestination? {
    val data = intent?.data ?: return null
    return resolveDestination(data.scheme, data.host, data.path)
}

/** Pure URI parser for routing tests and other non-Android callers. */
fun parseDestinationUri(uri: String?): AppDestination? = runCatching {
    val data = java.net.URI(uri ?: return null)
    resolveDestination(data.scheme, data.host, data.path)
}.getOrNull()
