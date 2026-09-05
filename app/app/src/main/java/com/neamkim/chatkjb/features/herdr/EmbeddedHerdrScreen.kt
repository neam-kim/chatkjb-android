package com.neamkim.chatkjb.features.herdr

import com.neamkim.chatkjb.core.web.findActivity
import android.annotation.SuppressLint
import android.app.Activity
import android.content.ActivityNotFoundException
import android.content.Intent
import android.net.Uri
import android.view.ViewGroup
import android.webkit.ValueCallback
import android.webkit.WebChromeClient
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebResourceResponse
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.viewinterop.AndroidView
import androidx.webkit.WebViewAssetLoader
import com.neamkim.chatkjb.core.web.configurePrivateWebContent

/**
 * The ChatKJB-skinned 0cv Herdr client rendered from assets in this APK.
 *
 * [WebViewAssetLoader] gives the static bundle a secure HTTPS origin, which is
 * required by the upstream encrypted transport and device-verification APIs.
 * Navigation is confined to that origin; legal/source links intentionally open
 * in the user's browser.
 */
@SuppressLint("SetJavaScriptEnabled")
@Composable
fun EmbeddedHerdrScreen(
    startUrl: String,
    onExit: () -> Unit,
) {
    val webViewState = remember { mutableStateOf<WebView?>(null) }
    val fileCallbackState = remember { mutableStateOf<ValueCallback<Array<Uri>>?>(null) }
    val fileChooser = rememberLauncherForActivityResult(
        ActivityResultContracts.StartActivityForResult(),
    ) { result ->
        val callback = fileCallbackState.value
        fileCallbackState.value = null
        callback?.onReceiveValue(
            WebChromeClient.FileChooserParams.parseResult(result.resultCode, result.data),
        )
    }

    BackHandler {
        val view = webViewState.value
        if (view != null && view.canGoBack()) view.goBack() else onExit()
    }

    Surface(modifier = Modifier.fillMaxSize(), color = Color(0xFFF7F7F7)) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = { context ->
                val assetLoader = WebViewAssetLoader.Builder()
                    .addPathHandler(
                        "/assets/",
                        WebViewAssetLoader.AssetsPathHandler(context),
                    )
                    .build()

                WebView(context).apply {
                    layoutParams = ViewGroup.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT,
                        ViewGroup.LayoutParams.MATCH_PARENT,
                    )
                    setBackgroundColor(android.graphics.Color.rgb(247, 247, 247))
                    configurePrivateWebContent()

                    webViewClient = EmbeddedHerdrWebViewClient(
                        activity = context.findActivity(),
                        assetLoader = assetLoader,
                    )
                    webChromeClient = object : WebChromeClient() {
                        override fun onShowFileChooser(
                            webView: WebView,
                            filePathCallback: ValueCallback<Array<Uri>>,
                            fileChooserParams: FileChooserParams,
                        ): Boolean {
                            fileCallbackState.value?.onReceiveValue(null)
                            fileCallbackState.value = filePathCallback
                            return try {
                                fileChooser.launch(fileChooserParams.createIntent())
                                true
                            } catch (_: ActivityNotFoundException) {
                                fileCallbackState.value = null
                                filePathCallback.onReceiveValue(null)
                                false
                            }
                        }
                    }
                    loadUrl(startUrl)
                    webViewState.value = this
                }
            },
            onRelease = { view ->
                fileCallbackState.value?.onReceiveValue(null)
                fileCallbackState.value = null
                webViewState.value = null
                view.stopLoading()
                view.destroy()
            },
        )
    }
}

private class EmbeddedHerdrWebViewClient(
    private val activity: Activity?,
    private val assetLoader: WebViewAssetLoader,
) : WebViewClient() {
    override fun shouldInterceptRequest(
        view: WebView,
        request: WebResourceRequest,
    ): WebResourceResponse? {
        val response = assetLoader.shouldInterceptRequest(request.url) ?: return null
        if (request.url.encodedPath == "/assets/herdr/index.html") {
            response.responseHeaders = response.responseHeaders.orEmpty() +
                ("Content-Security-Policy" to TAILNET_ONLY_CSP)
        }
        return response
    }

    override fun shouldOverrideUrlLoading(view: WebView, request: WebResourceRequest): Boolean {
        if (request.url.host == WebViewAssetLoader.DEFAULT_DOMAIN) return false
        runCatching { activity?.startActivity(Intent(Intent.ACTION_VIEW, request.url)) }
        return true
    }

    override fun onReceivedError(
        view: WebView,
        request: WebResourceRequest,
        error: WebResourceError,
    ) {
        // Subresource and relay errors are rendered by the bundled UI. Preserve
        // the app surface if the immutable entry document itself cannot load.
        if (!request.isForMainFrame) return
        view.loadData(
            "<main style='font-family:sans-serif;padding:24px'>" +
                "<h1>ChatKJB</h1><p>The embedded agent UI could not be loaded.</p></main>",
            "text/html",
            "UTF-8",
        )
    }
}

private const val TAILNET_ONLY_CSP =
    "default-src 'self'; " +
        "connect-src 'self' https://neam-macmini.taild81d38.ts.net:8443 " +
        "wss://neam-macmini.taild81d38.ts.net:8443; " +
        "img-src 'self' data: blob:; font-src 'self'; " +
        "style-src 'self' 'unsafe-inline'; script-src 'self'; " +
        "worker-src 'self' blob:; media-src 'self' data: blob:; " +
        "object-src 'none'; base-uri 'self'; form-action 'self'; frame-src 'none'"
