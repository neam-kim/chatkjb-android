# Moonlight Android integration provenance

The tablet app contains an in-process transplant of the official Moonlight Android client at:

- repository: `https://github.com/moonlight-stream/moonlight-android`
- commit: `98c12bebffac592eb57cf25e9a4638b40aa2c17d`
- commit subject: `Update to OkHttp 5.5`
- recursive `moonlight-common-c`: `874ac9548f1bd6f095ef2b435c42cdde460e7821`
- license: GPL-3.0-only or GPL-3.0-or-later, with GPLv3 `moonlight-common-c`

`app/moonlight` is an Android library module. Its merged activities include the upstream `com.limelight.PcView`, `Game`, `AppView`, preferences, and supporting services. The launcher Server action constructs an explicit component intent for `com.limelight.PcView` in the host package (`com.termux`), so it runs in this APK process. It does not load the management-console `/server/` WebView and it does not invoke the standalone `com.limelight` application.

The host application ID is `com.termux`; Moonlight's provider authority is generated as `poster.${applicationId}` (`poster.com.termux`), which remains distinct from standalone Moonlight's provider authority.
