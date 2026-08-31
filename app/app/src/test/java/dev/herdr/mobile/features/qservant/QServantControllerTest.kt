package dev.herdr.mobile.features.qservant


import dev.herdr.mobile.features.chat.net.QServantModel
import dev.herdr.mobile.features.chat.net.ServerFrame
import org.junit.Assert.assertEquals
import org.junit.Test

class QServantControllerTest {
    @Test fun jobAndReportTransitions() {
        var s = QServantState()
        s = reduceQServant(s, dev.herdr.mobile.features.chat.net.ServerFrame.QServantJobFrame("r1", dev.herdr.mobile.features.chat.net.QServantJob("job-1", "recorded")))
        assertEquals(QServantJobState.Recorded, s.jobState)
        s = reduceQServant(s, dev.herdr.mobile.features.chat.net.ServerFrame.QServantJobFrame("r2", dev.herdr.mobile.features.chat.net.QServantJob("job-1", "running")))
        assertEquals(QServantJobState.Running, s.jobState)
        s = reduceQServant(s, dev.herdr.mobile.features.chat.net.ServerFrame.QServantJobFrame("r3", dev.herdr.mobile.features.chat.net.QServantJob("job-1", "completed")))
        assertEquals(QServantJobState.Completed, s.jobState)
    }

	@Test fun catalogChoosesCompatibleConfiguredDefault() {
		val frame = ServerFrame.QServantCatalogResult(
			reqId = "catalog",
			models = listOf(
				QServantModel("m1", "First", listOf("low", "high"), "low"),
				QServantModel("m2", "Second", listOf("medium", "high"), "medium"),
			),
			defaultModel = "m2",
			defaultEffort = "high",
		)
		val state = reduceQServant(QServantState(), frame)
		assertEquals("m2", state.selectedModel)
		assertEquals("high", state.selectedEffort)
	}

	@Test fun modelChangeRepairsIncompatibleEffort() {
		val controller = QServantController()
		controller.onFrame(ServerFrame.QServantCatalogResult(
			reqId = "catalog",
			models = listOf(
				QServantModel("m1", "First", listOf("low", "high"), "high"),
				QServantModel("m2", "Second", listOf("medium"), "medium"),
			),
			defaultModel = "m1",
			defaultEffort = "high",
		))
		controller.selectModel("m2")
		assertEquals("medium", controller.state.value.selectedEffort)
	}
}
