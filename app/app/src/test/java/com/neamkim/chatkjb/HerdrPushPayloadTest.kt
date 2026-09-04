package com.neamkim.chatkjb

import com.neamkim.chatkjb.features.herdr.push.parseHerdrPush
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class HerdrPushPayloadTest {
    @Test
    fun parsesFinishedPayloadAndIgnoresFutureFields() {
        val payload = parseHerdrPush(
            """{"kind":"finished","paneId":"w1:p1","workspaceId":"w1","title":"Done","body":"Tests passed","future":true}"""
                .encodeToByteArray(),
        )
        assertEquals("finished", payload?.kind)
        assertEquals("w1:p1", payload?.paneId)
        assertEquals("Tests passed", payload?.body)
    }

    @Test
    fun rejectsMalformedPayload() {
        assertNull(parseHerdrPush("not-json".encodeToByteArray()))
    }
}
