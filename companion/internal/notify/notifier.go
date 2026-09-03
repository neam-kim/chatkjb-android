package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/state"
)

type Push struct {
	Kind        string `json:"kind"`
	PaneID      string `json:"paneId"`
	WorkspaceID string `json:"workspaceId"`
	Title       string `json:"title"`
	Body        string `json:"body"`
}

type Notifier interface {
	Notify(ctx context.Context, p Push) error
}

// ShouldNotify encodes the two v1 triggers. displayName is the friendly pane
// name shown in the title (the project/cwd basename); it falls back to the
// workspace id when empty. body carries the pane-derived detail line: the last
// non-empty output line for a blocked pane, and a short completion summary for
// a finished pane, so the phone shows what actually finished.
func ShouldNotify(tr state.Transition, displayName, body string) (Push, bool) {
	name := displayName
	if name == "" {
		name = tr.WorkspaceID
	}
	switch {
	case tr.To == "blocked":
		return Push{Kind: "blocked", PaneID: tr.PaneID, WorkspaceID: tr.WorkspaceID,
			Title: name + " needs you", Body: body}, true
	case tr.From == "working" && (tr.To == "idle" || tr.To == "done"):
		return Push{Kind: "finished", PaneID: tr.PaneID, WorkspaceID: tr.WorkspaceID,
			Title: name + " finished", Body: body}, true
	case tr.To == "working":
		return Push{Kind: "clear", PaneID: tr.PaneID, WorkspaceID: tr.WorkspaceID}, true
	default:
		return Push{}, false
	}
}

type HTTPNotifier struct {
	endpoint string
	hc       *http.Client
}

func NewHTTPNotifier(endpoint string, hc *http.Client) *HTTPNotifier {
	if hc == nil {
		hc = http.DefaultClient
	}
	return &HTTPNotifier{endpoint: endpoint, hc: hc}
}

func (n *HTTPNotifier) Notify(ctx context.Context, p Push) error {
	if n.endpoint == "" {
		return nil // no endpoint registered yet
	}
	b, _ := json.Marshal(p)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("push endpoint returned %d", resp.StatusCode)
	}
	return nil
}

// summaryMaxLen bounds the finished-notification body. Push surfaces truncate
// long text anyway, and a short line keeps the notification scannable.
const summaryMaxLen = 140

// Summarize turns a pane's trailing terminal output into one short line
// describing what finished. Agent TUIs frame their transcript with bullet
// markers, box drawing, spinners and a trailing input prompt, so the last raw
// line is usually chrome. The terminal also hard-wraps prose across rows, so a
// single line is often a mid-sentence fragment. We therefore take the last
// contiguous block of real content and join it back into one sentence.
func Summarize(paneOutput string) string {
	lines := strings.Split(strings.TrimRight(paneOutput, "\n"), "\n")

	// Walk back to the end of the last content block, skipping trailing chrome.
	end := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if cleanSummaryLine(lines[i]) != "" {
			end = i
			break
		}
	}
	if end < 0 {
		return ""
	}
	// Extend upward while rows keep contributing content; a blank row or a
	// chrome row ends the paragraph.
	start := end
	for start > 0 {
		// A bullet marker starts a distinct message, so it bounds the block.
		if isBlockStart(lines[start]) {
			break
		}
		prev := lines[start-1]
		if strings.TrimSpace(prev) == "" || cleanSummaryLine(prev) == "" {
			break
		}
		start--
	}

	var parts []string
	for i := start; i <= end; i++ {
		if c := cleanSummaryLine(lines[i]); c != "" {
			parts = append(parts, c)
		}
	}
	return truncateSummary(joinWrapped(parts))
}

// isBlockStart reports whether a row begins a new transcript message rather
// than continuing a wrapped one.
func isBlockStart(raw string) bool {
	line := strings.TrimSpace(raw)
	for _, marker := range []string{"•", "·", "›", "»", "└", "├"} {
		if strings.HasPrefix(line, marker) {
			return true
		}
	}
	return false
}

