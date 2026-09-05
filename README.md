# ChatKJB Android Tab

Galaxy Tab용 비공개 포크입니다. Kim JongBeom launcher에서 Homepage, KJBMail, 그리고 로컬 Termux를 엽니다.

이 빌드는 기존 `com.termux` 패키지를 데이터 유지 업데이트로 교체합니다. 삼성 키보드는 그대로 두고, D2Coding을 기본 터미널 글꼴로, IME Composing Preview를 기본 활성화합니다.

## 구성

- `app/`: ChatKJB launcher + KJBMail host
- `vendor/termux-app/`: 한글 IME composing 지원 Termux
- `vendor/fonts/`: D2Coding Regular (OFL-1.1)
- `KJBMail/`: 메일 호스트 소스

## 빌드

JDK 21 이상과 Android SDK/NDK가 필요합니다.

```bash
cd app
ANDROID_HOME=$HOME/Library/Android/sdk ./gradlew :app:assembleDebug
```

APK는 `app/app/build/outputs/apk/debug/`에 생성됩니다. 설치 시 기존 Termux 서명(`testkey_untrusted.jks`)과 `applicationId=com.termux`, `versionCode=119`를 사용합니다.
