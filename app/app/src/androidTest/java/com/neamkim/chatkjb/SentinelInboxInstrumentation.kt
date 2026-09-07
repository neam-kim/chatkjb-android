package com.neamkim.chatkjb

import android.app.Activity
import android.app.Instrumentation
import android.app.NotificationManager
import android.os.Bundle
import com.neamkim.chatkjb.features.herdr.push.AutomationInbox
import com.neamkim.chatkjb.features.herdr.push.HerdrPushPayload
import com.neamkim.chatkjb.features.herdr.push.HerdrNotifications

/** Local-only UI/OS notification fixture. Never sends a network push or starts an agent. */
class SentinelInboxInstrumentation : Instrumentation() {
    private var action = ""
    private val fixture = HerdrPushPayload(
        kind = "sentinel-problem", paneId = "sentinel:ui-test",
        title = "Sentinel UI test", body = "Problem! No:900003",
    )
    override fun onCreate(arguments: Bundle?) {
        super.onCreate(arguments)
        action = arguments?.getString("action").orEmpty()
        start()
    }
    override fun onStart() {
        val result = Bundle()
        try {
            runOnMainSync {
                val prefs = AutomationInbox.preferences(targetContext)
                when (action) {
                    "seed" -> {
                        check(AutomationInbox.sentinel(targetContext) == null) { "Existing user alert must be preserved" }
                        check(!prefs.contains("test_previous_ack")) { "Fixture already active" }
                        check(prefs.edit().putString("test_previous_ack", prefs.getString("sentinel_ack", "")).commit())
                        check(AutomationInbox.update(targetContext, fixture))
                        HerdrNotifications.post(targetContext, fixture)
                        check(prefs.edit().commit())
                    }
                    "verify-and-restore" -> {
                        check(AutomationInbox.sentinel(targetContext) == null) { "Done did not remove fixture" }
                        check(targetContext.getSystemService(NotificationManager::class.java).activeNotifications.none {
                            it.id == fixture.paneId.hashCode()
                        }) { "Done did not cancel the OS notification" }
                        check(prefs.getString("sentinel_ack", null) == "sentinel-problem\nProblem! No:900003") { "Done did not acknowledge fixture" }
                        check(!AutomationInbox.update(targetContext, fixture)) { "Acknowledged duplicate was delivered again" }
                        check(AutomationInbox.sentinel(targetContext) == null)
                        val previous = prefs.getString("test_previous_ack", null)
                        check(previous != null) { "No fixture backup" }
                        val edit = prefs.edit().remove("test_previous_ack")
                        if (previous.isEmpty()) edit.remove("sentinel_ack") else edit.putString("sentinel_ack", previous)
                        check(edit.commit())
                    }
                    else -> error("Unknown fixture action")
                }
            }
            if (action == "seed") {
                val manager = targetContext.getSystemService(NotificationManager::class.java)
                val deadline = android.os.SystemClock.elapsedRealtime() + 3000
                while (manager.activeNotifications.none { it.id == fixture.paneId.hashCode() } &&
                    android.os.SystemClock.elapsedRealtime() < deadline) Thread.sleep(50)
                check(manager.activeNotifications.any { it.id == fixture.paneId.hashCode() }) { "OS notification was not posted" }
            }
            result.putString("stream", "Sentinel fixture $action: PASS (inbox, acknowledgement, OS notification)\n")
            finish(Activity.RESULT_OK, result)
        } catch (error: Throwable) {
            result.putString("stream", "Sentinel fixture $action: FAIL ${error.message}\n")
            finish(Activity.RESULT_CANCELED, result)
        }
    }
}
