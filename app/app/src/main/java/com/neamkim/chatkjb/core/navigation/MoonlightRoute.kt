package com.neamkim.chatkjb.core.navigation

/** In-process Moonlight entry point used by the launcher Server action. */
object MoonlightRoute {
    const val activityClassName = "com.limelight.PcView"

    fun activityClassNameFor(destination: AppDestination): String? =
        if (destination == AppDestination.MOONLIGHT) activityClassName else null
}
