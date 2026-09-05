package net.thunderbird.android.provider

import net.thunderbird.android.R
import net.thunderbird.core.ui.theme.api.ThemeProvider

internal class TbThemeProvider : ThemeProvider {
    override val appThemeResourceId = R.style.Theme_Thunderbird_Dark
    override val appLightThemeResourceId = R.style.Theme_Thunderbird_Light
    override val appDarkThemeResourceId = R.style.Theme_Thunderbird_Dark
    override val dialogThemeResourceId = R.style.Theme_Thunderbird_Dark_Dialog
    override val translucentDialogThemeResourceId = R.style.Theme_Thunderbird_Dark_Dialog_Translucent
}
