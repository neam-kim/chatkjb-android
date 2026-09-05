# KJBMail

[![Latest release](https://img.shields.io/github/v/release/neam-kim/KJBMail?style=for-the-badge)](https://github.com/neam-kim/KJBMail/releases/latest)
[![License](https://img.shields.io/github/license/neam-kim/KJBMail)](LICENSE)
[![Issues](https://img.shields.io/github/issues/neam-kim/KJBMail)](https://github.com/neam-kim/KJBMail/issues)

KJBMail is an open-source Android email client focused on private, user-controlled mail. It can manage multiple accounts and supports the mail protocols and account providers implemented in this repository, including IMAP and POP3. Available functionality must be verified in the app and release notes; this fork does not promise services that are not implemented here.

## Project status and downloads

KJBMail is developed in the [`neam-kim/KJBMail`](https://github.com/neam-kim/KJBMail) repository. Releases, checksums, known issues, and support requests are published there:

- [Releases](https://github.com/neam-kim/KJBMail/releases)
- [Issues and feature requests](https://github.com/neam-kim/KJBMail/issues)
- [Security reporting](SECURITY.md)

There are no KJBMail store listings or release channels documented here until they are published by this project. Do not install a build from an upstream listing and assume it is a KJBMail build.

## Repository structure

- `app-thunderbird`: KJBMail application entry point and product wiring (the module name is retained for compatibility).
- `app-common`: shared application integration and dependency wiring.
- `feature`, `core`, `components`, `mail`, `backend`: user-facing features, shared infrastructure, UI, and mail protocol implementations.
- `mail-host`: KJBMail host/branding resources and application metadata used by the product build.
- `app-k9mail` and `legacy`: retained upstream-compatible application and legacy code; they are not the KJBMail product identity.
- `docs`: engineering, architecture, contributor, release, and user documentation.

## Building and developing

A recent Android development environment is required:

- JDK 21 or newer (the build rejects older Java versions).
- Android SDK installed, with the SDK/platform and build-tools versions selected by the Gradle configuration.
- macOS, Linux, or Windows with enough disk space for Gradle and Android dependencies.

From the repository root:

```bash
./gradlew :app-thunderbird:assembleDebug
./gradlew test
./gradlew lint detekt spotlessCheck
```

Use `./gradlew tasks` to discover additional tasks. Do not commit generated build output, signing keys, OAuth credentials, or personal mail data.

## Fork and OAuth configuration

A KJBMail fork must use its own OAuth client registration and a redirect URI that is different from the upstream and other installed applications. Configure the OAuth factories for the product's debug and release variants in the relevant `app-thunderbird` source sets. Keep client identifiers and secrets in local/CI secret configuration; never commit them. A fork also needs its own application ID, signing certificate, store listing, and release process before public distribution. The existing package and namespace identifiers are intentionally retained for compatibility and are not a branding claim.

## Privacy and security

KJBMail is intended to keep mail access under the user's control. Credentials, tokens, and message contents must not be logged or sent to an analytics service. Mail is transferred to the configured providers using the protocols and security settings supported by the app; provider terms and server configuration still apply. Review [SECURITY.md](SECURITY.md) before reporting a vulnerability. This repository is not a substitute for a separately published legal privacy notice; any public distribution should provide one describing its data handling and contact details.

## Upstream relationship and attribution

KJBMail is a fork/derivative work of the open-source Thunderbird for Android project, which itself is based on the long-running K-9 Mail codebase. Upstream architecture, source comments, copyright notices, and documentation remain where they are needed for attribution and compatibility. Upstream names and marks are not the KJBMail product name; references in `legacy`, package names, module names, APIs, schemas, and copyright/license notices are retained intentionally.

KJBMail is licensed under the [Apache License 2.0](LICENSE). The [NOTICE](NOTICE) file retains required upstream attribution, including K-9 Mail and Android Open Source Project copyright notices. Apache licensing does not grant permission to use Thunderbird, K-9 Mail, Mozilla, or other upstream trademarks as KJBMail branding.

## Contributing

Read [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md), follow the repository's Kotlin and Android conventions, and open changes in the KJBMail issue tracker. Please avoid unrelated renames of package, namespace, module, or API identifiers.
