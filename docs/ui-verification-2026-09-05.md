# Android UI verification — 2026-09-05

## Device and artifact

Verified on the S26 Ultra (SM-S948N), primary user 0, in portrait with the
existing dark system appearance, accounts and relay pairing. The tablet was not
modified. Installation used `scripts/install-android-primary-user.sh`; before
and after both installs, only SYSTEM user 0 and PRIVATE Secure Folder existed.
No CLONE or MANAGED profiles existed; the package was verified for user 0.

Initial refactor APK SHA-256:
`82353ab1fe01db599401008e61a5d8ce3dc25f008f3d6e9425cbc59662b61bbb`

Final APK including status-bar correction SHA-256:
`d1798f1426ec80e9ec1f7a8329915e0dd3a9fa5accc0f4af7c39b9d48186a7d3`

## Exercised flows

Actions were performed through Android-agent MCP and every transition was
observed with a fresh screenshot. Element inspection returned empty results;
OCR was unavailable because the MCP runtime lacked rapidocr. Screenshot-guided
pixel taps were therefore used. Private inbox, finance and gallery contents are
not retained in this report.

| Surface | Actions and observed result |
| --- | --- |
| Launcher | Opened by explicit home deep link; all five entries and logo visible and reachable. |
| Site | Opened homepage, followed Curriculum Vitae link, Back returned to homepage, another Back returned to native launcher. |
| Finance | Opened authenticated finance view with fresh data; existing login survived APK update; Back returned to launcher. |
| Email | Opened embedded inbox, focused search, typed an intentionally unmatched query, submitted and observed empty results; Back restored inbox, another Back restored launcher. |
| Moonlight Server | Opened embedded host list, opened Basic Settings, returned to host list, then launcher. Existing discovered Mac host visible. |
| Command Center | Logo opened settings; Skill Suggestions and Sentinel empty states rendered correctly. |
| AutoBot | Opened launchd console, observed live populated service counters/list, typed an unmatched filter, observed empty filtered results, cleared filter and observed list restoration; dismissed keyboard and returned to Command Center. |
| Management Server | Opened Local Server Console, observed live process data, selected Loopback filter and observed matching processes; Back returned to Command Center, then launcher. |
| ChatKJB | Existing encrypted relay connection showed 1/1 relays and live pane inventory. Opened an existing idle pane and saw its terminal/conversation. Typed `ui-check` into the reply field and cleared it without sending. |
| Attachments | Opened native photo picker from pane attachment button, cancelled, reopened and cancelled again. Returned to pane without sending or selecting a photo. |
| Chat back stack | Pane → inventory → native launcher, verified after each Back. |
| Final APK regression | Reopened launcher → Command Center → dark management console → Command Center → launcher → dark ChatKJB inventory → launcher; status-bar icon contrast remained appropriate on each surface. |

## Issue fixed

The native launcher and Command Center have fixed light backgrounds, but inherited
light status-bar icons from dark system appearance, making the clock and indicators
almost invisible. `LauncherStatusBar()` in `KimJbLauncher.kt` now sets dark icons
while either native light surface is active, and restores the previous appearance
on disposal. The correction was rebuilt, installed and visually verified through
light/dark screen transitions. It does not change WebView policy or navigation.

After the correction, `:app:testDebugUnitTest :app:assembleDebug` passed (17 tests,
zero failures/errors/skips; BUILD SUCCESSFUL). Android-agent crash/ANR query for
`com.neamkim.chatkjb` returned zero entries before and after the exercise.
The device was left on the corrected native launcher.

## Scope limits

This verifies the exercised portrait UI flows, not every product capability.
No email or agent messages were sent, files uploaded, live services stopped or
restarted, Moonlight pairing/stream started, or account credentials changed.
Fresh OAuth/second-factor login, disconnected-network variants, landscape/tablet
layouts, and actual attachment upload were not exercised. Existing authenticated
Finance and relay sessions did remain usable after installation.
