# Contributing to ChatKJB

Thanks for your interest. ChatKJB is a small, opinionated companion for
[herdr](https://herdr.dev), and contributions that keep it focused and well-built are
welcome.

## Before you start

- **Bugs:** open an issue using the bug report template with a clear reproduction.
- **Features / larger changes:** open an issue to discuss the idea before writing a big
  PR, so we can agree on scope and direction first.
- **Understand your code.** Using AI to help write code is fine — submitting code you
  can't explain is not. Be ready to describe what your change does and how it behaves at
  the edges.

## Project layout

- `companion/` — the Go daemon (`herdr-mobiled`). Talks to herdr's socket API and serves
  the app over WebSocket.
- `app/` — the Android app (Kotlin + Jetpack Compose), plus the vendored Termux
  `terminal-emulator` / `terminal-view` modules.
- `docs/` — design specs and implementation plans for each feature.

## Development

Requirements: Go 1.23+, JDK 21, Android SDK (compileSdk 37), and NDK 29.0.14206865.

```bash
# Companion: build + test
cd companion
go build ./...
go test ./...
gofmt -l .          # should print nothing

# App: both package variants, unit tests + debug builds
cd app
JAVA_HOME="${JAVA_HOME:?set JDK 21}" \
ANDROID_HOME="${ANDROID_HOME:?set Android SDK}" \
./gradlew :app:testUniversalDebugUnitTest :app:testLegacyPhoneDebugUnitTest \
  :app:assembleUniversalDebug :app:assembleLegacyPhoneDebug
```

The `universal` (`com.termux`) artifact is the default for new installs and
contains the native Termux runtime. The `legacyPhone`
(`com.neamkim.chatkjb`) artifact preserves existing phone installations and
fails closed on tablet-class windows with an upgrade prompt. Runtime selection
uses the active display's maximum window metrics and a 600 dp shortest-edge
boundary; package choice never acts as a user-visible mode switch. Use
`scripts/install-android-primary-user.sh` for automatic identity selection,
or `scripts/install-android-universal-user0.sh` for an explicit tablet
user-0 install that leaves any clone profile untouched. Both scripts accept an
APK file, an APK directory, or no APK path (the standard build output tree is
used); use `--serial SERIAL` when selecting a device without an APK path. The
package application ID is checked before installation.

Run `scripts/test-installer-policy.sh` to exercise the identity/profile matrix
with local mock `adb` and `aapt` binaries. It does not contact a device.

CI runs the same Go and Android checks on every pull request; please make sure they pass
locally first.

## Pull requests

- Keep PRs focused — one logical change per PR.
- Add or update tests for behavior changes.
- Follow the existing code style (`gofmt` for Go; match the surrounding Kotlin/Compose
  idiom).
- Note user-facing changes in [`CHANGELOG.md`](CHANGELOG.md) under `## Unreleased`.
- Don't commit secrets, tailnet IPs, or personal paths.

## Security

Please report security vulnerabilities privately — see [`SECURITY.md`](SECURITY.md).
Do not open a public issue for a vulnerability.

## License

By contributing, you agree that your contributions are licensed under the project's
[AGPL-3.0-or-later](LICENSE) license.
