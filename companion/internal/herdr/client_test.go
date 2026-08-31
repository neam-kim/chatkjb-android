package herdr

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestClientListPanes(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetPanes([]PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "working"}})
	c := New(f.SocketPath())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	panes, err := c.ListPanes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(panes) != 1 || panes[0].PaneID != "w6:p1" {
		t.Fatalf("bad panes: %+v", panes)
	}
}

func TestClientSendTextReachesHerdr(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	if err := c.SendText(context.Background(), "w6:p1", "y"); err != nil {
		t.Fatal(err)
	}
	select {
	case params := <-f.lastSend:
		if params["pane_id"] != "w6:p1" || params["text"] != "y" {
			t.Fatalf("bad send params: %+v", params)
		}
	case <-time.After(time.Second):
		t.Fatal("send_text never reached herdr")
	}
}

func TestClientSendKeysSendsArray(t *testing.T) {
	// herdr requires `keys` to be a sequence (array); a bare string is rejected.
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	if err := c.SendKeys(context.Background(), "w6:p1", "enter"); err != nil {
		t.Fatal(err)
	}
	select {
	case params := <-f.lastSend:
		keys, ok := params["keys"].([]any)
		if !ok || len(keys) != 1 || keys[0] != "enter" {
			t.Fatalf("keys must be a 1-element array [\"enter\"], got %#v", params["keys"])
		}
	case <-time.After(time.Second):
		t.Fatal("send_keys never reached herdr")
	}
}

func TestClientCallPropagatesRPCError(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	_, err := c.Call(context.Background(), "bogus.method", nil)
	if err == nil {
		t.Fatal("expected error for unknown method")
	}
	var rpcErr *RPCError
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected *RPCError, got %T: %v", err, err)
	}
	if rpcErr.Code != "unknown_method" {
		t.Fatalf("want code unknown_method, got %q", rpcErr.Code)
	}
}

func TestListWorkspacesAndTabs(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetWorkspaces([]WorkspaceInfo{
		{WorkspaceID: "w3", Label: "apollo", Number: 1, AgentStatus: "idle", PaneCount: 1, TabCount: 1},
		{WorkspaceID: "w5", Label: "wt-cost-dashboards", Number: 2, Focused: true, PaneCount: 1, TabCount: 1,
			Worktree: &WorktreeInfo{RepoName: "ops", IsLinkedWorktree: true}},
	})
	f.SetTabs([]TabInfo{
		{TabID: "w7:t1", Label: "1", Number: 1, WorkspaceID: "w7", AgentStatus: "idle", PaneCount: 1},
		{TabID: "w7:t2", Label: "2", Number: 2, WorkspaceID: "w7", AgentStatus: "unknown", PaneCount: 1},
	})
	c := New(f.SocketPath())

	ws, err := c.ListWorkspaces(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ws) != 2 || ws[0].Label != "apollo" || ws[1].Worktree == nil || ws[1].Worktree.RepoName != "ops" {
		t.Fatalf("bad workspaces: %+v", ws)
	}

	tabs, err := c.ListTabs(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tabs) != 2 || tabs[1].Label != "2" || tabs[1].WorkspaceID != "w7" {
		t.Fatalf("bad tabs: %+v", tabs)
	}
}

func TestClientRenameAndCloseReachHerdr(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	ctx := context.Background()

	cases := []struct {
		call   func() error
		method string
		key    string // param key carrying the id
		id     string
		label  string // "" for close
	}{
		{func() error { return c.RenameWorkspace(ctx, "w7", "omega3") }, "workspace.rename", "workspace_id", "w7", "omega3"},
		{func() error { return c.RenameTab(ctx, "w7:t1", "build") }, "tab.rename", "tab_id", "w7:t1", "build"},
		{func() error { return c.RenamePane(ctx, "w7:p2", "logs") }, "pane.rename", "pane_id", "w7:p2", "logs"},
		{func() error { return c.CloseWorkspace(ctx, "w7") }, "workspace.close", "workspace_id", "w7", ""},
		{func() error { return c.CloseTab(ctx, "w7:t1") }, "tab.close", "tab_id", "w7:t1", ""},
		{func() error { return c.ClosePane(ctx, "w7:p2") }, "pane.close", "pane_id", "w7:p2", ""},
	}
	for _, tc := range cases {
		if err := tc.call(); err != nil {
			t.Fatalf("%s: %v", tc.method, err)
		}
		select {
		case rec := <-f.lastCall:
			if rec.Method != tc.method {
				t.Fatalf("want method %s, got %s", tc.method, rec.Method)
			}
			if rec.Params[tc.key] != tc.id {
				t.Fatalf("%s: want %s=%s, got %v", tc.method, tc.key, tc.id, rec.Params[tc.key])
			}
			if tc.label != "" && rec.Params["label"] != tc.label {
				t.Fatalf("%s: want label=%s, got %v", tc.method, tc.label, rec.Params["label"])
			}
		case <-time.After(time.Second):
			t.Fatalf("%s never reached herdr", tc.method)
		}
	}
}

