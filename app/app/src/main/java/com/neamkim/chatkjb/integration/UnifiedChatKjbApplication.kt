package com.neamkim.chatkjb.integration

import com.termux.app.TermuxApplication
import com.termux.shared.termux.TermuxConstants
import com.termux.shared.termux.settings.preferences.TermuxPreferenceConstants
import com.termux.shared.termux.settings.preferences.TermuxAppSharedPreferences
import net.thunderbird.android.ThunderbirdApp

/**
 * Universal tablet host that starts ThunderbirdApp, then Termux, and sets up
 * the D2Coding terminal font and Korean IME composing preview.
 */
class UnifiedChatKjbApplication : ThunderbirdApp() {
    override fun onCreate() {
        super.onCreate()
        TermuxApplication.initializeTermux(this)
        installDefaultTermuxFont()
        enableImeComposingPreview()
    }

    private fun installDefaultTermuxFont() {
        installFontIfMissing(TermuxConstants.TERMUX_FONT_FILE) {
            assets.open(FONT_ASSET)
        }
    }

    private fun enableImeComposingPreview() {
        val stored = getSharedPreferences(
            TermuxConstants.TERMUX_DEFAULT_PREFERENCES_FILE_BASENAME_WITHOUT_EXTENSION,
            MODE_PRIVATE,
        )
        if (!stored.contains(TermuxPreferenceConstants.TERMUX_APP.KEY_IME_COMPOSING_ENABLED)) {
            TermuxAppSharedPreferences.build(this, false)?.setImeComposingEnabled(true)
        }
    }

    companion object {
        private const val FONT_ASSET = "fonts/D2Coding-Regular.ttf"
    }
}
