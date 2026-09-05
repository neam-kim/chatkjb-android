# Mail feature

Mail is provided by the repository-owned `mail-host` module under `KJBMail/`.
The sources are ordinary tracked files, not a Git submodule. The Gradle composite
build always resolves this internal directory; external `kjbmail.dir`,
`KJBMAIL_DIR` and sibling-checkout fallbacks are no longer used.

The host app starts `net.thunderbird.android.MailEntryActivity` through
`core/navigation/EmailRoute.kt`. The dependency is declared in
`app/settings.gradle.kts` and `app/app/build.gradle.kts`.

See the root `docs/source-ownership.md` for source provenance and maintenance.
