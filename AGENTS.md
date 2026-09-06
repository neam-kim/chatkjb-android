# Project Agent Instructions

## Android physical-device installation

- Never install this app on a physical Android device with a raw `adb install -r` command.
- For general physical devices, use `scripts/install-android-primary-user.sh <apk> [adb-serial]`. It auto-selects the installed identity (`com.neamkim.chatkjb` for existing phone in-place upgrade, or `com.termux` universal for a new install). When installing `com.neamkim.chatkjb`, it fails closed if any `profile.CLONE` or `profile.MANAGED` user exists.
- For tablet user-0 installations where an existing `profile.CLONE` exists and must not be altered, use the distinct procedure `scripts/install-android-universal-user0.sh <apk> [adb-serial]`. It installs only `com.termux` for user 0 and leaves the tablet clone profile intact.
- Before and after a physical phone install of `com.neamkim.chatkjb`, verify `cmd user list -v` contains no clone or managed profile and verify `com.neamkim.chatkjb` is installed for user 0.
- `dev.herdr.mobile` belongs to a third-party Google Play app; never install a ChatKJB APK under that package ID again.
- Do not remove or modify Samsung Secure Folder (`profile.PRIVATE`) while enforcing this policy.

## Embedded Herdr transport

- The production relay is loopback-only at `127.0.0.1:8375` and is exposed only inside the tailnet by Tailscale Serve HTTPS on `neam-macmini.taild81d38.ts.net:8443`.
- Pair the embedded client with a `relay=` setup fragment only. Never add `gateways=`, enable Funnel, or configure Cloudflare/WebRTC/STUN/UPnP/PCP/NAT-PMP for ChatKJB.
- Never commit, print, or capture the relay setup fragment or token. Seed it through the local `kimjb://open/chat#...` deep link so the embedded client imports it into origin-local storage.
- Preserve the existing Tailscale Serve mappings on ports `10110` and `8787`, and do not modify the separate `net.neam.herdr-mobiled` service.
