package com.neamkim.chatkjb

import com.neamkim.chatkjb.core.navigation.ManagementConsoleRoute
import com.neamkim.chatkjb.core.navigation.parseDestinationUri
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ManagementConsoleRouteTest {
    @Test fun consolesAllowTheirOwnRootAndNestedPages() {
        for (entry in listOf(ManagementConsoleRoute.autoBotUrl, ManagementConsoleRoute.serverUrl)) {
            assertTrue(ManagementConsoleRoute.isAllowed(entry, entry))
            assertTrue(ManagementConsoleRoute.isAllowed(entry + "details?tab=logs#latest", entry))
        }
    }

    @Test fun consolesRejectSiblingTraversalAndForeignOrigins() {
        val entry = ManagementConsoleRoute.autoBotUrl
        for (candidate in listOf(
            ManagementConsoleRoute.serverUrl,
            entry + "../server/",
            entry + "%2e%2e/server/",
            entry.replace("https:", "http:"),
            entry.replace(":8443", ":443"),
            entry.replace("neam-macmini", "attacker"),
            entry.replace("https://", "https://user@"),
            "file:///autobot/",
        )) assertFalse(candidate, ManagementConsoleRoute.isAllowed(candidate, entry))
        assertFalse(ManagementConsoleRoute.isAllowed("https://example.com/autobot/", "https://example.com/autobot/"))
    }

    @Test fun managementScreensCannotBeOpenedByPublicDeepLinks() {
        for (path in listOf("command_center", "autobot", "autobot_console", "server_console")) {
            assertNull(parseDestinationUri("kimjb://open/$path"))
        }
    }
}
