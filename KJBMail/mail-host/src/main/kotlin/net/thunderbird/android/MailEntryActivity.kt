package net.thunderbird.android

import android.content.Intent
import android.os.Bundle
import com.fsck.k9.ui.base.BaseActivity
import net.thunderbird.app.common.startup.StartupRouter
import org.koin.android.ext.android.inject

/**
 * Native mail entry point owned by the unified app.
 *
 * The shared [StartupRouter] launches mail with `NEW_TASK|CLEAR_TASK`, which is correct for the
 * standalone mail app — mail is the task root there. In the unified app the KimJB launcher owns
 * the task root, so those flags wipe it out and Back from mail exits the app instead of returning
 * to the launcher. Strip the task-reset flags so mail stacks on top of the launcher instead.
 */
class MailEntryActivity : BaseActivity() {
    private val startupRouter: StartupRouter by inject()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        startupRouter.routeToNextScreen(this)
        finish()
    }

    override fun startActivity(intent: Intent) {
        super.startActivity(intent.keepInHostTask())
    }

    override fun startActivity(intent: Intent, options: Bundle?) {
        super.startActivity(intent.keepInHostTask(), options)
    }
}

private fun Intent.keepInHostTask(): Intent = apply {
    flags = flags and (Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK).inv()
}
