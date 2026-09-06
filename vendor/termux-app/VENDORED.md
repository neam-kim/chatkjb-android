# Embedded Termux source

Imported from the tablet ChatKJB working tree during Android consolidation on 2026-09-06. The app, shared runtime, terminal emulator and terminal view production sources retain their tablet behavior; upstream tests, licenses and relevant documentation are retained.

Build through the owning repository's `app/gradlew`; this directory is not a second standalone app project. Upstream release automation, app-store artwork and duplicate Gradle wrappers are omitted. `app/build.gradle` downloads the pinned bootstrap packages and verifies their SHA-256 digests. Downloaded ZIPs and native/compiler outputs are local build assets ignored by Git.

The `universal` host retains `com.termux`, its bootstrap prefix and the existing untrusted test signing identity. The `legacyPhone` host excludes the Termux runtime components and native libraries. Runtime display classification selects the ChatKJB destination in the shared host.
