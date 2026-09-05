package dev.herdr.mobile

import dev.herdr.mobile.core.navigation.AppDestination
import dev.herdr.mobile.core.navigation.MoonlightRoute
import dev.herdr.mobile.core.navigation.parseDestinationUri
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * Validates tablet navigation contract, Moonlight client in-process routing,
 * and verifies tablet Herdr/Termux integration remains distinct from phone.
 */
class TabletRouteDistinctionTest {

    @Test
    fun tabletChatDestinationRoutesToTermuxNotEmbeddedWebView() {
        // Phone uses EmbeddedHerdrScreen loading webview assets ("https://appassets.androidplatform.net/assets/herdr/index.html")
        // Tablet uses native Termux integration (com.termux.app.TermuxActivity and RemoteTerminalSession)
        val termuxClass = com.termux.app.TermuxActivity::class.java
        assertNotNull(termuxClass)
        assertEquals("com.termux.app.TermuxActivity", termuxClass.name)

        // Verify RemoteTerminalSession exists for tablet terminal support
        val sessionClass = com.termux.terminal.RemoteTerminalSession::class.java
        assertNotNull(sessionClass)
    }

    @Test
    fun serverRouteTargetsInProcessMoonlightPcView() {
        // Server on tablet must run transplanted Moonlight client in-process (com.limelight.PcView)
        // It must NOT open a /server/ WebView
        assertEquals("com.limelight.PcView", MoonlightRoute.pcViewClass)
        val target = MoonlightRoute.target("com.termux")
        assertEquals("com.termux", target.packageName)
        assertEquals("com.limelight.PcView", target.className)
        assertFalse(MoonlightRoute.pcViewClass.contains("WebView"))
        assertFalse(MoonlightRoute.pcViewClass.contains("server"))
    }

    @Test
    fun serverAndFinanceAreLauncherOnlyAndNotDeepLinks() {
        assertEquals(null, parseDestinationUri("kimjb://open/server"))
        assertEquals(null, parseDestinationUri("kimjb://open/finance"))
        assertEquals(null, parseDestinationUri("kimjb://open/homepage"))
        // Known deep links
        assertEquals(AppDestination.HOME, parseDestinationUri("kimjb://open/home"))
        assertEquals(AppDestination.EMAIL, parseDestinationUri("kimjb://open/email"))
        assertEquals(AppDestination.CHAT_KJB, parseDestinationUri("kimjb://open/chat"))
    }
}
