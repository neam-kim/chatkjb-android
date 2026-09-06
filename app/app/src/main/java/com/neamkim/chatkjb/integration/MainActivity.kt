package com.neamkim.chatkjb.integration

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.view.Display
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import com.neamkim.chatkjb.BuildConfig
import com.neamkim.chatkjb.core.navigation.AppDestination
import com.neamkim.chatkjb.core.navigation.ChatBackend
import com.neamkim.chatkjb.core.navigation.ChatBackendPolicy
import com.neamkim.chatkjb.core.navigation.DEVICE_CLASS_DISPLAY_STATE_KEY
import com.neamkim.chatkjb.core.navigation.DEVICE_CLASS_STATE_KEY
import com.neamkim.chatkjb.core.navigation.DeviceClass
import com.neamkim.chatkjb.core.navigation.EmailRoute
import com.neamkim.chatkjb.core.navigation.TermuxRoute
import com.neamkim.chatkjb.core.navigation.parseDestinationIntent
import com.neamkim.chatkjb.core.navigation.restoreSessionDeviceClass
import androidx.window.layout.WindowMetricsCalculator
import com.limelight.PcView
import com.neamkim.chatkjb.features.herdr.push.HerdrPushRegistration
import org.unifiedpush.android.connector.UnifiedPush

class MainActivity : ComponentActivity() {
    private lateinit var sessionDeviceClass: DeviceClass
    private var sessionDisplayId: Int = Display.DEFAULT_DISPLAY
    private var requestedDestination by mutableStateOf<AppDestination?>(null)
    private var setupFragment by mutableStateOf<String?>(null)
    private var routeRevision by mutableIntStateOf(0)

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

        sessionDisplayId = activeDisplayId()
        routeRevision = savedInstanceState?.getInt(ROUTE_REVISION_STATE_KEY) ?: 0
        sessionDeviceClass = restoreSessionDeviceClass(
            savedClass = savedInstanceState?.getString(DEVICE_CLASS_STATE_KEY),
            savedDisplayId = savedInstanceState?.takeIf {
                it.containsKey(DEVICE_CLASS_DISPLAY_STATE_KEY)
            }?.getInt(DEVICE_CLASS_DISPLAY_STATE_KEY),
            currentDisplayId = sessionDisplayId,
        ) ?: classifyActiveDisplay()

        if (savedInstanceState == null) {
            updateRequestedRoute(intent)
        } else {
            restoreRequestedRouteData(intent)
        }

        setContent {
            ChatKjbApp(
                requestedDestination = requestedDestination,
                setupFragment = setupFragment,
                routeRevision = routeRevision,
                deviceClass = sessionDeviceClass,
                nativeTermuxAvailable = BuildConfig.NATIVE_TERMUX_AVAILABLE,
                openEmail = ::openEmail,
                openMoonlight = {
                    startActivity(android.content.Intent(this, PcView::class.java))
                },
                openTermux = ::openTermux,
                onFinish = ::finish,
            )
        }
    }

    override fun onNewIntent(intent: android.content.Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        updateRequestedRoute(intent)
    }

    override fun onSaveInstanceState(outState: Bundle) {
        outState.putString(DEVICE_CLASS_STATE_KEY, sessionDeviceClass.name)
        outState.putInt(DEVICE_CLASS_DISPLAY_STATE_KEY, sessionDisplayId)
        outState.putInt(ROUTE_REVISION_STATE_KEY, routeRevision)
        super.onSaveInstanceState(outState)
    }

    private fun classifyActiveDisplay(): DeviceClass {
        val bounds = WindowMetricsCalculator.getOrCreate()
            .computeMaximumWindowMetrics(this)
            .bounds
        val shortestEdgeDp = ChatBackendPolicy.maximumShortestEdgeDp(
            widthPx = bounds.width(),
            heightPx = bounds.height(),
            density = resources.displayMetrics.density,
        )
        return ChatBackendPolicy.deviceClassFor(shortestEdgeDp)
    }

    private fun openEmail(): Boolean = runCatching {
        startActivity(EmailRoute.nativeLaunchIntent(BuildConfig.APPLICATION_ID))
    }.isSuccess

    private fun openTermux() {
        startActivity(TermuxRoute.launchIntent(this))
    }

    @Suppress("DEPRECATION")
    private fun activeDisplayId(): Int = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
        display?.displayId ?: windowManager.defaultDisplay.displayId
    } else {
        windowManager.defaultDisplay.displayId
    }

    private fun updateRequestedRoute(newIntent: android.content.Intent) {
        routeRevision++
        requestedDestination = parseDestinationIntent(newIntent)
        setupFragment = setupFragmentFor(newIntent)
    }

    private fun restoreRequestedRouteData(newIntent: android.content.Intent) {
        requestedDestination = parseDestinationIntent(newIntent)
        setupFragment = setupFragmentFor(newIntent)
    }

    private fun setupFragmentFor(newIntent: android.content.Intent): String? =
        newIntent.data
            ?.takeIf {
                requestedDestination == AppDestination.CHAT_KJB &&
                    ChatBackendPolicy.backendFor(
                        sessionDeviceClass,
                        BuildConfig.NATIVE_TERMUX_AVAILABLE,
                    ) == ChatBackend.EMBEDDED_HERDR
            }
            ?.encodedFragment

}

private const val ROUTE_REVISION_STATE_KEY = "chatkjb.route_revision"
