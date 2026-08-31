package dev.herdr.mobile

import dev.herdr.mobile.features.chat.net.CompanionClient
import dev.herdr.mobile.features.chat.net.ServerFrame
import dev.herdr.mobile.features.chat.net.ClientMsg
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.runBlocking
import okhttp3.*
import okhttp3.mockwebserver.*
import org.junit.After
import org.junit.Assert.*
import org.junit.Test

class CompanionClientTest {
    // A per-test OkHttpClient the test owns, so teardown can forcibly release it.
    // CompanionClient's default ctor makes its own (unreachable) client; injecting
    // one lets @After evict the connection pool and shut the dispatcher BEFORE
    // MockWebServer.shutdown(). Without that, close() only starts an async graceful
    // WS close, the socket lingers, and MockWebServer blocks 5s waiting for its
    // reader task then throws "Gave up waiting for queue to shut down".
    private val http = OkHttpClient()
    private lateinit var server: MockWebServer
    private lateinit var client: CompanionClient

    @After fun teardown() {
        // Order matters: stop the client (graceful close + no reconnect), then FORCE
        // the WebSocket's underlying call shut (cancelAll) so the socket closes now —
        // graceful ws.close() alone depends on thread scheduling that, under
        // cross-test load, loses to MockWebServer.shutdown()'s 5s per-queue wait.
        // evictAll only closes idle pooled connections, not an active WebSocket.
        if (::client.isInitialized) client.close()
        http.dispatcher.cancelAll()
        http.dispatcher.executorService.shutdown()
        http.connectionPool.evictAll()
        if (::server.isInitialized) server.shutdownQuietly()
    }

