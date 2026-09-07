package com.neamkim.chatkjb.integration

import androidx.activity.compose.BackHandler
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import com.neamkim.chatkjb.core.navigation.AppDestination
import com.neamkim.chatkjb.core.navigation.ChatBackend
import com.neamkim.chatkjb.core.navigation.ChatBackendPolicy
import com.neamkim.chatkjb.core.navigation.DeviceClass
import com.neamkim.chatkjb.core.navigation.HerdrRoute
import com.neamkim.chatkjb.core.navigation.HomepageRoute
import com.neamkim.chatkjb.core.navigation.ManagementConsoleRoute
import com.neamkim.chatkjb.features.console.ManagementConsoleScreen
import com.neamkim.chatkjb.features.herdr.NativeBackendUnavailableScreen
import com.neamkim.chatkjb.features.homepage.HomepageWebScreen
import com.neamkim.chatkjb.features.homepage.KimJbConsoleSettings
import com.neamkim.chatkjb.features.homepage.KimJbLauncher
import com.neamkim.chatkjb.features.herdr.EmbeddedHerdrScreen

/** App composition and screen transitions; platform launches are supplied by the activity. */
@Composable
internal fun ChatKjbApp(
    requestedDestination: AppDestination?,
    setupFragment: String?,
    routeRevision: Int,
    deviceClass: DeviceClass,
    nativeTermuxAvailable: Boolean,
    openEmail: () -> Boolean,
    openMoonlight: () -> Unit,
    openTermux: () -> Unit,
    onFinish: () -> Unit,
) {
    var destination by rememberSaveable { mutableStateOf(requestedDestination) }
    var lastHandledRevision by rememberSaveable { mutableStateOf(routeRevision) }
    var sentinelRequest by rememberSaveable { mutableStateOf<String?>(null) }
    var emailError by remember { mutableStateOf(false) }

    if (lastHandledRevision != routeRevision) {
        lastHandledRevision = routeRevision
        destination = requestedDestination
    }

    LaunchedEffect(destination) {
        when (destination) {
            AppDestination.EMAIL -> {
                destination = null
                emailError = !openEmail()
            }
            AppDestination.CHAT_KJB -> {
                if (ChatBackendPolicy.backendFor(deviceClass, nativeTermuxAvailable) == ChatBackend.NATIVE_TERMUX) {
                    destination = null
                    openTermux()
                }
            }
            else -> Unit
        }
    }

    MaterialTheme {
        if (sentinelRequest != null) {
            EmbeddedHerdrScreen(
                startUrl = HerdrRoute.embeddedUrl + "#sentinel=" + sentinelRequest,
                authorizeSentinelStart = true,
                onExit = { sentinelRequest = null },
            )
        } else when (destination) {
            AppDestination.CHAT_KJB -> {
                when (ChatBackendPolicy.backendFor(deviceClass, nativeTermuxAvailable)) {
                    ChatBackend.EMBEDDED_HERDR -> {
                        EmbeddedHerdrScreen(
                            startUrl = HerdrRoute.embeddedUrl(setupFragment),
                            onExit = { destination = null },
                        )
                    }
                    ChatBackend.UNIVERSAL_UPGRADE_REQUIRED -> {
                        NativeBackendUnavailableScreen(
                            onExit = { destination = null },
                        )
                    }
                    ChatBackend.NATIVE_TERMUX -> Unit
                }
            }
            AppDestination.AUTOBOT -> {
                ManagementConsoleScreen(
                    startUrl = ManagementConsoleRoute.autoBotUrl,
                    onExit = { destination = AppDestination.CONSOLE_SETTINGS },
                )
            }
            AppDestination.SERVER -> {
                ManagementConsoleScreen(
                    startUrl = ManagementConsoleRoute.serverUrl,
                    onExit = { destination = AppDestination.CONSOLE_SETTINGS },
                )
            }
            AppDestination.MOONLIGHT -> {
                // Launcher Server opens the transplanted Moonlight client in this process.
                LaunchedEffect(Unit) {
                    openMoonlight()
                    destination = null
                }
            }
            AppDestination.CONSOLE_SETTINGS -> {
                KimJbConsoleSettings(
                    onInvestigateSentinel = { item ->
                        val target = org.json.JSONObject()
                            .put("body", item.body).put("receivedAt", item.updatedAt)
                        sentinelRequest = java.net.URLEncoder.encode(target.toString(), "UTF-8")
                    },
                    onAutoBot = { destination = AppDestination.AUTOBOT },
                    onServer = { destination = AppDestination.SERVER },
                )
                BackHandler { destination = null }
            }
            AppDestination.HOMEPAGE, AppDestination.FINANCE -> {
                HomepageWebScreen(
                    onExit = { destination = null },
                    startUrl = if (destination == AppDestination.FINANCE) {
                        HomepageRoute.financeUrl
                    } else {
                        HomepageRoute.canonicalUrl
                    },
                    onDestination = { target ->
                        when (target) {
                            AppDestination.EMAIL -> {
                                emailError = !openEmail()
                                destination = null
                            }
                            AppDestination.CHAT_KJB -> {
                                destination = AppDestination.CHAT_KJB
                            }
                            AppDestination.HOME -> destination = null
                            AppDestination.HOMEPAGE,
                            AppDestination.FINANCE,
                            AppDestination.CONSOLE_SETTINGS,
                            AppDestination.AUTOBOT,
                            AppDestination.SERVER,
                            AppDestination.MOONLIGHT,
                            -> Unit
                        }
                    },
                )
            }
            AppDestination.HOME, AppDestination.EMAIL, null -> {
                KimJbLauncher(
                    emailError = emailError,
                    onHomepage = { destination = AppDestination.HOMEPAGE },
                    onFinance = {
                        emailError = false
                        destination = AppDestination.FINANCE
                    },
                    onEmail = { emailError = !openEmail() },
                    onChat = {
                        emailError = false
                        destination = AppDestination.CHAT_KJB
                    },
                    onServer = {
                        emailError = false
                        destination = AppDestination.MOONLIGHT
                    },
                    onConsoleSettings = {
                        emailError = false
                        destination = AppDestination.CONSOLE_SETTINGS
                    },
                )
                BackHandler(onBack = onFinish)
            }
        }
    }
}