func TestClientListWorktrees(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetWorktrees([]WorktreeEntry{
		{Path: "/repo", IsLinkedWorktree: false, OpenWorkspaceID: "w1", Label: "app"},
		{Path: "/repo-wt", Branch: "feat/x", IsLinkedWorktree: true, OpenWorkspaceID: "w2", Label: "app"},
	})
	c := New(f.SocketPath())
	wts, err := c.ListWorktrees(context.Background(), "w1")
	if err != nil {
		t.Fatalf("ListWorktrees: %v", err)
	}
	if len(wts) != 2 || wts[0].OpenWorkspaceID != "w1" || wts[1].Branch != "feat/x" || !wts[1].IsLinkedWorktree {
		t.Fatalf("unexpected entries: %+v", wts)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Method != "worktree.list" || rec.Params["workspace_id"] != "w1" {
			t.Fatalf("bad worktree.list params: %+v", rec)
		}
	case <-time.After(time.Second):
		t.Fatal("worktree.list not recorded")
	}
}

func TestClientCreateMoveAndAgents(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	ctx := context.Background()

	// create methods return the new pane's ids parsed from each result envelope
	pid, tid, err := c.CreateWorkspace(ctx)
	if err != nil || pid != "wZ:p1" || tid != "term_ws" {
		t.Fatalf("CreateWorkspace: %q %q %v", pid, tid, err)
	}
	pid, tid, err = c.CreateTab(ctx, "w7")
	if err != nil || pid != "w7:pT" || tid != "term_tab" {
		t.Fatalf("CreateTab: %q %q %v", pid, tid, err)
	}
	pid, tid, err = c.SplitPane(ctx, "w7:p2", "", "down")
	if err != nil || pid != "w7:pS" || tid != "term_split" {
		t.Fatalf("SplitPane: %q %q %v", pid, tid, err)
	}
	pid, tid, err = c.StartAgent(ctx, "claude", []string{"claude"}, "w7", "", "down")
	if err != nil || pid != "w7:pA" || tid != "term_agent" {
		t.Fatalf("StartAgent: %q %q %v", pid, tid, err)
	}
	if err := c.MovePane(ctx, "w7:p2", "new_tab", "", ""); err != nil {
		t.Fatalf("MovePane: %v", err)
	}
	names, err := c.ListAgentNames(ctx)
	if err != nil || len(names) != 2 || names[0] != "claude" || names[1] != "codex" {
		t.Fatalf("ListAgentNames: %v %v", names, err)
	}

	// verify the params the split/agent/move calls sent
	got := map[string]map[string]any{}
	for i := 0; i < 3; i++ {
		select {
		case rec := <-f.lastCall:
			got[rec.Method] = rec.Params
		case <-time.After(time.Second):
			t.Fatal("missing recorded call")
		}
	}
	if got["pane.split"]["direction"] != "down" || got["pane.split"]["target_pane_id"] != "w7:p2" {
		t.Fatalf("pane.split params: %v", got["pane.split"])
	}
	if got["agent.start"]["name"] != "claude" || got["agent.start"]["split"] != "down" {
		t.Fatalf("agent.start params: %v", got["agent.start"])
	}
	dest, _ := got["pane.move"]["destination"].(map[string]any)
	if dest["type"] != "new_tab" {
		t.Fatalf("pane.move destination: %v", got["pane.move"])
	}
}

