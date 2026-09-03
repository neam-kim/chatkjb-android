# ChatKJB

<p align="center">Homepage, Email/KJBMail, Finance와 Herdr agent control을 하나의 Android 앱에 내장한 개인 생산성 앱</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-AGPL--3.0-666666?labelColor=333333" alt="AGPL-3.0-or-later" /></a>
  <a href="https://github.com/neam-kim/ChatKJB/actions"><img src="https://img.shields.io/github/actions/workflow/status/neam-kim/ChatKJB/ci.yml?label=CI" alt="CI status" /></a>
</p>

ChatKJB는 `Kim JongBeom` launcher를 시작점으로 Homepage, KJBMail, Finance와 Herdr agent control을 한 프로세스에서 제공합니다. ChatKJB 버튼은 외부 앱이나 브라우저를 실행하지 않고, APK에 포함된 0cv Herdr Mobile Relay 프런트엔드를 앱 내부의 안전한 HTTPS WebView origin에서 엽니다.

## 제공 기능

- **Kim JongBeom launcher**: Homepage, Email, Finance, ChatKJB 진입점과 로고로 들어가는 별도 tailnet 전용 AutoBot/Server 설정 페이지, `kimjb://open/...` / 레거시 `kjbmail://open` deep link.
- **Homepage**: 앱 내부 화면에서 Homepage를 열고 launcher로 돌아옵니다.
- **Email / KJBMail**: `dev.herdr.kjbmail:mail-host`의 Thunderbird Mail host를 통합해 `MailEntryActivity`를 실행합니다. KJBMail 소스는 이 저장소에 포함되지 않습니다.
- **ChatKJB agent UI**: workspace/tab/agent 상태, 대화 기록, terminal, 승인·질문 응답, 파일 첨부와 relay 설정을 앱 안에서 제공합니다.
- **내장 웹 번들**: `app/app/src/main/assets/herdr/`에 검증된 Svelte 번들을 포함하며 `WebViewAssetLoader`의 HTTPS origin으로만 제공합니다.

## Architecture

```text
KJBMail composite build ------------------------> ChatKJB Android mail host
Herdr <---- Unix socket ---- 0cv relay (127.0.0.1:8375)
                              ^
                              | Tailscale Serve TLS (:8443, tailnet only)
                              |
                  encrypted WSS from embedded Svelte UI
                              ^
                              |
                     ChatKJB Android WebView
```

ChatKJB의 Android package/applicationId는 `com.neamkim.chatkjb`입니다. 제3자 Play 스토어 앱의 `dev.herdr.mobile`과 충돌하지 않지만, ChatKJB는 그 앱을 설치하거나 실행하지 않습니다.

## 보안 모델과 현재 제약

relay key와 메시지는 WebSocket 경로에서도 유지되는 0cv E2EE transport를 사용합니다. 번들은 `file://`이 아니라 `https://appassets.androidplatform.net`에서 제공되고, 파일·콘텐츠 직접 접근과 mixed content를 차단합니다. Content Security Policy는 relay 연결을 `wss://neam-macmini.taild81d38.ts.net:8443`으로 제한하며 relay key는 WebView origin의 local storage에만 남습니다. 외부 gateway, Cloudflare tunnel, WebRTC/STUN과 라우터 포트 매핑은 이 배치 경로에서 사용하지 않습니다.

현재 제약: KJBMail composite build가 없으면 Android 빌드는 의도적으로 명확한 설정 오류로 중단됩니다. WebView의 백그라운드 Web Push 지원은 일반 설치형 브라우저 PWA와 다를 수 있으므로, foreground agent control을 우선 보장합니다.

## 설치

### Android

요구사항: JDK 21, Android SDK (compileSdk 36). Relay는 [`0cv/herdr-mobile-relay`](https://github.com/0cv/herdr-mobile-relay)의 설치 지침을 따릅니다.

```bash
git clone --recurse-submodules https://github.com/neam-kim/ChatKJB.git
cd ChatKJB/app
ANDROID_HOME=$HOME/Android/Sdk ./gradlew :app:assembleDebug
# app/app/build/outputs/apk/debug/app-debug.apk

# 물리 기기에는 반드시 primary user 0 전용 설치 guard를 사용합니다.
# clone/work profile이 있으면 설치 전후 검증이 fail closed됩니다.
ANDROID_HOME=$HOME/Android/Sdk ../scripts/install-android-primary-user.sh \
  "$PWD/app/build/outputs/apk/debug/app-debug.apk" [adb-serial]
```

기존 clone은 `git submodule update --init --recursive`로 KJBMail을 가져오십시오. KJBMail은 [`neam-kim/KJBMail`](https://github.com/neam-kim/KJBMail)의 공개 source-only 저장소를 고정 커밋 `5082a97c66aa76d447ed8c5d4e5111db37cdb3ad`로 연결한 submodule이며, 기본 Gradle composite dependency로 사용됩니다. 다른 checkout을 사용하려면 `-Pkjbmail.dir=...` 또는 `KJBMAIL_DIR`로 명시할 수 있습니다.

## 개발 및 검증

```bash
cd app && ANDROID_HOME=$HOME/Android/Sdk ./gradlew :app:testDebugUnitTest :app:assembleDebug
```

내장 프런트엔드 소스는 `/Volumes/NEAM_SSD/herdr-mobile-relay`의 `chatkjb-embedded` 브랜치에서 관리하고, 검증된 `frontend/dist/`만 APK assets로 동기화합니다.

```bash
scripts/sync-embedded-herdr.sh /Volumes/NEAM_SSD/herdr-mobile-relay
```

동기화 스크립트는 upstream 기준 커밋을 확인하고 lint, type check, 287개 unit test, production build와 payload size gate를 모두 통과한 번들만 복사합니다. 기준 커밋과 라이선스는 [`THIRD-PARTY-NOTICES.md`](THIRD-PARTY-NOTICES.md)에 기록합니다.

이 Mac에서는 `scripts/run-embedded-herdr-relay.sh`의 설치 복사본을 로그인 사용자 Application Support의 LaunchAgent에서 실행합니다. 따라서 Herdr setup pane이나 별도 terminal pane을 계속 열어 둘 필요가 없습니다. relay는 loopback에만 바인딩되고 Tailscale Serve가 tailnet 전용 HTTPS/WSS `:8443`을 종단합니다. 기존 `:10110`과 `:8787` Serve 매핑은 별개로 유지합니다. macOS가 외장 볼륨의 로그인 서비스 실행을 제한하기 때문에 소스와 실행 복사본을 분리하며, relay token은 LaunchAgent plist가 아니라 Herdr plugin의 권한 제한 `relay.env`에서만 읽습니다.

`docs/`에는 기능 설계 자료가 있습니다. 이 저장소의 upstream 기반 코드와 bundled Termux/JetBrains Mono 고지는 [`THIRD-PARTY-NOTICES.md`](THIRD-PARTY-NOTICES.md)에 기록되어 있습니다. 원저작자 attribution과 upstream AGPLv3 조건을 변경하지 않습니다.

## License

ChatKJB는 **AGPL-3.0-or-later**로 배포됩니다. [`LICENSE`](LICENSE), [`THIRD-PARTY-NOTICES.md`](THIRD-PARTY-NOTICES.md)를 참조하십시오.
