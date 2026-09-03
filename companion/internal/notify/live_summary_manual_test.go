package notify

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestLiveSummaryDryRun is a manual probe: with HERDR_LIVE_PANES set to a
// comma-separated pane list it prints the summary that would be pushed for
// each real pane on this machine. Skipped in normal runs.
func TestLiveSummaryDryRun(t *testing.T) {
	panes := os.Getenv("HERDR_LIVE_PANES")
	if panes == "" {
		t.Skip("set HERDR_LIVE_PANES to run the live dry run")
	}
	for _, pane := range strings.Split(panes, ",") {
		out, err := exec.Command("/Users/neam/.local/bin/herdr",
			"pane", "read", pane, "--source", "recent-unwrapped", "--lines", "60").Output()
		if err != nil {
			t.Logf("%s: read error %v", pane, err)
			continue
		}
		t.Logf("%s => %q", pane, Summarize(string(out)))
	}
}
