package com.neamkim.chatkjb.core.navigation

import android.content.Intent

/**
 * Mail is hosted in this process by the `mail-host` module, so it is reached by
 * starting the embedded entry point directly rather than resolving another package.
 */
object EmailRoute {
    fun nativeLaunchIntent(packageName: String): Intent = Intent(Intent.ACTION_MAIN).apply {
        setClassName(packageName, "net.thunderbird.android.MailEntryActivity")
    }
}
