# Source ownership and maintenance

ChatKJB Android is one source repository. A normal clone contains all project-owned
sources required for the Android build and for rebuilding the embedded Herdr bundle.
There are no Git submodules or sibling-checkout fallbacks. Public package registries
(Maven, Bun and Go) and the Android/JDK toolchains remain build dependencies.
The Mac Herdr service and Tailscale remain runtime services; this source migration
does not redeploy or reconfigure them.

## Where to change code

| Location | Responsibility |
| --- | --- |
| `app/app/src/main/java/com/neamkim/chatkjb/integration/MainActivity.kt` | Android lifecycle, permission registration and launching embedded activities |
| `app/app/src/main/java/com/neamkim/chatkjb/integration/ChatKjbApp.kt` | Compose screen transitions and launcher state |
| `app/app/src/main/java/com/neamkim/chatkjb/core/navigation/` | Destination contract and separately named route policies |
| `app/app/src/main/java/com/neamkim/chatkjb/core/web/` | Shared Activity lookup and private WebView settings |
| `app/app/src/main/java/com/neamkim/chatkjb/features/` | Screen-specific behavior, navigation clients, cookies and attachment lifecycle |
| `KJBMail/` | Repository-owned mail implementation and Gradle build logic |
| `embedded/herdr-mobile-relay/` | Repository-owned Herdr frontend, relay, protocol fixtures and tests |
| `app/app/src/main/java/com/limelight/`, `app/app/src/main/jni/` | Already-vendored Moonlight Java/native implementation |

KJBMail retains its internal Gradle composite boundary to preserve its convention
plugins, version catalogs and mail module graph. `app/settings.gradle.kts` always
includes `../KJBMail`; no environment or property can switch to another checkout.

The homepage deliberately retains its separate OAuth cookie and viewport settings.
Only the identical settings for Herdr and the management consoles are shared.
Route allowlists, the embedded CSP, file chooser cleanup and back behavior remain
screen-owned. Do not unify those policies merely because the screens use WebView.

## Updating sources

Edit the in-repository source and include the changes in the same review as the host.
For upstream imports, record the upstream commit and preserve all licenses and headers
in `THIRD-PARTY-NOTICES.md`. Do not copy `.git`, local properties, credentials, build
outputs or package caches from another checkout.

Run `scripts/sync-embedded-herdr.sh` from any directory after Herdr frontend changes.
It accepts no external source path, installs with the frozen lockfile and validates
before replacing the checked-in Android assets. Commit source and generated assets
together. CI rebuilds and checks that the generated assets match the committed files.
The imported upstream Makefile contains deployment options for other deployments;
ChatKJB uses the tailnet-only wrapper documented in the root README and AGENTS.md.

## Verification

```sh
scripts/sync-embedded-herdr.sh
(cd app && ./gradlew :app:testDebugUnitTest :app:assembleDebug)
(cd embedded/herdr-mobile-relay && go test ./... && go build ./cmd/herdr-mobile-relay)
```

Android UI verification remains a separate device step. Source/build checks do not
prove live OAuth, file picker, Moonlight pairing or Tailscale connectivity.

## Verification receipt — 2026-09-05

- KJBMail: 7,937 tracked file blobs and executable modes match commit
  `5082a97c66aa76d447ed8c5d4e5111db37cdb3ad` exactly.
- Herdr: 547 tracked file blobs and executable modes match commit
  `90ad7a26fb5c32b55662601c7b3b427e8de893f9` exactly.
- No Git submodule entries remain. The historical KJBMail Git object database is
  retained locally under the parent repository's `.git/modules` for recovery;
  it is not part of the source tree and is not required to build.
- `scripts/sync-embedded-herdr.sh`: lint and type checks passed; 287 tests in 16
  files passed; build and size gates passed. Generated Android assets match the
  previously committed assets byte-for-byte.
- `:app:testDebugUnitTest :app:assembleDebug`: 17 tests passed and APK built.
  Repeating with both `KJBMAIL_DIR=/nonexistent/external-mail` and
  `-Pkjbmail.dir=/nonexistent/external-mail` also passed.
- Exported the staged Git tree to a disposable directory without `.git`, local
  properties, package caches or existing build outputs. The same Android tests
  and APK build passed there using JDK 21 and the installed SDK. The first attempt
  exposed the old 2GB heap limit; `app/gradle.properties` now uses 4GB and four
  workers. The successful retry completed 1,727 tasks, with 542 up-to-date from
  that isolated attempt, and did not use the original checkout's build outputs.
- Full `go test ./...` and `go build -o bin/herdr-mobile-relay
  ./cmd/herdr-mobile-relay` passed. On macOS the tests need a short, canonical
  temporary path (`TMPDIR=/private/tmp/<short-task-directory>`): `/var` aliases
  break path comparisons and long temporary paths exceed Unix socket limits.
  `TestDiscardPersistsWithoutLeavingTombstone` failed once and passed on retry
  without source changes; this remains an observed upstream test flake.
- No physical-device installation, live-service deployment or UI verification was
  performed. Generated UI assets and native route policies were preserved.

Independent `combo/Reviewer` review returned `PASS_WITH_CONCERNS` with no blockers.
The clean-export build completed successfully after the review request. CI now also
rejects untracked generated assets and explicitly creates its binary output directory.
The observed upstream test flake and lack of physical-device UI verification remain
recorded limitations.

Follow-up physical-device UI verification and the launcher status-bar correction
are recorded in [ui-verification-2026-09-05.md](ui-verification-2026-09-05.md).
