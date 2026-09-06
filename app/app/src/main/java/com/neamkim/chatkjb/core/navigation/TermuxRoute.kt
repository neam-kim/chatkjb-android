package com.neamkim.chatkjb.core.navigation

import android.content.Context
import android.content.Intent

/** Native Termux entry point resolved only after the universal policy selects it. */
object TermuxRoute {
    const val activityClassName = "com.termux.app.TermuxActivity"

    fun launchIntent(context: Context): Intent = Intent().apply {
        setClassName(context, activityClassName)
    }
}
