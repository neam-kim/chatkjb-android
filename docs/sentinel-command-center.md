# Sentinel actions in Command Center

Tapping the launcher logo opens Command Center. A Sentinel card shows its locally
stored reception timestamp (including time zone), **원인 조사**, and **Done**.
Reception time is not claimed to be the original host event time.

**원인 조사** opens the embedded Herdr connection and requests one Codex investigation
in the focused existing Space. The fixed preset is `gpt-5.6-sol` with
`model_reasoning_effort="high"`, working in `/Volumes/NEAM_SSD/security-sentinel`.
The initial prompt identifies the alert number and reception time and requests a
read-only investigation with sensitive-value redaction. The agent must distinguish
historical evidence from a newer current state. The resulting tab can be opened
from the confirmation screen.

The authenticated relay advertises `sentinel_investigation`. Older relays cannot
silently launch the default model. The generic agent launcher retains its existing
home-directory guard; only this fixed preset may use the verified canonical
Sentinel directory. Missing/moved/symlinked canonical paths fail closed.

Native automatic dispatch is authorized only by the internal card button. A URL
fragment alone requires a visible web-button action. Native authorization is issued
once per WebView instance. Launch receipts are persisted before dispatch; repeated
entry does not automatically dispatch a second request. Both receipt keys and
existing-agent matching include the Space. If completion is uncertain, inspect the
Herdr tab list instead of blindly retrying. There is no automatic retry/reset UI.

**Done** removes only the matching current card, persists acknowledgement, and
cancels that card's Android notification ID. A preference listener updates Compose
immediately. An identical delayed delivery stays acknowledged; a new monotonic
Sentinel problem number is a new alert. Acknowledgement is local to this device and
does not change Sentinel's host baseline, health, or findings.

## Verification, 2026-09-06

- S26 guarded user-0 install passed; no clone/work profile, Secure Folder preserved.
- Actual No.21 showed reception time and both buttons. One native tap created
  `w1P:tD` / `w1P:pE` in existing focused Space `w1P`; Herdr and the phone terminal
  showed `gpt-5.6-sol high` and the cause-investigation prompt. This real
  investigation tab was retained for the owner.
- Done tested using the local-only `SentinelInboxInstrumentation` fixture: visible
  No.900003 card disappeared immediately after the actual tap. Fresh-process
  verification confirmed persistent acknowledgement, duplicate-delivery suppression, matching Android OS notification cancellation,
  and restoration of pre-test preferences. The separate test APK was uninstalled.
- Frontend: 292 tests, lint, type check, production build, size gate passed. Ego
  browser fixtures exercised success, terminal navigation, duplicate re-entry,
  disconnected state, and uncertain failure at desktop and 390px mobile widths.
- Android unit tests, production APK and instrumentation APK builds passed.
- Go Sentinel admission/argv/lifecycle tests passed, including canonical SSD
  admission and preservation of the generic home-only restriction. Go build passed.
- The broader affected Go packages are not all green on this host. Original HEAD
  source supplied through a Go overlay reproduces the short-deadline refusal,
  inter-key, and child-PID timing failures. Long macOS Unix-socket temporary paths
  also require a short task-local TMPDIR. These unrelated failures were not hidden
  or repaired by relaxing this feature's tests.
- Single-bundle frontend budget intentionally increased from 129 to 131 KiB;
  measured final payload is approximately 133.4 KB gzip. Asset revision 308.
- The old tablet canonical path recorded in wiki is absent. This change belongs
  to the available phone project; tablet installation was not performed.

Runtime release:
`~/.local/share/herdr-mobile-relay/releases/0.19.1-chatkjb-sentinel-20260906-darwin-arm64`.
Only `com.neamkim.chatkjb.herdr-relay` was restarted. The old release is retained
for rollback; separate legacy relay and Tailscale mappings were not changed.

Independent final review: **PASS_WITH_CONCERNS**, no blockers. Remaining concerns
are the baseline-reproduced Go timing failures and deliberately conservative
non-retry behavior. The review also raised a conditional clear-notification concern:
the actual `HerdrUnifiedPushReceiver` explicitly routes `clear`, `sentinel-clear`,
and `skill-clear` to cancellation and calls `post` only in the else branch, so
that condition is not present in the production receiver.