// joinWrapped reassembles hard-wrapped rows. CJK text wraps without a space,
// so a separator is inserted only between two ASCII-word boundaries.
func joinWrapped(parts []string) string {
	var b strings.Builder
	for i, p := range parts {
		if i > 0 && needsSpaceJoin(parts[i-1], p) {
			b.WriteString(" ")
		}
		b.WriteString(p)
	}
	return strings.TrimSpace(b.String())
}

func needsSpaceJoin(prev, next string) bool {
	if prev == "" || next == "" {
		return false
	}
	last := []rune(prev)[len([]rune(prev))-1]
	first := []rune(next)[0]
	return last < 0x2E80 && first < 0x2E80
}

// cleanSummaryLine strips agent-TUI decoration from one line and returns ""
// when the line carries no human-readable content.
func cleanSummaryLine(raw string) string {
	line := strings.TrimSpace(raw)
	if line == "" {
		return ""
	}
	// Drop box drawing, rules, and transcript gutters that carry no content.
	if strings.Trim(line, "─━—-=_│┃|╭╮╯╰┌┐└┘├┤┬┴┼ ") == "" {
		return ""
	}
	// Drop decorated separators that embed a label, e.g.
	// "─ Worked for 14m 16s ────────────────". These mark the end of a turn
	// rather than describing the result.
	if strings.HasPrefix(line, "─") || strings.HasPrefix(line, "━") {
		return ""
	}
	// Strip leading transcript gutters and bullets ("• ", "│ ", "└ ", "› ").
	line = strings.TrimLeft(line, "•·›»▌▍│┃└├┌|> ")
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	for _, skip := range summarySkipPrefixes {
		if strings.HasPrefix(line, skip) {
			return ""
		}
	}
	if isAgentStatusLine(line) {
		return ""
	}
	if isShellPrompt(line) {
		return ""
	}
	// A line of pure punctuation or a lone spinner frame is not a summary.
	if strings.IndexFunc(line, isSummaryContent) < 0 {
		return ""
	}
	return line
}

// isAgentStatusLine matches the persistent Codex footer, e.g.
// "gpt-5.6-sol high · Full Access · never · Context 95% left · weekly 77% left"
// and the "Working (0s • esc to interrupt)" activity line. These are always
// present on a settled pane and would otherwise mask the real result.
func isAgentStatusLine(line string) bool {
	// The footer is rendered on one row and may be cut short by the terminal
	// width, so match on its stable leading segments rather than the tail.
	for _, marker := range agentStatusMarkers {
		if strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

// agentStatusMarkers appear only in Codex's persistent footer and activity
// line, never in assistant prose.
var agentStatusMarkers = []string{
	"Full Access",
	"% left",
	"esc to interrupt",
	"background terminal running",
	"to view transcript",
	"Worked for ",
}

// isShellPrompt matches an idle shell prompt line such as
// "neam@neamui-Macmini ~ %", which means the pane finished with no agent
// narration worth summarizing.
func isShellPrompt(line string) bool {
	if !strings.Contains(line, "@") {
		return false
	}
	switch line[len(line)-1] {
	case '%', '$', '#':
		return true
	}
	return false
}

// summarySkipPrefixes are agent-TUI affordances rather than results: the idle
// input prompt, the interrupt hint, and the raw command/output echo lines.
var summarySkipPrefixes = []string{
	"Ask Codex",
	"Ask ",
	"Working (",
	"esc to interrupt",
	"ctrl + t",
	"Ran ",
	"$ ",
}

func isSummaryContent(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r > 0x2FFF
}

func truncateSummary(line string) string {
	runes := []rune(line)
	if len(runes) <= summaryMaxLen {
		return line
	}
	return strings.TrimSpace(string(runes[:summaryMaxLen-1])) + "…"
}
