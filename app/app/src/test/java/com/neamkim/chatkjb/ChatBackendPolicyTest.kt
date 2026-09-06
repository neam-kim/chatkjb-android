package com.neamkim.chatkjb

import com.neamkim.chatkjb.core.navigation.ChatBackend
import com.neamkim.chatkjb.core.navigation.ChatBackendPolicy
import com.neamkim.chatkjb.core.navigation.DeviceClass
import com.neamkim.chatkjb.core.navigation.restoreSessionDeviceClass
import com.neamkim.chatkjb.core.navigation.TermuxRoute
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class ChatBackendPolicyTest {
    @Test
    fun deviceClassUsesTheConfirmed600DpBoundary() {
        assertEquals(DeviceClass.PHONE, ChatBackendPolicy.deviceClassFor(599))
        assertEquals(DeviceClass.TABLET, ChatBackendPolicy.deviceClassFor(600))
        assertEquals(DeviceClass.TABLET, ChatBackendPolicy.deviceClassFor(601))
        assertEquals(600, ChatBackendPolicy.TABLET_MIN_SHORTEST_EDGE_DP)
    }

    @Test
    fun activeDisplayBoundsClassifyFoldedAndUnfoldedSurfaces() {
        assertEquals(
            DeviceClass.PHONE,
            ChatBackendPolicy.deviceClassFor(
                ChatBackendPolicy.maximumShortestEdgeDp(1080, 2399, 2f),
            ),
        )
        assertEquals(
            DeviceClass.TABLET,
            ChatBackendPolicy.deviceClassFor(
                ChatBackendPolicy.maximumShortestEdgeDp(1800, 2560, 3f),
            ),
        )
    }

    @Test
    fun backendSelectionSeparatesUniversalAndLegacyTabletBehavior() {
        assertEquals(
            ChatBackend.EMBEDDED_HERDR,
            ChatBackendPolicy.backendFor(DeviceClass.PHONE, nativeTermuxAvailable = true),
        )
        assertEquals(
            ChatBackend.EMBEDDED_HERDR,
            ChatBackendPolicy.backendFor(DeviceClass.PHONE, nativeTermuxAvailable = false),
        )
        assertEquals(
            ChatBackend.NATIVE_TERMUX,
            ChatBackendPolicy.backendFor(DeviceClass.TABLET, nativeTermuxAvailable = true),
        )
        assertEquals(
            ChatBackend.UNIVERSAL_UPGRADE_REQUIRED,
            ChatBackendPolicy.backendFor(DeviceClass.TABLET, nativeTermuxAvailable = false),
        )
    }

    @Test
    fun savedClassStaysStableAcrossRotationAndMultiWindowResize() {
        // The live session keeps its saved class even if a resize would now be
        // below the threshold. A different display gets a fresh classification.
        assertEquals(
            DeviceClass.TABLET,
            restoreSessionDeviceClass(
                savedClass = DeviceClass.TABLET.name,
                savedDisplayId = 7,
                currentDisplayId = 7,
            ),
        )
        assertEquals(
            DeviceClass.PHONE,
            restoreSessionDeviceClass(
                savedClass = DeviceClass.PHONE.name,
                savedDisplayId = 7,
                currentDisplayId = 7,
            ),
        )
        assertNull(
            restoreSessionDeviceClass(
                savedClass = DeviceClass.TABLET.name,
                savedDisplayId = 7,
                currentDisplayId = 8,
            ),
        )
        assertNull(restoreSessionDeviceClass("unknown", savedDisplayId = 7, currentDisplayId = 7))
    }

    @Test
    fun termuxRouteUsesTheCanonicalActivityClassName() {
        assertEquals("com.termux.app.TermuxActivity", TermuxRoute.activityClassName)
    }

    @Test
    fun purePolicyLeavesDeviceIdentityAndPackageStringsOut() {
        // The policy maps only pure metrics and variant capability, never
        // Build.MODEL, Build.DEVICE, serial, or package applicationId.
        val phoneClass = ChatBackendPolicy.deviceClassFor(500)
        val tabletClass = ChatBackendPolicy.deviceClassFor(800)

        assertEquals(DeviceClass.PHONE, phoneClass)
        assertEquals(DeviceClass.TABLET, tabletClass)

        // Universal package assumption: nativeTermuxAvailable = true
        assertEquals(ChatBackend.EMBEDDED_HERDR, ChatBackendPolicy.backendFor(phoneClass, true))
        assertEquals(ChatBackend.NATIVE_TERMUX, ChatBackendPolicy.backendFor(tabletClass, true))

        // Legacy package assumption: nativeTermuxAvailable = false
        assertEquals(ChatBackend.EMBEDDED_HERDR, ChatBackendPolicy.backendFor(phoneClass, false))
        assertEquals(
            ChatBackend.UNIVERSAL_UPGRADE_REQUIRED,
            ChatBackendPolicy.backendFor(tabletClass, false),
        )
    }
}
