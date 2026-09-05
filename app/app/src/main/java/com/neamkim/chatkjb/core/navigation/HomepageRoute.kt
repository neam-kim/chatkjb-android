package com.neamkim.chatkjb.core.navigation

import android.net.Uri

/** Fixed web surface; never accept an arbitrary URL from an intent or UI. */
object HomepageRoute {
    /** Materialized only on Android; pure policy tests use [canonicalUrl]. */
    val uri: Uri by lazy { Uri.parse(canonicalUrl) }
    const val canonicalUrl = "https://kimjb.com/"

    /**
     * Admin-only asset view on the same origin.
     *
     * The site deliberately stopped listing this under Information, so the app
     * launcher is now the entry point. The server still gates it behind the
     * admin session, and an unauthenticated visit lands on the sign-in page that
     * this WebView already knows how to complete.
     */
    const val financeUrl = "https://kimjb.com/finance/"

    fun isAllowed(candidate: Uri): Boolean =
        isAllowed(candidate.toString())

    /** Pure validation helper so URL policy can be tested without Android shadows. */
    fun isAllowed(candidate: String): Boolean = runCatching {
        val parsed = java.net.URI(candidate)
        parsed.scheme.equals("https", ignoreCase = true) &&
            parsed.host.equals("kimjb.com", ignoreCase = true) &&
            (parsed.port == -1 || parsed.port == 443)
    }.getOrDefault(false)

    /**
     * Hosts the Google sign-in leg navigates through.
     *
     * These must stay inside this WebView. Handing them to the system browser
     * completes the OAuth round trip in *that* browser's cookie jar, so the session
     * cookie never reaches the app and the user appears logged out on every visit.
     * Worse, a leg that starts here and finishes there arrives without the matching
     * cookies and Google answers 400.
     *
     * Google redirects the sign-in leg to a country-local domain part way through
     * (observed: accounts.google.co.jp), so the accounts host is matched by pattern
     * rather than by a fixed list. The TLD shape is deliberately narrow — a bare
     * two-letter ccTLD, or co./com. plus one — because those are registry-controlled
     * and cannot be bought out from under Google the way a vanity gTLD could.
     */
    private val accountsHost =
        Regex("""^accounts\.google\.(com|[a-z]{2}|(co|com)\.[a-z]{2})$""")

    private val signInHosts = setOf("accounts.youtube.com")

    fun isSignInHost(candidate: Uri): Boolean = isSignInHost(candidate.toString())

    fun isSignInHost(candidate: String): Boolean = runCatching {
        val parsed = java.net.URI(candidate)
        val host = parsed.host?.lowercase() ?: return@runCatching false
        parsed.scheme.equals("https", ignoreCase = true) &&
            (host in signInHosts || accountsHost.matches(host)) &&
            (parsed.port == -1 || parsed.port == 443)
    }.getOrDefault(false)

    /** Everything this WebView may load: the site itself plus the sign-in round trip. */
    fun isInAppNavigation(candidate: Uri): Boolean = isInAppNavigation(candidate.toString())

    fun isInAppNavigation(candidate: String): Boolean =
        isAllowed(candidate) || isSignInHost(candidate)
}
