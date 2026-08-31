package herdr

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalPaneListResult(t *testing.T) {
	raw := `{"id":"a","result":{"type":"pane_list","panes":[
	  {"pane_id":"w6:p1","workspace_id":"w6","tab_id":"w6:t1","focused":true,
	   "cwd":"/home/me/proj","agent":"claude","agent_status":"working","revision":0}]}}`
	var resp Response
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	var res paneListResult
	if err := json.Unmarshal(resp.Result, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Panes) != 1 {
		t.Fatalf("want 1 pane, got %d", len(res.Panes))
	}
	p := res.Panes[0]
	if p.PaneID != "w6:p1" || p.Agent != "claude" || p.AgentStatus != "working" || !p.Focused {
		t.Fatalf("bad pane: %+v", p)
	}
}

func TestPaneInfoParsesTerminalID(t *testing.T) {
	raw := `{"type":"pane_list","panes":[{"pane_id":"w7:p2","terminal_id":"term_abc","agent_status":"unknown"}]}`
	var res paneListResult
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Panes) != 1 || res.Panes[0].TerminalID != "term_abc" {
		t.Fatalf("terminal_id not parsed: %+v", res.Panes)
	}
}

func TestUnmarshalRPCError(t *testing.T) {
	var resp Response
	if err := json.Unmarshal([]byte(`{"id":"a","error":{"code":"not_found","message":"pane not found"}}`), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil || resp.Error.Code != "not_found" {
		t.Fatalf("want not_found error, got %+v", resp.Error)
	}
}

func TestStartAgentRequestEncodesTimeoutMS(t *testing.T) {
	req := StartAgentRequest{
		Name:      "q-servant",
		Kind:      AgentKindCodex,
		PaneID:    "w1F:p1",
		Args:      []string{"-m", "gpt-5.4"},
		TimeoutMS: TimeoutMillis(30000),
	}
	raw, err := json.Marshal(req.params())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["name"] != "q-servant" || decoded["kind"] != "codex" || decoded["pane_id"] != "w1F:p1" {
		t.Fatalf("required fields: %s", raw)
	}
	if _, ok := decoded["timeout"]; ok {
		t.Fatalf("must encode timeout_ms, not timeout: %s", raw)
	}
	if decoded["timeout_ms"] != float64(30000) {
		t.Fatalf("timeout_ms: %s", raw)
	}
	args, ok := decoded["args"].([]any)
	if !ok || len(args) != 2 || args[0] != "-m" || args[1] != "gpt-5.4" {
		t.Fatalf("args: %s", raw)
	}
}

func TestStartAgentRequestAlwaysEncodesArgsArray(t *testing.T) {
	raw, err := json.Marshal(StartAgentRequest{Name: "q-servant", Kind: "codex", PaneID: "w1F:p1"}.params())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	args, ok := decoded["args"].([]any)
	if !ok || args == nil || len(args) != 0 {
		t.Fatalf("nil args must encode as []: %s", raw)
	}
}

func TestUnmarshalAgentStartedResult(t *testing.T) {
	raw := `{"type":"agent_started","argv":["-m","gpt-5.4"],"agent":{"name":"q-servant","pane_id":"w1F:p1","terminal_id":"term_agent","workspace_id":"w1F","tab_id":"w1F:t1","agent":"codex","agent_status":"idle","focused":false,"interactive_ready":true}}`
	var res agentStartedResult
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		t.Fatal(err)
	}
	if res.Type != "agent_started" || res.Agent.Name != "q-servant" || res.Agent.PaneID != "w1F:p1" || res.Agent.AgentStatus != "idle" {
		t.Fatalf("bad agent_started: %+v", res)
	}
}

func TestPromptAgentRequestEncodesTargetTextWait(t *testing.T) {
	req := PromptAgentRequest{
		Target: "q-servant",
		Text:   "transcribed work",
		Wait:   &AgentWaitOptions{TimeoutMS: TimeoutMillis(120000), Until: []string{"idle", "done", "blocked"}},
	}
	raw, err := json.Marshal(req.params())
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["target"] != "q-servant" || decoded["text"] != "transcribed work" {
		t.Fatalf("prompt fields: %s", raw)
	}
	wait, ok := decoded["wait"].(map[string]any)
	if !ok {
		t.Fatalf("wait missing: %s", raw)
	}
	if wait["timeout_ms"] != float64(120000) {
		t.Fatalf("wait.timeout_ms: %s", raw)
	}
}

func TestSelectWorkspaceByLabelExact(t *testing.T) {
	ws := []WorkspaceInfo{
		{WorkspaceID: "w1", Label: "mobile"},
		{WorkspaceID: "w1F", Label: QServantWorkspaceLabel},
	}
	got, err := SelectWorkspaceByLabel(ws, QServantWorkspaceLabel)
	if err != nil {
		t.Fatal(err)
	}
	if got.WorkspaceID != "w1F" {
		t.Fatalf("got %+v", got)
	}
}

func TestSelectWorkspaceByLabelMissing(t *testing.T) {
	_, err := SelectWorkspaceByLabel([]WorkspaceInfo{{WorkspaceID: "w1", Label: "mobile"}}, QServantWorkspaceLabel)
	if !IsWorkspaceNotFound(err) {
		t.Fatalf("want not_found, got %v", err)
	}
	if err.Error() != "herdr space 'Q Servant' not found" {
		t.Fatalf("message: %v", err)
	}
}

func TestSelectWorkspaceByLabelAmbiguous(t *testing.T) {
	_, err := SelectWorkspaceByLabel([]WorkspaceInfo{
		{WorkspaceID: "w1F", Label: QServantWorkspaceLabel},
		{WorkspaceID: "w2", Label: QServantWorkspaceLabel},
	}, QServantWorkspaceLabel)
	if !IsWorkspaceAmbiguous(err) {
		t.Fatalf("want ambiguous, got %v", err)
	}
}

func TestSelectWorkspaceByLabelNoFuzzyOrCaseFold(t *testing.T) {
	_, err := SelectWorkspaceByLabel([]WorkspaceInfo{
		{WorkspaceID: "w1", Label: "q servant"},
		{WorkspaceID: "w2", Label: "Q  Servant"},
		{WorkspaceID: "w3", Label: "Q-Servant"},
	}, QServantWorkspaceLabel)
	if !IsWorkspaceNotFound(err) {
		t.Fatalf("want not_found for near-matches, got %v", err)
	}
}
