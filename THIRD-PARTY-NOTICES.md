# Third-party notices

ChatKJB is licensed under **AGPL-3.0-or-later** (see [`LICENSE`](LICENSE)).
It bundles the following third-party components, each under its own license.

## Herdr Mobile Relay embedded frontend

- **Bundled output:** `app/app/src/main/assets/herdr/`
- **Source:** [0cv/herdr-mobile-relay](https://github.com/0cv/herdr-mobile-relay), based on commit `7400537`
- **Local source branch:** `/Volumes/NEAM_SSD/herdr-mobile-relay` branch `chatkjb-embedded`
- **License:** AGPL-3.0-or-later
- **Copyright:** © 2026 Christophe Vidal and contributors

ChatKJB modifies the product name, default visual theme, legal/source links,
asset paths, and a deployment-specific Content Security Policy for an
Android-embedded build. The encrypted WebSocket and Herdr protocol behavior
remain upstream-compatible; this deployment restricts ingress to its private
Tailscale tailnet and does not activate the optional gateway/WebRTC path.

## Termux terminal emulator and view source

- **Retained source modules (not linked into the current APK):** `app/terminal-emulator/` (`com.termux.terminal`) and
  `app/terminal-view/` (`com.termux.view`)
- **Source:** [termux/termux-app](https://github.com/termux/termux-app)
- **License:** GPL-3.0-only. These modules incorporate code from Jack Palevich's
  *Terminal Emulator for Android*, originally released under Apache-2.0.
- **Copyright:** © Fredrik Fornwall and the Termux contributors; portions
  © Jack Palevich and the Android Open Source Project.

GPL-3.0 and Apache-2.0 are both compatible with this project's AGPL-3.0-or-later
license. The upstream files retain their original headers where present.

## JetBrains Mono

- **File:** `app/app/src/main/assets/fonts/JetBrainsMono-Regular.ttf`
- **Source:** [JetBrains/JetBrainsMono](https://github.com/JetBrains/JetBrainsMono)
- **License:** SIL Open Font License 1.1 — full text in
  [`app/app/src/main/assets/fonts/OFL.txt`](app/app/src/main/assets/fonts/OFL.txt)

## Runtime dependencies

Go and Gradle dependencies (e.g. `github.com/coder/websocket`,
`github.com/creack/pty`, AndroidX, Jetpack Compose) are fetched at build time and
governed by their respective licenses as declared in `companion/go.mod` and the
Gradle version catalog (`app/gradle/libs.versions.toml`).

## Moonlight Android client (in-process Server entry)

- **Bundled source:** `app/app/src/main/java/com/limelight/`, `app/app/src/main/jni/`, and `app/app/src/main/res/`
- **Source:** [moonlight-stream/moonlight-android](https://github.com/moonlight-stream/moonlight-android)
- **Pinned commit:** `98c12bebffac592eb57cf25e9a4638b40aa2c17d` (`Update to OkHttp 5.5`), observed 2026-09-05
- **Submodule:** `moonlight-common-c` at `874ac9548f1bd6f095ef2b435c42cdde460e7821`
- **License:** GNU GPL v3.0-or-later; upstream notices and headers are retained in the transplanted source
- **Copyright:** Cameron Gutman, Diego Waxemberg, Aaron Neyer, and Moonlight contributors

ChatKJB starts `com.limelight.PcView` in-process from the launcher’s `Server` entry while retaining the host application ID `com.neamkim.chatkjb`. The transplanted provider is namespaced to `poster.com.neamkim.chatkjb`; no `com.limelight` package or application is installed or modified on devices. The existing Herdr embedded route and the tailnet management-console routes remain separate.
