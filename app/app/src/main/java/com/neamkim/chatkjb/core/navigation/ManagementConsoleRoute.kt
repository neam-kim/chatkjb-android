package com.neamkim.chatkjb.core.navigation

/** Fixed tailnet-only management consoles exposed beside the Herdr relay. */
object ManagementConsoleRoute {
    private const val origin = "https://neam-macmini.taild81d38.ts.net:8443"
    const val autoBotUrl = "$origin/autobot/"
    const val serverUrl = "$origin/server/"

    private val allowedRoots = setOf("/autobot/", "/server/")

    /**
     * Keep each console WebView inside its selected path on the existing Herdr
     * tailnet endpoint. A console cannot navigate into the relay or its sibling.
     */
    fun isAllowed(candidate: String, entryUrl: String): Boolean = runCatching {
        if (entryUrl != autoBotUrl && entryUrl != serverUrl) return@runCatching false
        val rawCandidate = java.net.URI(candidate)
        val candidateUri = java.net.URI(rawCandidate.scheme, rawCandidate.authority,
            rawCandidate.path, rawCandidate.query, rawCandidate.fragment).normalize()
        val entryUri = java.net.URI(entryUrl)
        val entryRoot = entryUri.path.takeIf { it in allowedRoots } ?: return@runCatching false

        candidateUri.scheme.equals("https", ignoreCase = true) &&
            candidateUri.host.equals(entryUri.host, ignoreCase = true) &&
            candidateUri.port == entryUri.port &&
            candidateUri.userInfo == null &&
            candidateUri.path.startsWith(entryRoot)
    }.getOrDefault(false)
}
