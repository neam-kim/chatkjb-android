# Third-party notices

ChatKJB is based on herdr-mobile and is licensed under **AGPL-3.0-or-later** (see [`LICENSE`](LICENSE)).
It bundles the following third-party components, each under its own license.

## Termux terminal emulator and view

- **Modules:** `app/terminal-emulator/` (`com.termux.terminal`) and
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

## D2Coding

- **File:** `app/app/src/main/assets/fonts/D2Coding-Regular.ttf`
- **Source:** [naver/d2-coding-font](https://github.com/naver/d2-coding-font)
- **License:** SIL Open Font License 1.1 — full text in `vendor/fonts/OFL.txt` and `app/app/src/main/assets/fonts/OFL.txt`

## Moonlight Android (transplanted in-process client)

- **Component:** `app/moonlight/`, integrated from the official `moonlight-stream/moonlight-android` repository
- **Upstream revision:** `98c12bebffac592eb57cf25e9a4638b40aa2c17d` (`Update to OkHttp 5.5`), observed 2026-09-05
- **Recursive submodule:** `moonlight-common-c` at `874ac9548f1bd6f095ef2b435c42cdde460e7821`
- **Source:** https://github.com/moonlight-stream/moonlight-android
- **License:** GPL-3.0-only or GPL-3.0-or-later; the bundled `moonlight-common-c` is GPL-3.0-or-later
- **License text:** [`app/moonlight/src/main/LICENSE.txt`](app/moonlight/src/main/LICENSE.txt)
- **Attribution:** Copyright Cameron Gutman, Diego Waxemberg, Aaron Neyer, and Moonlight contributors. Upstream source headers are retained.

This component is combined with ChatKJB's AGPL-3.0-or-later application under the GPLv3/AGPLv3 compatibility provision. Corresponding source for the combined work is retained in this project tree.
