package com.neamkim.chatkjb.integration

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import com.neamkim.chatkjb.BuildConfig
import com.neamkim.chatkjb.core.navigation.AppDestination
import com.neamkim.chatkjb.core.navigation.EmailRoute
import com.neamkim.chatkjb.core.navigation.parseDestinationIntent
import com.limelight.PcView
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
            ChatKjbApp(
                requestedDestination = requestedDestination,
                setupFragment = setupFragment,
                openEmail = ::openEmail,
                openMoonlight = {
                    startActivity(android.content.Intent(this, PcView::class.java))
                },
                onFinish = ::finish,
            )
        }
    }

    private fun openEmail(): Boolean = runCatching {
        startActivity(EmailRoute.nativeLaunchIntent(BuildConfig.APPLICATION_ID))
    }.isSuccess

}
