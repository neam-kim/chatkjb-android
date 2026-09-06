#!/usr/bin/env bash
# Unit/mock tests for ChatKJB installation script policy matrix.
# Executes purely local mock adb/aapt environments without physical devices.
set -euo pipefail

TEST_DIR="$(mktemp -d /tmp/chatkjb-installer-test.XXXXXX)"
cleanup() {
  find "$TEST_DIR" -type f -delete 2>/dev/null || true
  find "$TEST_DIR" -depth -type d -delete 2>/dev/null || true
}
trap cleanup EXIT

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="$TEST_DIR/bin"
APKS_DIR="$TEST_DIR/apks"
STATE_DIR="$TEST_DIR/state"
INSTALL_LOG="$STATE_DIR/install_log.txt"
mkdir -p "$BIN_DIR" "$APKS_DIR" "$STATE_DIR"

LEGACY_APK="$APKS_DIR/app-legacyPhone-debug.apk"
UNIVERSAL_APK="$APKS_DIR/app-universal-debug.apk"
UNKNOWN_APK="$APKS_DIR/unsupported.apk"
touch "$LEGACY_APK" "$UNIVERSAL_APK" "$UNKNOWN_APK"

cat <<'EOF' > "$BIN_DIR/aapt"
#!/usr/bin/env bash
set -euo pipefail
apk="$3"
case "$apk" in
  *legacyPhone/debug/*|*-legacyPhone-debug.apk|*app-legacyPhone-v5.apk)
    printf "package: name='com.neamkim.chatkjb' versionCode='5' versionName='1.2.2'\n"
    ;;
  *debug/app-debug.apk|*stale-legacy-v4.apk)
    printf "package: name='com.neamkim.chatkjb' versionCode='4' versionName='1.2.1'\n"
    ;;
  *universal/debug/*|*-universal-debug.apk|*app-universal-v119.apk)
    printf "package: name='com.termux' versionCode='119' versionName='0.119.0-chatkjb'\n"
    ;;
  *stale-universal-v118.apk)
    printf "package: name='com.termux' versionCode='118' versionName='0.118.0'\n"
    ;;
  *legacyPhone*)
    printf "package: name='com.neamkim.chatkjb' versionCode='4' versionName='1.2.1'\n"
    ;;
  *universal*)
    printf "package: name='com.termux' versionCode='119' versionName='0.119.0-chatkjb'\n"
    ;;
  *)
    printf "package: name='com.other.app' versionCode='1' versionName='1.0'\n"
    ;;
esac
EOF
chmod +x "$BIN_DIR/aapt"

cat <<'EOF' > "$BIN_DIR/adb"
#!/usr/bin/env bash
set -euo pipefail
STATE_DIR="${MOCK_STATE_DIR:?set MOCK_STATE_DIR}"

case "${1:-}" in
  get-state)
    printf 'device\n'
    ;;
  shell)
    shift
    case "${1:-}" in
      cmd)
        cat "$STATE_DIR/user_list.txt"
        ;;
      pm)
        # pm list packages --user 0 <pkg>
        pkg="$5"
        if [[ "$pkg" == "0" && -n "${6:-}" ]]; then
          pkg="$6"
        fi
        if grep -Fqx "$pkg" "$STATE_DIR/installed_packages.txt" 2>/dev/null; then
          printf 'package:%s\n' "$pkg"
        fi
        ;;
      *)
        exit 1
        ;;
    esac
    ;;
install)
    shift
    apk=""
    while [[ $# -gt 0 ]]; do
      if [[ "$1" != -* && "$1" != "0" ]]; then
        apk="$1"
      fi
      shift
    done
    pkg="$(aapt dump badging "$apk" | sed -n "s/^package: name='\([^']*\)'.*/\1/p")"
    version="$(aapt dump badging "$apk" | sed -n "s/^package:.*versionCode='\([^']*\)'.*/\1/p")"
    log_path="${MOCK_INSTALL_LOG:?set MOCK_INSTALL_LOG}"
    printf '%s|%s|%s\n' "$apk" "$pkg" "$version" >> "$log_path"
    echo "$pkg" >> "$STATE_DIR/installed_packages.txt"
    printf 'Success\n'
    ;;
  *)
    echo "mock adb: unhandled command: $*" >&2
    exit 1
    ;;
