# Project Agent Instructions

## Android physical-device installation

- Never install this app on a physical Android device with a raw `adb install -r` command.
- Always use `scripts/install-android-primary-user.sh <apk> [adb-serial]`. It installs only for primary user 0 and fails closed when a `profile.CLONE` or `profile.MANAGED` user exists.
- Before and after a physical-device install, verify `cmd user list -v` contains no clone or managed profile and verify `dev.herdr.mobile` is installed for user 0.
- Do not remove or modify Samsung Secure Folder (`profile.PRIVATE`) while enforcing this policy.
