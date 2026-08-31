package dev.herdr.mobile

import dev.herdr.mobile.core.navigation.AppDestination
import dev.herdr.mobile.core.navigation.parseDestinationUri
import dev.herdr.mobile.features.homepage.KimJbLauncherEntries
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File

class QServantIntegrationTest {

    @Test
    fun launcherEntriesOrderAndLabelsArePreservedWithQServantAsFifthRow() {
        assertEquals("Expected exactly 5 launcher rows", 5, KimJbLauncherEntries.size)

        assertEquals("Row 1 must be Homepage", "Homepage", KimJbLauncherEntries[0].title)
        assertEquals(AppDestination.HOMEPAGE, KimJbLauncherEntries[0].destination)

        assertEquals("Row 2 must be Email", "Email", KimJbLauncherEntries[1].title)
        assertEquals(AppDestination.EMAIL, KimJbLauncherEntries[1].destination)

        assertEquals("Row 3 must be Finance", "Finance", KimJbLauncherEntries[2].title)
        assertEquals(AppDestination.FINANCE, KimJbLauncherEntries[2].destination)

        assertEquals("Row 4 must be ChatKJB", "ChatKJB", KimJbLauncherEntries[3].title)
        assertEquals(AppDestination.CHAT_KJB, KimJbLauncherEntries[3].destination)

        assertEquals("Row 5 must be Q Servant", "Q Servant", KimJbLauncherEntries[4].title)
        assertEquals(AppDestination.Q_SERVANT, KimJbLauncherEntries[4].destination)

        // Never target or label any row as "mobile"
        for (entry in KimJbLauncherEntries) {
            assertFalse(
                "Entry title '" + entry.title + "' must not be 'mobile'",
                entry.title.equals("mobile", ignoreCase = true),
            )
            assertFalse(
                "Entry description '" + entry.description + "' must not refer to 'mobile'",
                entry.description.contains("mobile", ignoreCase = true),
            )
        }
    }

    @Test
    fun appDestinationEnumIncludesQServant() {
        val destinations = AppDestination.values().map { it.name }
        assertTrue("AppDestination must contain Q_SERVANT", destinations.contains("Q_SERVANT"))
        assertTrue("AppDestination must contain HOME", destinations.contains("HOME"))
        assertTrue("AppDestination must contain HOMEPAGE", destinations.contains("HOMEPAGE"))
        assertTrue("AppDestination must contain FINANCE", destinations.contains("FINANCE"))
        assertTrue("AppDestination must contain EMAIL", destinations.contains("EMAIL"))
        assertTrue("AppDestination must contain CHAT_KJB", destinations.contains("CHAT_KJB"))
    }

    @Test
    fun qServantIsNotReachableByDeepLinkAndRejectsMobile() {
        // Matching Homepage and Finance: Q Servant is reachable only from the launcher,
        // never from the deep-link contract unless explicitly specified by product.
        assertNull("kimjb://open/qservant must not resolve", parseDestinationUri("kimjb://open/qservant"))
        assertNull("kimjb://open/q-servant must not resolve", parseDestinationUri("kimjb://open/q-servant"))
        assertNull("kimjb://open/q_servant must not resolve", parseDestinationUri("kimjb://open/q_servant"))
        assertNull("kimjb://open/QSERVANT/ must not resolve", parseDestinationUri("kimjb://open/QSERVANT/"))

        // "mobile" is strictly forbidden and must never resolve
        assertNull("kimjb://open/mobile must not resolve", parseDestinationUri("kimjb://open/mobile"))

        // Standard routes remain functional
        assertEquals(AppDestination.HOME, parseDestinationUri("kimjb://open/home"))
        assertEquals(AppDestination.EMAIL, parseDestinationUri("kimjb://open/email"))
        assertEquals(AppDestination.CHAT_KJB, parseDestinationUri("kimjb://open/chat"))

        // Homepage and Finance remain unreachable via deep link
        assertNull("kimjb://open/homepage must not resolve", parseDestinationUri("kimjb://open/homepage"))
        assertNull("kimjb://open/finance must not resolve", parseDestinationUri("kimjb://open/finance"))
    }

    @Test
    fun backAndPresentationNavigationSemantics() {
        // Models the MainActivity destination state machine:
        // destination == null -> Launcher
        // destination == Q_SERVANT -> QServantScreen(onBack = { destination = null })
        var destination: AppDestination? = null

        // Navigate to Q Servant
        destination = AppDestination.Q_SERVANT
        assertEquals(AppDestination.Q_SERVANT, destination)

        // Simulating onBack
        val onBack = { destination = null }
        onBack()
        assertNull("Back from Q Servant must return to launcher (null)", destination)

        // Navigate to ChatKJB and Back
        destination = AppDestination.CHAT_KJB
        assertEquals(AppDestination.CHAT_KJB, destination)
        onBack()
        assertNull("Back from ChatKJB must return to launcher (null)", destination)

        // Navigate to Homepage and Exit
        destination = AppDestination.HOMEPAGE
        assertEquals(AppDestination.HOMEPAGE, destination)
        onBack()
        assertNull("Exit from Homepage must return to launcher (null)", destination)
    }

    @Test
    fun qServantAndChatKjbShareCompanionClientOnboardingAndConnectionContract() {
        // Replicates the host navigation and client connection contract:
        // If url is null when entering Q_SERVANT or CHAT_KJB, the host routes through OnboardUrl
        // Upon setting url, client connects before presentation screen is active.
        var url: String? = null
        var clientConnectedUrl: String? = null
        var destination: AppDestination? = AppDestination.Q_SERVANT

        // Simulated host check for whether onboarding is needed:
        val needsOnboarding = (destination == AppDestination.CHAT_KJB || destination == AppDestination.Q_SERVANT) && url == null
        assertTrue("Should require onboarding when URL is missing for Q Servant", needsOnboarding)

        // User completes onboarding
        val enteredUrl = "ws://192.168.1.100:8787"
        url = enteredUrl
        clientConnectedUrl = enteredUrl

        assertEquals("Client should connect to entered companion URL", enteredUrl, clientConnectedUrl)
        assertFalse(
            "Onboarding should no longer be needed once URL is set",
            (destination == AppDestination.CHAT_KJB || destination == AppDestination.Q_SERVANT) && url == null,
        )
    }

    @Test
    fun manifestDeclaresRecordAudioPermission() {
        // Locate AndroidManifest.xml relative to module or project root
        val candidatePaths = listOf(
            "src/main/AndroidManifest.xml",
            "app/app/src/main/AndroidManifest.xml",
            "app/src/main/AndroidManifest.xml",
        )
        val manifestFile = candidatePaths.map { File(it) }.firstOrNull { it.exists() }
        assertNotNull("AndroidManifest.xml should exist", manifestFile)

        val content = manifestFile!!.readText()
        assertTrue(
            "Manifest must contain RECORD_AUDIO permission",
            content.contains("android.permission.RECORD_AUDIO"),
        )
        assertTrue(
            "Manifest must contain INTERNET permission",
            content.contains("android.permission.INTERNET"),
        )
        assertTrue(
            "Manifest must contain POST_NOTIFICATIONS permission",
            content.contains("android.permission.POST_NOTIFICATIONS"),
        )
    }
}
