package com.neamkim.chatkjb.features.herdr.push

import android.content.Context
import org.unifiedpush.android.connector.FailedReason
import org.unifiedpush.android.connector.MessagingReceiver
import org.unifiedpush.android.connector.data.PushEndpoint
import org.unifiedpush.android.connector.data.PushMessage

class HerdrUnifiedPushReceiver : MessagingReceiver() {
    override fun onNewEndpoint(context: Context, endpoint: PushEndpoint, instance: String) {
        HerdrPushRegistration.saveAndSync(context, endpoint.url)
    }

    override fun onMessage(context: Context, message: PushMessage, instance: String) {
        parseHerdrPush(message.content)?.let { payload ->
            if (!AutomationInbox.update(context, payload)) return@let
            if (payload.kind in setOf("clear", "sentinel-clear", "skill-clear")) {
                HerdrNotifications.cancel(context, payload.paneId)
            } else {
                HerdrNotifications.post(context, payload)
            }
        }
    }

    override fun onRegistrationFailed(
        context: Context,
        reason: FailedReason,
        instance: String,
    ) = Unit

    override fun onUnregistered(context: Context, instance: String) {
        HerdrPushRegistration.clear(context)
    }
}