esac
EOF
chmod +x "$BIN_DIR/adb"

export PATH="$BIN_DIR:$PATH"
export ADB_BIN="$BIN_DIR/adb"
export ANDROID_HOME="$TEST_DIR/sdk"
export MOCK_INSTALL_LOG="$INSTALL_LOG"
mkdir -p "$ANDROID_HOME/build-tools/37.0.0"
cp "$BIN_DIR/aapt" "$ANDROID_HOME/build-tools/37.0.0/aapt"

setup_mock() {
  local users_text="$1"
  local installed_text="$2"
  printf '%s\n' "$users_text" > "$STATE_DIR/user_list.txt"
  printf '%s\n' "$installed_text" > "$STATE_DIR/installed_packages.txt"
  : > "$INSTALL_LOG"
  export MOCK_STATE_DIR="$STATE_DIR"
}

clean_users="Users:
  UserInfo{0:Owner:13} type=full.SYSTEM
  UserInfo{150:Secure Folder:30} type=profile.PRIVATE"

clone_users="Users:
  UserInfo{0:Owner:13} type=full.SYSTEM
  UserInfo{99:Dual Messenger:30} type=profile.CLONE
  UserInfo{150:Secure Folder:30} type=profile.PRIVATE"

echo "Running installer policy test matrix..."

# Test 1: No installed identity -> auto-selects universal APK from directory
setup_mock "$clean_users" ""
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$APKS_DIR"
if ! grep -Fqx "com.termux" "$STATE_DIR/installed_packages.txt"; then
  echo "FAIL Test 1: expected com.termux installed" >&2
  exit 1
fi
echo "PASS Test 1: no installed identity -> universal auto-selected"

# Test 2: Existing legacyPhone -> auto-selects legacy APK from directory
setup_mock "$clean_users" "com.neamkim.chatkjb"
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$APKS_DIR"
if ! grep -Fqx "com.neamkim.chatkjb" "$STATE_DIR/installed_packages.txt"; then
  echo "FAIL Test 2: expected com.neamkim.chatkjb retained" >&2
  exit 1
fi
echo "PASS Test 2: existing legacyPhone -> legacy APK auto-selected"

# Test 3: Existing universal -> auto-selects universal APK from directory
setup_mock "$clean_users" "com.termux"
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$APKS_DIR"
if ! grep -Fqx "com.termux" "$STATE_DIR/installed_packages.txt"; then
  echo "FAIL Test 3: expected com.termux retained" >&2
  exit 1
fi
echo "PASS Test 3: existing universal -> universal APK auto-selected"

# Test 4: Both identities present -> strictly rejected
setup_mock "$clean_users" "com.neamkim.chatkjb
com.termux"
set +e
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$APKS_DIR" 2>/dev/null
exit_code=$?
set -e
if [[ $exit_code -ne 6 ]]; then
  echo "FAIL Test 4: expected exit code 6 for both identities installed, got $exit_code" >&2
  exit 1
fi
echo "PASS Test 4: both identities installed -> rejected with exit code 6"

# Test 5: Phone with clone/managed profile -> rejected before install
setup_mock "$clone_users" "com.neamkim.chatkjb"
set +e
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$APKS_DIR" 2>/dev/null
exit_code=$?
set -e
if [[ $exit_code -ne 3 ]]; then
  echo "FAIL Test 5: expected exit code 3 for clone profile on legacyPhone, got $exit_code" >&2
  exit 1
fi
echo "PASS Test 5: phone clone/managed profile -> rejected before install with exit code 3"

# Test 6: Tablet with clone profile -> universal user-0 script installs without modifying clone profile
setup_mock "$clone_users" ""
"$REPO_ROOT/scripts/install-android-universal-user0.sh" "$APKS_DIR"
if ! grep -Fqx "com.termux" "$STATE_DIR/installed_packages.txt"; then
  echo "FAIL Test 6: expected com.termux installed for user 0" >&2
  exit 1
