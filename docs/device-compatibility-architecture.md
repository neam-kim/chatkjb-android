# Device Compatibility and Architecture Policy

ChatKJB unifies personal productivity surfaces (Homepage, KJBMail, Finance, Moonlight Server, and ChatKJB agent control) across phone and tablet devices using a single repository and common runtime shell.

## Package Variants

Two package variants support seamless updates without data loss:

1. **`universal` (`com.termux`)**:
   - Default artifact for new installs on any device.
   - Version code `>= 119`, version name `0.120.0-chatkjb`.
   - Signed with the established tablet test key (`testkey_untrusted.jks`).
   - Includes the Termux app runtime, terminal emulator/view modules, D2Coding monospace font asset, and Korean IME composing preview support.
   - Supports both runtime backends (embedded Herdr and native Termux).

2. **`legacyPhone` (`com.neamkim.chatkjb`)**:
   - Compatibility artifact for in-place upgrades of existing phone installations, preserving all local application data and relay secrets.
   - Version code `>= 4`, version name `1.2.2`.
   - Uses standard Android debug signing lineage.
   - Excludes Termux activities, services, providers, authorities, bootstrap files, and native libraries so it can coexist on the same system without provider conflicts.
   - `NATIVE_TERMUX_AVAILABLE` is set to `false` at compile time.

## Runtime Device Classification

Device classification is performed dynamically at the Activity boundary without hardcoding device models, serial numbers, or package names:

- **Maximum Window Metrics**: Classification inspects AndroidX `WindowManager` maximum window bounds for the Activity's current display.
- **Shortest Edge**: The shortest edge in density-independent pixels (`dp`) is evaluated against `TABLET_MIN_SHORTEST_EDGE_DP` (600 dp):
  - `>= 600 dp`: `DeviceClass.TABLET`
  - `< 600 dp`: `DeviceClass.PHONE`
- **Foldables**: Foldables automatically resolve according to their currently active display (folded outer screen classifies as phone, unfolded inner screen classifies as tablet).
- **Session Stability**: The resolved `DeviceClass` is persisted into saved instance state (`chatkjb.device_class` and `chatkjb.device_class_display`). Rotation and multi-window resizing preserve the existing session's backend, preventing disruptive layout flips. A new Activity on a different display resolves fresh metrics.

## Backend Selection Policy

The pure function `ChatBackendPolicy.backendFor(deviceClass, nativeTermuxAvailable)` determines the active backend:

| Device Class | Native Termux Available | Active Backend |
| :--- | :--- | :--- |
| `PHONE` | `true` (universal) | `EMBEDDED_HERDR` |
| `PHONE` | `false` (legacyPhone) | `EMBEDDED_HERDR` |
| `TABLET` | `true` (universal) | `NATIVE_TERMUX` |
| `TABLET` | `false` (legacyPhone) | `UNIVERSAL_UPGRADE_REQUIRED` |

- **Universal Tablet**: Launches native `com.termux.app.TermuxActivity` directly in-process.
- **Legacy Phone on Tablet**: Fails closed by presenting an explicit upgrade screen (`NativeBackendUnavailableScreen`), advising the user to install the universal package rather than silently falling back to embedded Herdr.

## Deep-Link and Launcher Navigation

- **Launcher Order**: The canonical launcher button order is strictly:
  1. `Site`
  2. `Email`
  3. `Finance`
  4. `Server` (in-process Moonlight PC view)
  5. `ChatKJB`
- **Deep Link Routes**: Supported URIs include `kimjb://open/{home,email,chat,notifications,server}` and legacy `kjbmail://open`. When `kimjb://open/chat` is opened with an encoded setup fragment, the fragment is routed only to embedded Herdr and is never persisted or logged.

## Physical Device Installation Policy

### 1. Primary User Installation (`scripts/install-android-primary-user.sh`)

- Automatically detects whether `com.neamkim.chatkjb` or `com.termux` is already installed for user 0.
- Accepts an APK file for explicit compatibility, an APK directory, or no path. A directory is searched by APK application ID; no path searches the standard `app/app/build/outputs/apk` tree. Use `--serial SERIAL` when selecting a device without supplying an APK path.
- If `com.neamkim.chatkjb` is present, `legacyPhone` is selected to preserve app data.
- If neither is present, `universal` (`com.termux`) is selected as the default.
- Explicit APK files are checked against the selected identity and fail closed on mismatch or unknown application IDs.
- For `legacyPhone`, the script strictly fails closed if any secondary business profile (`profile.CLONE` or `profile.MANAGED`) is detected before or after installation.

### 2. Tablet User-0 Procedure (`scripts/install-android-universal-user0.sh`)

- Specifically dedicated to installing `universal` (`com.termux`) for primary user 0.
- Leaves any pre-existing tablet clone profile completely untouched without modifying or removing it.
- Installs only with `--user 0` and verifies that the package is present for user 0.

## Continuous Integration

GitHub Actions (`.github/workflows/ci.yml`) builds and tests both variants in parallel:
- `Universal`: runs `:app:testUniversalDebugUnitTest` and `:app:assembleUniversalDebug`
- `LegacyPhone`: runs `:app:testLegacyPhoneDebugUnitTest` and `:app:assembleLegacyPhoneDebug`
- Caches Gradle caches, wrappers, Android build-cache, KJBMail Gradle cache, and Termux app Gradle cache across runs.

The local mock matrix can be run without a device:

```bash
scripts/test-installer-policy.sh
```

The universal manifest retains `android:sharedUserId="com.termux"` from the installed tablet app; omitting it prevents an in-place update with `INSTALL_FAILED_UID_CHANGED`. The legacy phone flavor has no shared user ID.