func TestStartAgentOnPaneSendsProtocol20Shape(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	agent, err := c.StartAgentOnPane(context.Background(), StartAgentRequest{
		Name:      "q-servant",
		Kind:      AgentKindCodex,
		PaneID:    "w1F:p2",
		Args:      []string{"-m", "gpt-5.4", "-c", `model_reasoning_effort="high"`},
		TimeoutMS: TimeoutMillis(30000),
	})
	if err != nil {
		t.Fatal(err)
	}
	if agent.PaneID != "w1F:p2" || agent.TerminalID != "term_agent" || agent.Name != "q-servant" {
		t.Fatalf("agent: %+v", agent)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Method != "agent.start" {
			t.Fatalf("method: %s", rec.Method)
		}
		if rec.Params["name"] != "q-servant" || rec.Params["kind"] != "codex" || rec.Params["pane_id"] != "w1F:p2" {
			t.Fatalf("required params: %+v", rec.Params)
		}
		if _, ok := rec.Params["timeout"]; ok {
			t.Fatalf("must not send timeout: %+v", rec.Params)
		}
		if rec.Params["timeout_ms"] != float64(30000) {
			t.Fatalf("timeout_ms: %+v", rec.Params)
		}
		if rec.Params["argv"] != nil || rec.Params["focus"] != nil || rec.Params["workspace_id"] != nil || rec.Params["split"] != nil {
			t.Fatalf("stale fields present: %+v", rec.Params)
		}
		args, ok := rec.Params["args"].([]any)
		if !ok || len(args) != 4 || args[0] != "-m" {
			t.Fatalf("args: %+v", rec.Params["args"])
		}
	case <-time.After(time.Second):
		t.Fatal("agent.start never reached herdr")
	}
}

func TestStartAgentOnPaneDefaultsKindCodex(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	if _, err := c.StartAgentOnPane(context.Background(), StartAgentRequest{Name: "q-servant", PaneID: "w1F:p1"}); err != nil {
		t.Fatal(err)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Params["kind"] != "codex" {
			t.Fatalf("kind: %+v", rec.Params)
		}
		args, ok := rec.Params["args"].([]any)
		if !ok || args == nil {
			t.Fatalf("args must be an array: %+v", rec.Params["args"])
		}
	case <-time.After(time.Second):
		t.Fatal("agent.start never reached herdr")
	}
}

func TestFindQServantWorkspaceExact(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetWorkspaces([]WorkspaceInfo{
		{WorkspaceID: "w1", Label: "mobile", Number: 1},
		{WorkspaceID: "w1F", Label: "Q Servant", Number: 2, ActiveTabID: "w1F:t1"},
	})
	c := New(f.SocketPath())
	ws, err := c.FindQServantWorkspace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ws.WorkspaceID != "w1F" || ws.Label != "Q Servant" {
		t.Fatalf("got %+v", ws)
	}
}

func TestFindQServantWorkspaceMissing(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetWorkspaces([]WorkspaceInfo{{WorkspaceID: "w1", Label: "mobile"}})
	c := New(f.SocketPath())
	_, err := c.FindQServantWorkspace(context.Background())
	if !IsWorkspaceNotFound(err) {
		t.Fatalf("want not_found, got %v", err)
	}
}

func TestFindQServantWorkspaceAmbiguous(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetWorkspaces([]WorkspaceInfo{
		{WorkspaceID: "w1F", Label: "Q Servant"},
		{WorkspaceID: "w2", Label: "Q Servant"},
	})
	c := New(f.SocketPath())
	_, err := c.FindQServantWorkspace(context.Background())
	if !IsWorkspaceAmbiguous(err) {
		t.Fatalf("want ambiguous, got %v", err)
	}
}

