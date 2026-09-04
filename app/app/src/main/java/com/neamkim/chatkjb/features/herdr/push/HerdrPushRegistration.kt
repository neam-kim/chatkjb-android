package com.neamkim.chatkjb.features.herdr.push

import android.content.Context
import android.util.Log
import com.neamkim.chatkjb.BuildConfig
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import java.net.HttpURLConnection
import java.net.URL
import kotlin.concurrent.thread

object HerdrPushRegistration {
    private const val TAG = "HerdrPushRegistration"
    private const val PREFERENCES = "herdr_push"
    private const val ENDPOINT = "endpoint"

    fun saveAndSync(context: Context, endpoint: String) {
        context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
            .edit()
            .putString(ENDPOINT, endpoint)
            .apply()
        sync(endpoint)
    }

    fun syncStored(context: Context) {
        val endpoint = context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
            .getString(ENDPOINT, null)
            ?: return
        sync(endpoint)
    }

    fun clear(context: Context) {
        context.getSharedPreferences(PREFERENCES, Context.MODE_PRIVATE)
            .edit()
            .remove(ENDPOINT)
            .apply()
    }

    private fun sync(endpoint: String) {
        val registrationUrl = BuildConfig.HERDR_PUSH_REGISTRATION_URL
        val token = BuildConfig.HERDR_PUSH_REGISTRATION_TOKEN
        if (registrationUrl.isBlank() || token.isBlank() || endpoint.isBlank()) return
        thread(name = "herdr-push-registration", isDaemon = true) {
            val connection = runCatching {
                (URL(registrationUrl).openConnection() as HttpURLConnection).apply {
                    requestMethod = "POST"
                    connectTimeout = 10_000
                    readTimeout = 10_000
                    doOutput = true
                    setRequestProperty("Authorization", "Bearer $token")
                    setRequestProperty("Content-Type", "application/json")
                    outputStream.use { stream ->
                        stream.write(
                            buildJsonObject { put("endpoint", endpoint) }
                                .toString()
                                .encodeToByteArray(),
                        )
                    }
                }
            }.getOrElse {
                Log.w(TAG, "Could not connect to the Herdr notification registrar", it)
                return@thread
            }
            try {
                if (connection.responseCode != HttpURLConnection.HTTP_NO_CONTENT) {
                    Log.w(TAG, "Herdr notification registration failed: HTTP ${connection.responseCode}")
                }
            } catch (error: Exception) {
                Log.w(TAG, "Could not finish Herdr notification registration", error)
            } finally {
                connection.disconnect()
            }
        }
    }
}
