# KJBMail security

## Reporting vulnerabilities

Please report suspected vulnerabilities privately through [GitHub Security Advisories for KJBMail](https://github.com/neam-kim/KJBMail/security/advisories/new). If that form is unavailable, open an issue only for non-sensitive coordination and do not include credentials, tokens, private keys, email addresses, or message contents.

Please include the affected version or commit, reproduction steps that do not disclose personal data, and the potential impact. We will acknowledge reports when possible and coordinate disclosure after a fix is available.

## Build and signing verification

KJBMail signing certificates are not published in this repository until an official public release establishes the fingerprints. Verify downloaded artifacts against the checksums and release information published in the [KJBMail Releases](https://github.com/neam-kim/KJBMail/releases) page. For local APK inspection:

```bash
apksigner verify -v --print-certs <path-to-apk>
```

The repository retains upstream security and attribution material in source history and documentation where applicable. Upstream Thunderbird for Android and K-9 Mail assessments or fingerprints must not be interpreted as KJBMail assurances.