func TestPromptAgentSendsTranscriptAsData(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	agent, err := c.PromptAgent(context.Background(), PromptAgentRequest{
		Target: "q-servant",
		Text:   "do the work",
		Wait:   &AgentWaitOptions{TimeoutMS: TimeoutMillis(120000)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if agent.AgentStatus != "working" {
		t.Fatalf("agent: %+v", agent)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Method != "agent.prompt" {
			t.Fatalf("method: %s", rec.Method)
		}
		if rec.Params["target"] != "q-servant" || rec.Params["text"] != "do the work" {
			t.Fatalf("params: %+v", rec.Params)
		}
		wait, ok := rec.Params["wait"].(map[string]any)
		if !ok || wait["timeout_ms"] != float64(120000) {
			t.Fatalf("wait: %+v", rec.Params["wait"])
		}
	case <-time.After(time.Second):
		t.Fatal("agent.prompt never reached herdr")
	}
}

func TestWaitAgentUsesProtocol20WaitShape(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	agent, err := c.WaitAgent(context.Background(), "q-servant-new", AgentWaitOptions{
		TimeoutMS: TimeoutMillis(60000),
		Until:     []string{"idle", "done"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if agent.Name != "q-servant-new" || agent.AgentStatus != "idle" || !agent.InteractiveReady {
		t.Fatalf("agent: %+v", agent)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Method != "agent.wait" || rec.Params["target"] != "q-servant-new" || rec.Params["timeout_ms"] != float64(60000) {
			t.Fatalf("agent.wait params: %+v", rec)
		}
		until, ok := rec.Params["until"].([]any)
		if !ok || len(until) != 2 || until[0] != "idle" || until[1] != "done" {
			t.Fatalf("agent.wait until: %#v", rec.Params["until"])
		}
	case <-time.After(time.Second):
		t.Fatal("agent.wait never reached herdr")
	}
}

func TestInterruptAgentSendsCtrlC(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	if err := c.InterruptAgent(context.Background(), "q-servant"); err != nil {
		t.Fatal(err)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Method != "agent.send_keys" {
			t.Fatalf("method: %s", rec.Method)
		}
		if rec.Params["target"] != "q-servant" {
			t.Fatalf("target: %+v", rec.Params)
		}
		keys, ok := rec.Params["keys"].([]any)
		if !ok || len(keys) != 1 || keys[0] != "ctrl+c" {
			t.Fatalf("keys must be [ctrl+c], got %#v", rec.Params["keys"])
		}
	case <-time.After(time.Second):
		t.Fatal("agent.send_keys never reached herdr")
	}
}

func TestGetListAndReadAgent(t *testing.T) {
	f := newFakeHerdr(t)
	f.readText["q-servant"] = "report ready"
	c := New(f.SocketPath())
	ctx := context.Background()
	info, err := c.GetAgent(ctx, "q-servant")
	if err != nil || info.Name != "q-servant" || info.AgentStatus != "idle" {
		t.Fatalf("GetAgent: %+v %v", info, err)
	}
	agents, err := c.ListAgents(ctx)
	if err != nil || len(agents) != 1 || agents[0].Name != "q-servant" {
		t.Fatalf("ListAgents: %+v %v", agents, err)
	}
	txt, err := c.ReadAgent(ctx, "q-servant", "recent_unwrapped", 80)
	if err != nil || txt != "report ready" {
		t.Fatalf("ReadAgent: %q %v", txt, err)
	}
}

func TestListPanesInWorkspaceFilters(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetPanes([]PaneInfo{
		{PaneID: "w1F:p1", WorkspaceID: "w1F"},
		{PaneID: "w1:p1", WorkspaceID: "w1"},
	})
	c := New(f.SocketPath())
	panes, err := c.ListPanesInWorkspace(context.Background(), "w1F")
	if err != nil {
		t.Fatal(err)
	}
	if len(panes) != 1 || panes[0].PaneID != "w1F:p1" {
		t.Fatalf("panes: %+v", panes)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Method != "pane.list" || rec.Params["workspace_id"] != "w1F" {
			t.Fatalf("pane.list params: %+v", rec)
		}
	case <-time.After(time.Second):
		t.Fatal("pane.list not recorded")
	}
}

func TestSplitPaneSendsDirection(t *testing.T) {
	f := newFakeHerdr(t)
	c := New(f.SocketPath())
	pid, tid, err := c.SplitPane(context.Background(), "w1F:p1", "w1F", "right")
	if err != nil || pid != "w7:pS" || tid != "term_split" {
		t.Fatalf("SplitPane: %q %q %v", pid, tid, err)
	}
	select {
	case rec := <-f.lastCall:
		if rec.Method != "pane.split" {
			t.Fatalf("method: %s", rec.Method)
		}
		if rec.Params["direction"] != "right" || rec.Params["target_pane_id"] != "w1F:p1" || rec.Params["workspace_id"] != "w1F" {
			t.Fatalf("params: %+v", rec.Params)
		}
	case <-time.After(time.Second):
		t.Fatal("pane.split not recorded")
	}
}
