package dev.herdr.mobile.features.qservant

import org.junit.Assert.*
import org.junit.Test

class QServantProtocolTest {
    @Test fun catalogNullQuotaAndUnknownFields() {
        val c = parseQServantFrame("""{"t":"qservant_catalog_result","reqId":"r1","quota":null,"future":true,"models":[{"id":"m","label":"Model","efforts":["low","high"],"unknown":1}]}""") as dev.herdr.mobile.features.chat.net.ServerFrame.QServantCatalogResult
        assertEquals(1, c.models.size)
        assertEquals("m", c.models.single().id)
    }

    @Test fun modelChangeConstrainsEffort() {
        val c = listOf(dev.herdr.mobile.features.chat.net.QServantModel("a", "A", listOf("low")), dev.herdr.mobile.features.chat.net.QServantModel("b", "B"))
        assertEquals(listOf("low"), compatibleEfforts(c, "a"))
        assertTrue(compatibleEfforts(c, "b").isEmpty())
    }

    @Test fun submitUsesV8AndAudioMp4() {
        val raw = QServantClientFrame.submit("r1", "a", null, byteArrayOf(1, 2))
        assertTrue(raw.contains("qservant_submit"))
        assertTrue(raw.contains("audioMime"))
        assertTrue(raw.contains("audioBase64"))
        assertFalse(raw.contains("\"v\""))
        assertTrue(raw.contains("\"effort\":\"\""))
    }

    @Test fun nestedJobPayloadPreservesStructuredReport() {
        val frame = parseQServantFrame("""{"t":"qservant_job","reqId":"r2","job":{"jobId":"j1","state":"completed","transcript":"hello","report":{"request":"r","work":"w","verification":"v","changes":["c"],"commit":"abc","result":"ok","success":true}}}""") as dev.herdr.mobile.features.chat.net.ServerFrame.QServantJobFrame
        assertEquals("j1", frame.job.jobId)
        assertEquals("hello", frame.job.transcript)
        assertEquals("w", frame.job.report?.work)
        assertEquals(listOf("c"), frame.job.report?.changes)
        assertTrue(frame.job.report?.success == true)
    }
}
