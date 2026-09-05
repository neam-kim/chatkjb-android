package com.neamkim.chatkjb.core.navigation

/** Herdr agent UI bundled inside the ChatKJB APK. */
object HerdrRoute {
    const val embeddedUrl =
        "https://appassets.androidplatform.net/assets/herdr/index.html"

    /**
     * A relay setup credential arrives only in the URI fragment, which is never
     * sent over HTTP. Preserve that property when handing it to the embedded UI.
     */
    fun embeddedUrl(setupFragment: String?): String {
        val fragment = setupFragment
            ?.takeIf { it.length <= 4_096 && it.none(Char::isISOControl) }
            ?: return embeddedUrl
        return "$embeddedUrl#$fragment"
    }
}
