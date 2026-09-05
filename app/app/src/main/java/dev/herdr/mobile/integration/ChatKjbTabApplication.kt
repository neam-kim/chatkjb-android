package dev.herdr.mobile.integration

import android.content.res.AssetManager
import com.termux.app.TermuxApplication
import com.termux.shared.termux.TermuxConstants
import com.termux.shared.termux.settings.preferences.TermuxAppSharedPreferences
import java.io.FileOutputStream
import net.thunderbird.android.ThunderbirdApp

/**
 * Tablet host that starts KJBMail, then Termux, then installs the D2Coding
 * default terminal font and Korean IME composing preview.
 */
class ChatKjbTabApplication : ThunderbirdApp() {
    override fun onCreate() {
        super.onCreate()
        TermuxApplication.initializeTermux(this)
        installDefaultTermuxFont(assets)
        enableImeComposingPreview()
    }

    private fun installDefaultTermuxFont(assets: AssetManager) {
        val fontFile = TermuxConstants.TERMUX_FONT_FILE
        fontFile.parentFile?.mkdirs()
        assets.open(FONT_ASSET).use { input ->
            FileOutputStream(fontFile).use { output ->
                input.copyTo(output)
            }
        }
    }

    private fun enableImeComposingPreview() {
        val preferences = TermuxAppSharedPreferences.build(this, false) ?: return
        if (!preferences.isImeComposingEnabled()) {
            preferences.setImeComposingEnabled(true)
        }
    }

    companion object {
        private const val FONT_ASSET = "fonts/D2Coding-Regular.ttf"
    }
}
