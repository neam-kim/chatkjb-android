package com.neamkim.chatkjb.core.web

import android.annotation.SuppressLint
import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.webkit.WebSettings
import android.webkit.WebView
import com.neamkim.chatkjb.BuildConfig

/** Common settings for the embedded client and private consoles. */
@SuppressLint("SetJavaScriptEnabled")
internal fun WebView.configurePrivateWebContent() {
    settings.javaScriptEnabled = true
    settings.domStorageEnabled = true
    settings.allowFileAccess = false
    settings.allowContentAccess = false
    settings.mixedContentMode = WebSettings.MIXED_CONTENT_NEVER_ALLOW
    settings.mediaPlaybackRequiresUserGesture = true
    if (BuildConfig.DEBUG) WebView.setWebContentsDebuggingEnabled(true)
}

internal tailrec fun Context.findActivity(): Activity? = when (this) {
    is Activity -> this
    is ContextWrapper -> baseContext.findActivity()
    else -> null
}
