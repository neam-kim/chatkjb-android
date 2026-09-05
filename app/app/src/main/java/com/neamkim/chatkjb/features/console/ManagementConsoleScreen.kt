package com.neamkim.chatkjb.features.console

import android.annotation.SuppressLint
import android.view.ViewGroup
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.systemBarsPadding
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.viewinterop.AndroidView
import com.neamkim.chatkjb.core.web.configurePrivateWebContent
import com.neamkim.chatkjb.core.navigation.ManagementConsoleRoute

/** Tailnet-only WebView for one of the two fixed Mac management consoles. */
@SuppressLint("SetJavaScriptEnabled")
@Composable
fun ManagementConsoleScreen(
    startUrl: String,
    onExit: () -> Unit,
) {
    val webViewState = remember { mutableStateOf<WebView?>(null) }

    BackHandler {
        val view = webViewState.value
        if (view != null && view.canGoBack()) view.goBack() else onExit()
    }

    Surface(modifier = Modifier.fillMaxSize(), color = Color(0xFF0C0E14)) {
        AndroidView(
            modifier = Modifier
                .fillMaxSize()
                .systemBarsPadding(),
            factory = { context ->
                WebView(context).apply {
                    layoutParams = ViewGroup.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT,
                        ViewGroup.LayoutParams.MATCH_PARENT,
                    )
                    setBackgroundColor(android.graphics.Color.rgb(12, 14, 20))
                    configurePrivateWebContent()
                    settings.useWideViewPort = true
                    settings.loadWithOverviewMode = false

                    webViewClient = ManagementConsoleWebViewClient(startUrl)
                    loadUrl(startUrl)
                    webViewState.value = this
                }
            },
            onRelease = { view ->
                webViewState.value = null
                view.stopLoading()
                view.destroy()
            },
        )
    }
}

private class ManagementConsoleWebViewClient(
    private val entryUrl: String,
) : WebViewClient() {
    override fun shouldOverrideUrlLoading(view: WebView, request: WebResourceRequest): Boolean =
        !ManagementConsoleRoute.isAllowed(request.url.toString(), entryUrl)

    override fun onReceivedError(
        view: WebView,
        request: WebResourceRequest,
        error: WebResourceError,
    ) {
        if (!request.isForMainFrame) return
        view.loadData(
            "<main style='color:#ededed;background:#0c0e14;font-family:sans-serif;padding:24px'>" +
                "<h1>Console unavailable</h1>" +
                "<p>Mac mini와 Tailscale 연결을 확인해 주세요.</p></main>",
            "text/html",
            "UTF-8",
        )
    }
}
