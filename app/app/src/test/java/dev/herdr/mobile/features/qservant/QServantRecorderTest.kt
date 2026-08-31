package dev.herdr.mobile.features.qservant

import org.junit.Assert.assertEquals
import org.junit.Test

class QServantRecorderTest {
    @Test fun permissionAndEncodingFailuresAreTruthful() {
        assertEquals(RecordingFailure.PermissionDenied, recordingFailureState(false))
        assertEquals(RecordingFailure.EncodingFailed, recordingFailureState(true, fileBytes = 0))
        assertEquals(RecordingFailure.TooLarge, recordingFailureState(true, fileBytes = Q_SERVANT_MAX_AUDIO_BYTES + 1))
    }
}
