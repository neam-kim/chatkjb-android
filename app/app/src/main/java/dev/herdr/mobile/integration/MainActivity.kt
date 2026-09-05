package dev.herdr.mobile.integration

import android.Manifest
import android.content.Intent
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.BackHandler
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.termux.app.TermuxActivity
import dev.herdr.mobile.core.navigation.AppDestination
import dev.herdr.mobile.core.navigation.EmailRoute
import dev.herdr.mobile.core.navigation.HomepageRoute
import dev.herdr.mobile.core.navigation.MoonlightRoute
import dev.herdr.mobile.core.navigation.parseDestinationIntent
import dev.herdr.mobile.features.homepage.HomepageWebScreen
import dev.herdr.mobile.features.homepage.KimJbLauncher
import dev.herdr.mobile.features.chat.ui.theme.HerdrTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        if (Build.VERSION.SDK_INT >= 33) {
            registerForActivityResult(ActivityResultContracts.RequestPermission()) {}
                .launch(Manifest.permission.POST_NOTIFICATIONS)
        }

        val requestedDestination = parseDestinationIntent(intent)

        setContent {
            var destination by remember { mutableStateOf(requestedDestination) }
            var emailError by remember { mutableStateOf(false) }

            LaunchedEffect(destination) {
                if (destination == AppDestination.SERVER) {
                    destination = null
                    openMoonlight()
                }
            }

            HerdrTheme {
                if (destination == AppDestination.HOMEPAGE || destination == AppDestination.FINANCE) {
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
                                    destination = null
                                    emailError = !openEmail()
                                }
                                AppDestination.CHAT_KJB -> {
                                    openTermux()
                                }
                                AppDestination.HOME -> destination = null
                                else -> Unit
                            }
                        },
                    )
                } else {
                    KimJbLauncher(
                        emailError = emailError,
                        onSite = { destination = AppDestination.HOMEPAGE },
                        onFinance = {
                            emailError = false
                            destination = AppDestination.FINANCE
                        },
                        onEmail = { emailError = !openEmail() },
                        onChat = {
                            emailError = false
                            openTermux()
                        },
                        onServer = {
                            emailError = false
                            openMoonlight()
                        },
                    )
                    BackHandler { finish() }
                }
            }
        }
    }

    private fun openEmail(): Boolean {
        val nativeIntent = EmailRoute.nativeLaunchIntent()
        return runCatching { startActivity(nativeIntent) }.isSuccess
    }

    private fun openTermux() {
        startActivity(Intent(this, TermuxActivity::class.java))
    }

    private fun openMoonlight() {
        startActivity(MoonlightRoute.launchIntent(packageName))
    }
}
