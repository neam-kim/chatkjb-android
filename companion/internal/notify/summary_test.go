package notify

import (
	"testing"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/state"
)

func transitionFinished() state.Transition {
	return state.Transition{PaneID: "w1C:p2E", WorkspaceID: "w1C", From: "working", To: "idle"}
}

// A finished Codex pane ends with rules, a spinner status line and the idle
// input prompt. The summary must skip that chrome and report the last real
// assistant line instead.
func TestSummarizeSkipsAgentChrome(t *testing.T) {
	pane := "• 중요한 문제를 발견했습니다.\n" +
		"• LaunchAgent plist를 템플릿으로 복구하고 재적재했습니다.\n" +
		"\n" +
		"──────────────────────────────\n" +
		"\n" +
		"› Ask Codex to do anything\n" +
		"\n"
	got := Summarize(pane)
	want := "LaunchAgent plist를 템플릿으로 복구하고 재적재했습니다."
	if got != want {
		t.Fatalf("summary = %q, want %q", got, want)
	}
}

func TestSummarizeSkipsCommandEchoAndWorkingLine(t *testing.T) {
	pane := "• 설정을 반영했습니다.\n" +
		"• Ran launchctl bootstrap gui/501 com.neam.security-sentinel.plist\n" +
		"  └ Bootstrap succeeded\n" +
		"• Working (0s • esc to interrupt)\n"
	// "Bootstrap succeeded" is real output, so it is a better summary than the
	// command echo above it.
	if got := Summarize(pane); got != "Bootstrap succeeded" {
		t.Fatalf("summary = %q", got)
	}
}

func TestSummarizeIgnoresDecorationOnlyPane(t *testing.T) {
	if got := Summarize("────────\n│  │\n\n   \n"); got != "" {
		t.Fatalf("decoration-only pane should summarize to empty, got %q", got)
	}
}

// Observed live on this machine: the Codex footer and activity line trail every
// settled pane and must never become the summary.
func TestSummarizeSkipsCodexStatusFooter(t *testing.T) {
	pane := "• S26U의 구형 dev.herdr.mobile만 제거했습니다.\n" +
		"\n" +
		"• Working (0s • esc to interrupt) · 1 background terminal running · /ps to view\n" +
		"\n" +
		"› Ask Codex to do anything\n" +
		"\n" +
		"  gpt-5.6-sol high · Full Access · never · Context 95% left\n"
	want := "S26U의 구형 dev.herdr.mobile만 제거했습니다."
	if got := Summarize(pane); got != want {
		t.Fatalf("summary = %q, want %q", got, want)
	}
}

func TestSummarizeSkipsWeeklyQuotaFooter(t *testing.T) {
	pane := "• 설정 반영을 마쳤습니다.\n" +
		"  anthropic/claude-opus-5 medium · Full Access · never · Context 100% left · weekly 77% left\n"
	if got := Summarize(pane); got != "설정 반영을 마쳤습니다." {
		t.Fatalf("summary = %q", got)
	}
}

// A pane that ended at a bare shell prompt has no result narration, so the
// body stays empty rather than showing the prompt itself.
func TestSummarizeSkipsIdleShellPrompt(t *testing.T) {
	if got := Summarize("neam@neamui-Macmini ~ %\n"); got != "" {
		t.Fatalf("shell prompt should not be a summary, got %q", got)
	}
}

func TestSummarizeTruncatesLongLine(t *testing.T) {
	long := ""
	for i := 0; i < 300; i++ {
		long += "a"
	}
	got := Summarize(long)
	if r := []rune(got); len(r) != summaryMaxLen || r[len(r)-1] != '…' {
		t.Fatalf("long summary should be truncated with an ellipsis, got %d runes", len(r))
	}
}

// The finished push must now carry the summary so the phone shows what
// completed rather than a bare "finished".
func TestFinishedPushCarriesSummary(t *testing.T) {
	p, ok := ShouldNotify(transitionFinished(), "omega3", "빌드와 테스트를 통과했습니다.")
	if !ok || p.Kind != "finished" {
		t.Fatalf("expected finished push, got %+v ok=%v", p, ok)
	}
	if p.Body != "빌드와 테스트를 통과했습니다." {
		t.Fatalf("finished body = %q, want the summary", p.Body)
	}
}
