> Historical tablet snapshot before the 2026-09-06 consolidation. Paths and validation below describe that earlier checkout; see [current architecture](../../device-compatibility-architecture.md).

# Tablet source ownership and maintenance

## One repository, repository-local builds

`ChatKJB-android-tab` owns the sources for the host, KJBMail, Termux, Moonlight and companion. Clone this repository normally; no recursive submodule fetch or sibling source checkout is required.

The root `gradlew` / `gradlew.bat` delegates to `app/`. `app/settings.gradle.kts` includes only paths inside this repository. KJBMail remains an internal composite build because it has its own convention plugins, version catalog and API/internal module boundaries. Flattening those boundaries is unnecessary to remove the external checkout dependency.

KJBMail was previously represented by gitlink `5082a97c66aa76d447ed8c5d4e5111db37cdb3ad`, even though this working tree had source files but no nested `.git` metadata. The integration retains that existing source snapshot directly; the old gitlink is provenance, not a claim that every local file matches that revision. Its LICENSE, NOTICE and source headers remain intact. `KJBMail/metadata` is a relative symlink into its own `app-metadata/` directory.

Termux and Moonlight were already embedded. Moonlight's generated `build/` files are now excluded from source control; local build products are left in place. Shared Gradle/SDK caches and normal Maven dependencies remain external tools, not source checkouts.

## Where to make changes

| Concern | Owner |
| --- | --- |
| Android lifecycle, native activity launch | `app/app/.../integration/MainActivity.kt` |
| Launcher tiles | `app/app/.../features/homepage/KimJbLauncher.kt` |
| Site/Finance WebView and navigation rules | `features/homepage/HomepageWebScreen.kt`, `core/navigation/` |
| Native tablet terminal | `vendor/termux-app/` |
| Email features and dependency wiring | `KJBMail/feature/`, `KJBMail/app-common/`, `KJBMail/mail-host/` |
| Streaming and JNI | `app/moonlight/` |
| Shared launcher theme | `core/ui/theme/` |
| Legacy incoming notifications | `core/push/`, `core/data/Settings.kt` |

Paths abbreviated as `features/` and `core/` are under `app/app/src/main/java/dev/herdr/mobile/`. The tablet launcher opens native `TermuxActivity` from `vendor/termux-app`. The unused Compose Herdr dashboard, its companion client/model, remote terminal adapter and exclusive tests/font have been removed. Shared launcher theme and manifest-registered push receiver remain. Existing stored preferences are not erased. The standalone `companion/` server and live infrastructure are outside this client cleanup.

## Validation

Run `./gradlew :app:testDebugUnitTest :app:assembleDebug` from the repository root. To verify source independence, export the candidate tracked tree into a private directory outside the repository and run the same command with `ANDROID_HOME` and JDK 21. Do not copy `local.properties`, `.gradle` or `build` directories from the working checkout. No device installation or remote service restart is part of a source-only refactor.
