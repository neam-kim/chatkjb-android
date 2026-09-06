# Third-party notices

ChatKJB is licensed under **AGPL-3.0-or-later** (see [`LICENSE`](LICENSE)).
It bundles the following third-party components, each under its own license.

## Herdr Mobile Relay embedded frontend

- **Bundled output:** `app/app/src/main/assets/herdr/`
- **Source:** [0cv/herdr-mobile-relay](https://github.com/0cv/herdr-mobile-relay), based on commit `7400537`
- **Repository-owned source:** `embedded/herdr-mobile-relay/` (frontend, relay, contracts, tests and build tooling)
- **Imported fork commit:** `90ad7a26fb5c32b55662601c7b3b427e8de893f9` (`chatkjb-embedded`), imported 2026-09-05
- **License:** AGPL-3.0-or-later
- **Copyright:** © 2026 Christophe Vidal and contributors

ChatKJB modifies the product name, default visual theme, legal/source links,
asset paths, and a deployment-specific Content Security Policy for an
Android-embedded build. The encrypted WebSocket and Herdr protocol behavior
remain upstream-compatible; this deployment restricts ingress to its private
Tailscale tailnet and does not activate the optional gateway/WebRTC path.

## Termux app, shared runtime, terminal emulator and view

- **Repository-owned source:** `vendor/termux-app/`, including the native app, shared runtime, terminal emulator and terminal view. The universal APK includes this runtime and uses it automatically on tablet-class displays.
- **Source:** [termux/termux-app](https://github.com/termux/termux-app)
- **License:** GPL-3.0-only. These modules incorporate code from Jack Palevich's
  *Terminal Emulator for Android*, originally released under Apache-2.0.
- **Copyright:** © Fredrik Fornwall and the Termux contributors; portions
  © Jack Palevich and the Android Open Source Project.

GPL-3.0 and Apache-2.0 are both compatible with this project's AGPL-3.0-or-later
license. The upstream files retain their original headers where present.

## Runtime dependencies

Go and Gradle dependencies (e.g. `github.com/coder/websocket`,
`github.com/creack/pty`, AndroidX, Jetpack Compose) are fetched at build time and
governed by their respective licenses as declared in `companion/go.mod` and the
Gradle version catalog (`app/gradle/libs.versions.toml`).

## Moonlight Android client (in-process Server entry)

- **Bundled source:** `app/app/src/main/java/com/limelight/`, `app/app/src/main/jni/`, and `app/app/src/main/res/`
- **Source:** [moonlight-stream/moonlight-android](https://github.com/moonlight-stream/moonlight-android)
- **Pinned commit:** `98c12bebffac592eb57cf25e9a4638b40aa2c17d` (`Update to OkHttp 5.5`), observed 2026-09-05
- **Vendored native source (no Git submodule required):** `moonlight-common-c` at `874ac9548f1bd6f095ef2b435c42cdde460e7821`
- **License:** GNU GPL v3.0-or-later; upstream notices and headers are retained in the transplanted source
- **Copyright:** Cameron Gutman, Diego Waxemberg, Aaron Neyer, and Moonlight contributors

ChatKJB starts `com.limelight.PcView` in-process from the launcher’s `Server` entry. Its provider uses `poster.${applicationId}` for the universal and legacy phone package identities. No standalone `com.limelight` application is required. The Herdr embedded route and tailnet management-console routes remain separate.

## KJBMail / Thunderbird for Android

- **Repository-owned source:** `KJBMail/`, including Gradle convention plugins, modules, tests and resources
- **Source:** [neam-kim/KJBMail](https://github.com/neam-kim/KJBMail), derived from Thunderbird for Android
- **Imported commit:** `5082a97c66aa76d447ed8c5d4e5111db37cdb3ad`, imported 2026-09-05 from the previously pinned submodule
- **License:** Apache-2.0 and component-specific notices; see `KJBMail/LICENSE`, `KJBMail/NOTICE`, and original file headers

The source snapshot is unchanged by the submodule-to-directory conversion. It is now tracked by the ChatKJB repository; updates are explicit source changes reviewed and tested together with the host app.

## D2Coding

- **Bundled file:** `app/app/src/main/assets/fonts/D2Coding-Regular.ttf`
- **Source:** [naver/d2-coding-font](https://github.com/naver/d2-coding-font)
- **License:** SIL Open Font License 1.1; retained with the font in `app/app/src/main/assets/fonts/D2Coding-OFL.txt`.

The native Termux host uses D2Coding and preserves the tablet Korean IME composing behavior.
