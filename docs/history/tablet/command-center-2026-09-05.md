> Historical tablet snapshot before the 2026-09-06 consolidation. Paths and validation below describe that earlier checkout; see [current architecture](../../device-compatibility-architecture.md).

# Tablet Command Center

The launcher logo opens the same Command Center sections as the phone: AutoBot, Server, Skill Suggestions, and Sentinel. There is no added Back button; Android Back returns from a console to Command Center, then to the launcher. The Command Center uses the phone sections and card styling within a centered column capped at 520dp, matching the tablet launcher row width. All buttons and automation cards share that width.

The launcher Server entry still opens embedded Moonlight, and ChatKJB still opens native Termux. The new console destinations use fixed HTTPS Tailscale URLs on port 8443 and restrict navigation to the selected console path. Private WebViews disable file/content access and mixed content. Console status-bar icons are light and restore on exit.

AutomationInbox persistence is ported from the phone. The existing tablet UnifiedPush receiver stores and clears Skill/Sentinel messages; automation notification taps target Command Center. Existing push registration infrastructure is unchanged. Remote message delivery, populated/clear automation states, and notification taps were not exercised end to end.

## Validation

- JDK 21: `./gradlew :app:testDebugUnitTest :app:assembleDebug --console=plain` passed. 21 tests, zero failures/errors/skips. `git diff --check` passed for changed application sources/tests.
- Data-preserving SDK install: `adb -s R54Y401XHZZ install --user 0 -r app/app/build/outputs/apk/debug/app-debug.apk`, exit 0, `Success`.
- Final APK SHA-256: `32dad10cc8c99fd5ab35ed1797867044a2d8be03aa5c3c807b3225aa9c4915cb`.
- SM-X620 screenshot-backed Android MCP interactions: dock launch → logo → all four Command Center sections; AutoBot → populated launchd console → system Back; Server → populated Local Server Console → system Back; Command Center → system Back → launcher.
- Existing launcher Server → Moonlight with saved Mac host → Back; ChatKJB → native Termux with populated existing Herdr workspace → keyboard hide/Back → launcher.
- Full-screen and approximately 432dp-wide Samsung popup window: logo entry and four sections verified. Popup AutoBot loads populated responsive cards and returns. Maximizing preserves Command Center. Final device state: full-screen Command Center.
- MCP crash query filtered to `com.termux`: count 0.
- Android MCP activity/element metadata is unreliable on this Samsung device; screenshots and observed native controls were used. No service start/stop/restart, prompt submission, or new push onboarding was performed.
- Independent combo/Reviewer: PASS_WITH_CONCERNS. Remaining concerns: untested automation delivery/populated/clear/tap paths and external policy evidence. Manifest keeps default standard launch mode, and automation PendingIntent uses CLEAR_TOP without SINGLE_TOP; no singleTop onNewIntent path was introduced.

## Tailscale access

The phone-only 8443 ACL blocked tablet consoles. The final effective access delta is only `galaxy-tab → macmini TCP 8443`, with an acceptance test for that edge. Hosts, SSH policies/tests, tag ownership, and auto-approvers are unchanged. Reloaded live policy matched the candidate semantically.

The visual grant editor initially converted the existing SMB rule to an equivalent grant without changing the intended 8443 rule. The JSON editor restored SMB to ACL form and applied the intended source addition. Effective-edge comparison confirmed no removed permissions and no other added permissions. See `evidence/command-center-policy-2026-09-05.txt`.

Build/install receipts are in `evidence/command-center-build-2026-09-05.log` and `evidence/command-center-install-2026-09-05.log`.

## Final width correction

The final user correction requires the tablet launcher's constrained width. All Command Center content now shares a centered `widthIn(max = 520.dp).fillMaxWidth()` column. Final APK build/test and installation passed. Physical final-device screenshots show launcher rows and Command Center buttons/cards at identical x=460..980 bounds in 1440px half-resolution landscape captures. The approximately 432dp popup window also passed system Back → launcher → logo → Command Center, then maximize; all four sections remained visible with no added Back control. Routing and notification implementations were unchanged after the earlier full navigation checks.