    @Test fun receivesWelcomeAndPanes() = runBlocking {
        server = MockWebServer()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                webSocket.send("""{"t":"welcome","herdrVersion":"0.7.1","herdrProtocol":14}""")
                webSocket.send("""{"t":"panes","panes":[]}""")
            }
        }))
        server.start()
        client = CompanionClient(http)
        // Thread-safe: appended on Dispatchers.Default by the collector while the
        // runBlocking thread iterates it in the withTimeout poll below — a plain
        // mutableListOf races (ConcurrentModificationException under load).
        val collected = java.util.concurrent.CopyOnWriteArrayList<ServerFrame>()
        val job = launch(Dispatchers.Default) { client.frames.collect { collected.add(it) } }
        client.connect(server.url("/").toString().replace("http", "ws"))
        withTimeout(3000) {
            while (collected.none { it is ServerFrame.Panes }) delay(20)
        }
        assertTrue(collected.any { it is ServerFrame.Welcome })
        assertTrue(collected.any { it is ServerFrame.Panes })
        job.cancel()
    }

    @Test fun reconnectsAfterServerClose() = runBlocking {
        server = MockWebServer()
        // first connection: open then immediately close from the server side
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                webSocket.close(1000, "bye")
            }
        }))
        // second connection (after the client auto-reconnects): stays open, sends welcome
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                webSocket.send("""{"t":"welcome"}""")
            }
        }))
        server.start()
        client = CompanionClient(http)
        // Thread-safe: appended on Dispatchers.Default by the collector while the
        // runBlocking thread iterates it in the withTimeout poll below — a plain
        // mutableListOf races (ConcurrentModificationException under load).
        val collected = java.util.concurrent.CopyOnWriteArrayList<ServerFrame>()
        val job = launch(Dispatchers.Default) { client.frames.collect { collected.add(it) } }
        client.connect(server.url("/").toString().replace("http", "ws"))
        // welcome only arrives on the SECOND connection, so seeing it proves reconnect
        withTimeout(8000) {
            while (collected.none { it is ServerFrame.Welcome }) delay(50)
        }
        assertTrue(collected.any { it is ServerFrame.Welcome })
        job.cancel()
    }

    @Test fun closeImpactReturnsSiblings() = runBlocking {
        server = MockWebServer()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                if (!text.contains("\"reqId\"")) return
                val reqId = Regex("\"reqId\":\"([^\"]+)\"").find(text)!!.groupValues[1]
                if (text.contains("\"close_impact\"")) {
                    webSocket.send("""{"t":"close_impact","reqId":"$reqId","workspaceId":"w1","alsoCloses":[{"workspaceId":"w2","label":"ops"}]}""")
                }
            }
        }))
        server.start()
        client = CompanionClient(http)
        client.connect(server.url("/").toString().replace("http", "ws"))
        val result = withTimeout(3000) { client.closeImpact("w1") }
        assertEquals(1, result.size)
        assertEquals("ops", result.single().label)
    }

    @Test fun closeImpactReturnsEmptyOnError() = runBlocking {
        server = MockWebServer()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                if (!text.contains("\"reqId\"")) return
                val reqId = Regex("\"reqId\":\"([^\"]+)\"").find(text)!!.groupValues[1]
                if (text.contains("\"close_impact\"")) {
                    webSocket.send("""{"t":"error","reqId":"$reqId","code":"close_impact_failed","message":"boom"}""")
                }
            }
        }))
        server.start()
        client = CompanionClient(http)
       client.connect(server.url("/").toString().replace("http", "ws"))
       val result = withTimeout(3000) { client.closeImpact("w1") }
       assertTrue(result.isEmpty())
   }

    @Test fun qservantCatalogCompletesPendingReply() = runBlocking {
        server = MockWebServer()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                if (!text.contains("\"qservant_catalog\"")) return
                val reqId = Regex("\"reqId\":\"([^\"]+)\"").find(text)!!.groupValues[1]
                webSocket.send("""{"t":"qservant_catalog_result","reqId":"$reqId","models":[{"id":"live/a","label":"A","efforts":["low"],"quota":null}],"defaultModel":"live/a"}""")
            }
        }))
        server.start()
        client = CompanionClient(http)
        client.connect(server.url("/").toString().replace("http", "ws"))
        val frame = withTimeout(3000) { client.qservantCatalog() }
        assertTrue(frame is ServerFrame.QServantCatalogResult)
        val catalog = frame as ServerFrame.QServantCatalogResult
        assertEquals("live/a", catalog.defaultModel)
        assertEquals("A", catalog.models.single().label)
        assertNull(catalog.models.single().quota)
    }

    @Test fun qservantSubmitDeliversJobAndErrorFrames() = runBlocking {
        server = MockWebServer()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                if (!text.contains("\"reqId\"")) return
                val reqId = Regex("\"reqId\":\"([^\"]+)\"").find(text)!!.groupValues[1]
                when {
                    text.contains("\"qservant_submit\"") -> {
                        assertTrue(text.contains("\"audioMime\":\"audio/mp4\""))
                        assertTrue(text.contains("\"audioBase64\":\"YWI=\""))
                        webSocket.send("""{"t":"qservant_job","reqId":"$reqId","job":{"jobId":"job-9","state":"completed","report":{"request":"r","work":"w","verification":"v","changes":[],"result":"ok","success":true}}}""")
                    }
                    text.contains("\"qservant_status\"") -> {
                        webSocket.send("""{"t":"qservant_error","reqId":"$reqId","jobId":"job-9","code":"stt_failed","message":"boom"}""")
                    }
                }
            }
        }))
        server.start()
        client = CompanionClient(http)
        client.connect(server.url("/").toString().replace("http", "ws"))
        val job = withTimeout(3000) { client.qservantSubmit("live/a", "high", "YWI=") }
        assertTrue(job is ServerFrame.QServantJobFrame)
        job as ServerFrame.QServantJobFrame
        assertEquals("job-9", job.job.jobId)
        assertEquals("completed", job.job.state)
        assertTrue(job.job.report!!.success)
        val err = withTimeout(3000) { client.qservantStatus("job-9") }
        assertTrue(err is ServerFrame.QServantError)
        err as ServerFrame.QServantError
        assertEquals("stt_failed", err.code)
        assertEquals("job-9", err.jobId)
    }

    @Test fun awaitReplyCompletesQServantFramesOnSharedSocket() = runBlocking {
        server = MockWebServer()
        server.enqueue(MockResponse().withWebSocketUpgrade(object : WebSocketListener() {
            override fun onMessage(webSocket: WebSocket, text: String) {
                if (!text.contains("\"qservant_cancel\"")) return
                val reqId = Regex("\"reqId\":\"([^\"]+)\"").find(text)!!.groupValues[1]
                webSocket.send("""{"t":"qservant_job","reqId":"$reqId","job":{"jobId":"job-1","state":"cancelled"}}""")
            }
        }))
        server.start()
        client = CompanionClient(http)
        client.connect(server.url("/").toString().replace("http", "ws"))
        val collected = java.util.concurrent.CopyOnWriteArrayList<ServerFrame>()
        val job = launch(Dispatchers.Default) { client.frames.collect { collected.add(it) } }
        val frame = withTimeout(3000) { client.awaitReply("qx", ClientMsg.qservantCancel("qx", "job-1")) }
        assertTrue(frame is ServerFrame.QServantJobFrame)
        assertEquals("cancelled", (frame as ServerFrame.QServantJobFrame).job.state)
        withTimeout(3000) { while (collected.none { it is ServerFrame.QServantJobFrame }) delay(20) }
        assertTrue(collected.any { it is ServerFrame.QServantJobFrame })
        job.cancel()
    }
}
