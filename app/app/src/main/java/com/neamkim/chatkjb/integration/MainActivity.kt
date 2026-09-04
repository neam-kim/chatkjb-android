package com.neamkim.chatkjb.integration

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.BackHandler
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.neamkim.chatkjb.BuildConfig
import com.neamkim.chatkjb.core.navigation.AppDestination
import com.neamkim.chatkjb.core.navigation.EmailRoute
import com.neamkim.chatkjb.core.navigation.HerdrRoute
import com.neamkim.chatkjb.core.navigation.HomepageRoute
import com.neamkim.chatkjb.core.navigation.ManagementConsoleRoute
import com.neamkim.chatkjb.core.navigation.parseDestinationIntent
import com.neamkim.chatkjb.features.console.ManagementConsoleScreen
import com.neamkim.chatkjb.features.homepage.HomepageWebScreen
import com.neamkim.chatkjb.features.homepage.KimJbConsoleSettings
import com.neamkim.chatkjb.features.homepage.KimJbLauncher
import com.neamkim.chatkjb.features.herdr.EmbeddedHerdrScreen
import com.neamkim.chatkjb.features.herdr.push.HerdrPushRegistration
import org.unifiedpush.android.connector.UnifiedPush

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

		if (Build.VERSION.SDK_INT >= 33 &&
			checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED
		) {
			registerForActivityResult(ActivityResultContracts.RequestPermission()) {}
				.launch(Manifest.permission.POST_NOTIFICATIONS)
		}
		UnifiedPush.tryUseCurrentOrDefaultDistributor(this) { success ->
			if (success) UnifiedPush.register(this)
		}
		HerdrPushRegistration.syncStored(applicationContext)

        val requestedDestination = parseDestinationIntent(intent)
        val setupFragment = intent.data
            ?.takeIf { requestedDestination == AppDestination.CHAT_KJB }
            ?.encodedFragment

        setContent {
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
                if (destination == AppDestination.CHAT_KJB) {
                    EmbeddedHerdrScreen(
                        startUrl = HerdrRoute.embeddedUrl(setupFragment),
                        onExit = { destination = null },
                    )
                } else if (destination == AppDestination.AUTOBOT || destination == AppDestination.SERVER) {
                    ManagementConsoleScreen(
                        startUrl = if (destination == AppDestination.AUTOBOT) {
                            ManagementConsoleRoute.autoBotUrl
                        } else {
                            ManagementConsoleRoute.serverUrl
                        },
                        onExit = { destination = AppDestination.CONSOLE_SETTINGS },
                    )
                } else if (destination == AppDestination.CONSOLE_SETTINGS) {
                    KimJbConsoleSettings(
                        onAutoBot = { destination = AppDestination.AUTOBOT },
                        onServer = { destination = AppDestination.SERVER },
                    )
                    BackHandler { destination = null }
                } else if (destination == AppDestination.HOMEPAGE || destination == AppDestination.FINANCE) {
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
                                -> Unit
                            }
                        },
                    )
                } else {
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
                        onConsoleSettings = {
                            emailError = false
                            destination = AppDestination.CONSOLE_SETTINGS
                        },
                    )
                    BackHandler { finish() }
                }
            }
        }
    }

    private fun openEmail(): Boolean = runCatching {
        startActivity(EmailRoute.nativeLaunchIntent(BuildConfig.APPLICATION_ID))
    }.isSuccess

}
