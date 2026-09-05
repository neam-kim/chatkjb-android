package com.neamkim.chatkjb.features.herdr.push

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.net.Uri
import androidx.core.app.NotificationCompat
import com.neamkim.chatkjb.R
import com.neamkim.chatkjb.integration.MainActivity

object HerdrNotifications {
    private const val BLOCKED_CHANNEL = "herdr_blocked"
    private const val FINISHED_CHANNEL = "herdr_finished"
    private const val AUTOMATION_CHANNEL = "chatkjb_automation_alerts"

    fun cancel(context: Context, paneId: String) {
        context.getSystemService(NotificationManager::class.java).cancel(paneId.hashCode())
    }

    fun post(context: Context, payload: HerdrPushPayload) {
        if (payload.kind == "clear") return
        ensureChannels(context)
        val intent = Intent(context, MainActivity::class.java).apply {
            action = Intent.ACTION_VIEW
            data = Uri.parse(
                if (payload.kind.startsWith("sentinel-") || payload.kind.startsWith("skill-")) {
                    "kimjb://open/notifications"
                } else {
                    "kimjb://open/chat"
                },
            )
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
            putExtra("paneId", payload.paneId)
        }
        val pendingIntent = PendingIntent.getActivity(
            context,
            payload.paneId.hashCode(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE,
        )
        val automation = payload.kind.startsWith("sentinel-") || payload.kind.startsWith("skill-")
        val blocked = payload.kind == "blocked"
        val notification = NotificationCompat.Builder(
            context,
            when {
                automation -> AUTOMATION_CHANNEL
                blocked -> BLOCKED_CHANNEL
                else -> FINISHED_CHANNEL
            },
        )
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(payload.title)
            .setContentText(payload.body)
            .setStyle(NotificationCompat.BigTextStyle().bigText(payload.body))
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .setPriority(
                if (blocked || automation) NotificationCompat.PRIORITY_HIGH
                else NotificationCompat.PRIORITY_DEFAULT,
            )
            .build()
        context.getSystemService(NotificationManager::class.java)
            .notify(payload.paneId.hashCode(), notification)
    }

    private fun ensureChannels(context: Context) {
        val manager = context.getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(
            NotificationChannel(
                AUTOMATION_CHANNEL,
                "ChatKJB 알림 센터",
                NotificationManager.IMPORTANCE_HIGH,
            ),
        )
        manager.createNotificationChannel(
            NotificationChannel(
                BLOCKED_CHANNEL,
                "Herdr needs attention",
                NotificationManager.IMPORTANCE_HIGH,
            ),
        )
        manager.createNotificationChannel(
            NotificationChannel(
                FINISHED_CHANNEL,
                "Herdr finished",
                NotificationManager.IMPORTANCE_DEFAULT,
            ),
        )
    }
}
