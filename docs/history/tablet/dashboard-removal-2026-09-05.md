> Historical tablet snapshot before the 2026-09-06 consolidation. Paths and validation below describe that earlier checkout; see [current architecture](../../device-compatibility-architecture.md).

# Unused Compose dashboard removal

The tablet ChatKJB menu continues to launch native TermuxActivity. Removed its unused alternate Compose dashboard, view model, companion WebSocket client/protocol/repository, tree/actions, remote terminal adapter, exclusive font and tests. The launcher theme moved unchanged to core/ui/theme (unused status/avatar helpers removed). Removed exclusive Gradle dependencies/catalog entries. The manifest-registered push receiver remains; its JSON configuration now lives beside PushPayload, preserving unknown-key behavior. Existing saved data, server sources and running infrastructure remain intact.

44 removed files are recoverable in `/Users/neam/.Trash/ChatKJB-dashboard-removal-20260905-01a0706d`, preserving project-relative paths. They were restored byte-for-byte from the pre-removal Git index after the user requested Trash handling. No scratch directory was created.

## Validation

`./gradlew :app:testDebugUnitTest :app:assembleDebug` with JDK21 and the configured SDK: BUILD SUCCESSFUL, 18 tests, zero failures/errors/skips. Removed 93 tests that exclusively exercised retired code. This was an incremental build, not a fresh isolated build.

Installed successfully to Galaxy Tab SM-X620 user 0. APK SHA256: `d8e5d89f36670901eddbc2d08d91da8e8ec8210e704abfff1d724dfd9e213520`.

Android MCP screenshot-backed interaction verified launcher → ChatKJB → native Termux → populated live Herdr, keyboard-hide/back → launcher; Server → embedded Moonlight PC list/back; Email → populated existing inbox/back. Final state is launcher. No prompt input was modified or submitted. Crash buffer filtered to com.termux returned count 0. The MCP has no install capability, so SDK adb was used only for the authorized APK installation.

Site/Finance, compact widths, incoming notification delivery and actual streaming were not repeated in this removal pass. Earlier broader checks are historical evidence in ui-verification-2026-09-05.md, not verification of this APK.

Independent combo review: PASS_WITH_CONCERNS, no blockers. Limitations: incremental build and scoped runtime/reference checks. Trash contents were subsequently rechecked against the pre-removal index with SHA256 receipts in evidence/dashboard-removal-trash-2026-09-05.json. Prior session state had all pre-removal source edits staged; the index includes the earlier extracted files.