fi
if ! grep -Eq 'type=profile\.CLONE' "$STATE_DIR/user_list.txt"; then
  echo "FAIL Test 6: clone profile was modified or removed" >&2
  exit 1
fi
echo "PASS Test 6: tablet with clone profile -> universal user-0 installed and clone profile preserved"

# Test 7: Unknown APK applicationId -> rejected
setup_mock "$clean_users" ""
set +e
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$UNKNOWN_APK" 2>/dev/null
exit_code=$?
set -e
if [[ $exit_code -ne 5 ]]; then
  echo "FAIL Test 7: expected exit code 5 for unknown APK, got $exit_code" >&2
  exit 1
fi
echo "PASS Test 7: unknown APK applicationId -> rejected with exit code 5"

# Test 8: Stale baseline APK (debug/app-debug.apk v4) alongside current variant (legacyPhone/debug v5)
# Verifies automatic selection picks the current variant, not the stale baseline.
STALE_TREE="$TEST_DIR/stale_tree"
mkdir -p "$STALE_TREE/debug" "$STALE_TREE/legacyPhone/debug" "$STALE_TREE/universal/debug"
touch "$STALE_TREE/debug/app-debug.apk"
touch "$STALE_TREE/legacyPhone/debug/app-legacyPhone-debug.apk"
touch "$STALE_TREE/universal/debug/app-universal-debug.apk"

setup_mock "$clean_users" "com.neamkim.chatkjb"
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$STALE_TREE"
if ! grep -Fqx "$STALE_TREE/legacyPhone/debug/app-legacyPhone-debug.apk|com.neamkim.chatkjb|5" "$INSTALL_LOG"; then
  echo "FAIL Test 8: current legacyPhone APK content was not selected" >&2
  cat "$INSTALL_LOG" >&2
  exit 1
fi
echo "PASS Test 8: stale baseline debug APK does not win over current legacyPhone path"

# Test 9: Generic directory with multiple versionCodes selects highest numeric versionCode.
MULTI_VER_DIR="$TEST_DIR/multi_ver"
mkdir -p "$MULTI_VER_DIR"
touch "$MULTI_VER_DIR/stale-legacy-v4.apk"
touch "$MULTI_VER_DIR/app-legacyPhone-v5.apk"

setup_mock "$clean_users" "com.neamkim.chatkjb"
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$MULTI_VER_DIR"
if ! grep -Fqx "$MULTI_VER_DIR/app-legacyPhone-v5.apk|com.neamkim.chatkjb|5" "$INSTALL_LOG"; then
  echo "FAIL Test 9: highest-version legacy APK content was not selected" >&2
  cat "$INSTALL_LOG" >&2
  exit 1
fi
echo "PASS Test 9: generic directory selects highest numeric versionCode"

# Test 10: Explicit APK file remains the caller's exact target.
setup_mock "$clean_users" "com.neamkim.chatkjb"
"$REPO_ROOT/scripts/install-android-primary-user.sh" "$MULTI_VER_DIR/stale-legacy-v4.apk"
if ! grep -Fqx "$MULTI_VER_DIR/stale-legacy-v4.apk|com.neamkim.chatkjb|4" "$INSTALL_LOG"; then
  echo "FAIL Test 10: explicit APK content was not preserved" >&2
  cat "$INSTALL_LOG" >&2
  exit 1
fi
echo "PASS Test 10: explicit APK file remains exact caller target"

# Test 11: Universal wrapper selects the highest universal version from a generic directory.
touch "$MULTI_VER_DIR/stale-universal-v118.apk"
touch "$MULTI_VER_DIR/app-universal-v119.apk"
setup_mock "$clean_users" ""
"$REPO_ROOT/scripts/install-android-universal-user0.sh" "$MULTI_VER_DIR"
if ! grep -Fqx "$MULTI_VER_DIR/app-universal-v119.apk|com.termux|119" "$INSTALL_LOG"; then
  echo "FAIL Test 11: highest-version universal APK content was not selected" >&2
  cat "$INSTALL_LOG" >&2
  exit 1
fi
echo "PASS Test 11: universal wrapper selects highest numeric versionCode"

echo "All 11 installer policy tests passed successfully."
