package com.neamkim.chatkjb.integration

import androidx.activity.compose.BackHandler
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.neamkim.chatkjb.core.navigation.AppDestination
import com.neamkim.chatkjb.core.navigation.HerdrRoute
import com.neamkim.chatkjb.core.navigation.HomepageRoute
import com.neamkim.chatkjb.core.navigation.ManagementConsoleRoute
import com.neamkim.chatkjb.features.console.ManagementConsoleScreen
import com.neamkim.chatkjb.features.homepage.HomepageWebScreen
import com.neamkim.chatkjb.features.homepage.KimJbConsoleSettings
import com.neamkim.chatkjb.features.homepage.KimJbLauncher
import com.neamkim.chatkjb.features.herdr.EmbeddedHerdrScreen

/** App composition and screen transitions; platform launches are supplied by the activity. */
@Composable
internal fun ChatKjbApp(
    requestedDestination: AppDestination?,
    setupFragment: String?,
    openEmail: () -> Boolean,
    openMoonlight: () -> Unit,
    onFinish: () -> Unit,
) {
    var destination by remember { mutableStateOf(requestedDestination) }
    var emailError by remember { mutableStateOf(false) }

    LaunchedEffect(destination) {
        when (destination) {
            AppDestination.EMAIL -> {
                destination = null
                emailError = !openEmail()
            }
            else -> Unit
        }
    }

    MaterialTheme {
        when (destination) {
            AppDestination.CHAT_KJB -> {
                EmbeddedHerdrScreen(
                    startUrl = HerdrRoute.embeddedUrl(setupFragment),
                    onExit = { destination = null },
                )
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
