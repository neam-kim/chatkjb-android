package com.neamkim.chatkjb.core.navigation

/** The runtime window class used to select the ChatKJB backend. */
enum class DeviceClass {
    PHONE,
    TABLET,
}

/** Backends exposed by the shared ChatKJB route. */
enum class ChatBackend {
    EMBEDDED_HERDR,
    NATIVE_TERMUX,
    UNIVERSAL_UPGRADE_REQUIRED,
}

/** Pure device-class/backend policy; it has no package or device identity inputs. */
object ChatBackendPolicy {
    const val TABLET_MIN_SHORTEST_EDGE_DP = 600

    fun deviceClassFor(maximumShortestEdgeDp: Int): DeviceClass =
        if (maximumShortestEdgeDp >= TABLET_MIN_SHORTEST_EDGE_DP) {
            DeviceClass.TABLET
        } else {
            DeviceClass.PHONE
        }

    fun backendFor(deviceClass: DeviceClass, nativeTermuxAvailable: Boolean): ChatBackend =
        when (deviceClass) {
            DeviceClass.PHONE -> ChatBackend.EMBEDDED_HERDR
            DeviceClass.TABLET -> if (nativeTermuxAvailable) {
                ChatBackend.NATIVE_TERMUX
            } else {
                ChatBackend.UNIVERSAL_UPGRADE_REQUIRED
            }
        }

    /** Convert maximum window bounds to the shortest edge in dp. */
    fun maximumShortestEdgeDp(widthPx: Int, heightPx: Int, density: Float): Int {
        require(widthPx >= 0) { "widthPx must be non-negative" }
        require(heightPx >= 0) { "heightPx must be non-negative" }
        require(density > 0f) { "density must be positive" }
        return (minOf(widthPx, heightPx) / density).toInt()
    }
}

/** Restore only for the same display; a new display is freshly classified. */
fun restoreSessionDeviceClass(
    savedClass: String?,
    savedDisplayId: Int?,
    currentDisplayId: Int,
): DeviceClass? {
    if (savedDisplayId != currentDisplayId) return null
    return savedClass?.let { value -> runCatching { DeviceClass.valueOf(value) }.getOrNull() }
}

const val DEVICE_CLASS_STATE_KEY = "chatkjb.device_class"
const val DEVICE_CLASS_DISPLAY_STATE_KEY = "chatkjb.device_class_display"
