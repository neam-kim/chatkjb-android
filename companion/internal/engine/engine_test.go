package engine

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/herdr"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/qservant"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/state"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/wsserver"
)

func dialWS(t *testing.T, srv *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	c, _, err := websocket.Dial(context.Background(), url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func readFrame(t *testing.T, c *websocket.Conn) map[string]any {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, b, err := c.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	json.Unmarshal(b, &m)
	return m
}

func TestServerSendsWelcomeThenSnapshotThenUpdate(t *testing.T) {
	store := state.NewStore()
	store.Apply([]herdr.PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "working"}})

	srvObj := wsserver.NewServer(wsserver.AllowAll{}, fakeRPC{})
	srvObj.SetInitialSnapshot(store.Snapshot)
	httpSrv := httptest.NewServer(srvObj.Handler())
	defer httpSrv.Close()

	c := dialWS(t, httpSrv)
	defer c.Close(websocket.StatusNormalClosure, "")

	if f := readFrame(t, c); f["t"] != "welcome" {
		t.Fatalf("want welcome, got %v", f["t"])
	}
	if f := readFrame(t, c); f["t"] != "panes" {
		t.Fatalf("want panes, got %v", f["t"])
	}
	if f := readFrame(t, c); f["t"] != "workspaces" {
		t.Fatalf("want workspaces, got %v", f["t"])
	}
	if f := readFrame(t, c); f["t"] != "tabs" {
		t.Fatalf("want tabs, got %v", f["t"])
	}

	srvObj.Broadcast([]byte(`{"t":"pane_update","pane":{"paneId":"w6:p1","agentStatus":"blocked"}}`))
	f := readFrame(t, c)
	if f["t"] != "pane_update" {
		t.Fatalf("want pane_update, got %v", f["t"])
	}
}

type fakeRPC struct{}

func (fakeRPC) ReadPane(context.Context, string, string, int) (string, error) { return "", nil }
func (fakeRPC) SendText(context.Context, string, string) error                { return nil }
func (fakeRPC) SendKeys(context.Context, string, string) error                { return nil }
func (fakeRPC) RenameWorkspace(context.Context, string, string) error         { return nil }
func (fakeRPC) RenameTab(context.Context, string, string) error               { return nil }
func (fakeRPC) RenamePane(context.Context, string, string) error              { return nil }
func (fakeRPC) CloseWorkspace(context.Context, string) error                  { return nil }
func (fakeRPC) CloseTab(context.Context, string) error                        { return nil }
func (fakeRPC) ClosePane(context.Context, string) error                       { return nil }

func (fakeRPC) CreateWorkspace(context.Context) (string, string, error)   { return "", "", nil }
func (fakeRPC) CreateTab(context.Context, string) (string, string, error) { return "", "", nil }
func (fakeRPC) SplitPane(context.Context, string, string, string) (string, string, error) {
	return "", "", nil
}
func (fakeRPC) StartAgent(context.Context, string, []string, string, string, string) (string, string, error) {
	return "", "", nil
}
func (fakeRPC) MovePane(context.Context, string, string, string, string) error { return nil }
func (fakeRPC) ListAgentNames(context.Context) ([]string, error)               { return nil, nil }
func (fakeRPC) ListWorktrees(context.Context, string) ([]herdr.WorktreeEntry, error) {
	return nil, nil
}

type fakeHerdr struct {
	ln         net.Listener
	path       string
	mu         sync.Mutex
	panes      []herdr.PaneInfo
	workspaces []herdr.WorkspaceInfo
	tabs       []herdr.TabInfo
	subs       []chan map[string]any
}

func newFakeHerdr(t *testing.T) *fakeHerdr {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "herdr-eng-*")
	if err != nil {
		dir = t.TempDir()
	}
	path := filepath.Join(dir, "herdr.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	f := &fakeHerdr{ln: ln, path: path}
	go f.serve()
	t.Cleanup(func() {
		ln.Close()
		f.mu.Lock()
		for _, ch := range f.subs {
			close(ch)
		}
		f.subs = nil
		f.mu.Unlock()
	})
	return f
}

func (f *fakeHerdr) SocketPath() string { return f.path }

func (f *fakeHerdr) SetPanes(p []herdr.PaneInfo) { f.mu.Lock(); f.panes = p; f.mu.Unlock() }

func (f *fakeHerdr) SetWorkspaces(w []herdr.WorkspaceInfo) {
	f.mu.Lock()
	f.workspaces = w
	f.mu.Unlock()
}

func (f *fakeHerdr) SetTabs(t []herdr.TabInfo) { f.mu.Lock(); f.tabs = t; f.mu.Unlock() }

func (f *fakeHerdr) serve() {
	for {
		c, err := f.ln.Accept()
		if err != nil {
			return
		}
		go f.handle(c)
	}
}

func (f *fakeHerdr) handle(c net.Conn) {
	defer c.Close()
	line, err := bufio.NewReader(c).ReadBytes('\n')
	if err != nil {
		return
	}
	var req struct {
		ID     string         `json:"id"`
		Method string         `json:"method"`
		Params map[string]any `json:"params"`
	}
	if json.Unmarshal(line, &req) != nil {
		return
	}
	enc := json.NewEncoder(c)
	switch req.Method {
	case "ping":
		enc.Encode(map[string]any{"id": req.ID, "result": map[string]any{"type": "pong", "version": "0.7.1", "protocol": 14}})
	case "pane.list":
		f.mu.Lock()
		panes := f.panes
		f.mu.Unlock()
		enc.Encode(map[string]any{"id": req.ID, "result": map[string]any{"type": "pane_list", "panes": panes}})
	case "pane.read":
		enc.Encode(map[string]any{"id": req.ID, "result": map[string]any{"type": "pane_read", "read": map[string]any{"pane_id": req.Params["pane_id"], "source": req.Params["source"], "text": ""}}})
	case "workspace.list":
		f.mu.Lock()
		workspaces := f.workspaces
		f.mu.Unlock()
		enc.Encode(map[string]any{"id": req.ID, "result": map[string]any{"type": "workspace_list", "workspaces": workspaces}})
	case "tab.list":
		f.mu.Lock()
		tabs := f.tabs
		f.mu.Unlock()
		enc.Encode(map[string]any{"id": req.ID, "result": map[string]any{"type": "tab_list", "tabs": tabs}})
	case "events.subscribe":
		ch := make(chan map[string]any, 16)
		f.mu.Lock()
		f.subs = append(f.subs, ch)
		f.mu.Unlock()
		enc.Encode(map[string]any{"id": req.ID, "result": map[string]any{"type": "subscription_started"}})
		for ev := range ch {
			if err := enc.Encode(ev); err != nil {
				return
			}
		}
	default:
		enc.Encode(map[string]any{"id": req.ID, "error": map[string]any{"code": "unknown_method", "message": req.Method}})
	}
}

func (f *fakeHerdr) PushEvent(paneID, status string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, ch := range f.subs {
		select {
		case ch <- map[string]any{"type": "pane.agent_status_changed", "pane_id": paneID, "agent_status": status}:
		default:
		}
	}
}

func TestEngineFiresBlockedPushToRegisteredEndpoint(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetPanes([]herdr.PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "working"}})

	gotPush := make(chan map[string]any, 1)
	pushSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]any
		json.NewDecoder(r.Body).Decode(&m)
		gotPush <- m
	}))
	defer pushSrv.Close()

	e := New(Config{SocketPath: f.SocketPath(), ListenAddr: "127.0.0.1:0", PollInterval: 100 * time.Millisecond})
	e.setEndpoint(pushSrv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go e.pollLoop(ctx)

	time.Sleep(200 * time.Millisecond)
	f.SetPanes([]herdr.PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "blocked"}})

	select {
	case m := <-gotPush:
		if m["kind"] != "blocked" || m["workspaceId"] != "w6" {
			t.Fatalf("bad push: %v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no push fired")
	}
}

func TestResumeFiresClearPush(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetPanes([]herdr.PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "blocked"}})

	gotPush := make(chan map[string]any, 4)
	pushSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]any
		json.NewDecoder(r.Body).Decode(&m)
		gotPush <- m
	}))
	defer pushSrv.Close()

	e := New(Config{SocketPath: f.SocketPath(), ListenAddr: "127.0.0.1:0", PollInterval: 100 * time.Millisecond})
	e.setEndpoint(pushSrv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go e.pollLoop(ctx)

	time.Sleep(200 * time.Millisecond)
	f.SetPanes([]herdr.PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "working"}})

	deadline := time.After(2 * time.Second)
	for {
		select {
		case m := <-gotPush:
			if m["kind"] == "clear" {
				if m["paneId"] != "w6:p1" || m["workspaceId"] != "w6" {
					t.Fatalf("bad clear push: %v", m)
				}
				return
			}
		case <-deadline:
			t.Fatal("no clear push fired on resume to working")
		}
	}
}

func TestEngineSubscriptionTriggersFastPoll(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetPanes([]herdr.PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "working"}})

	gotPush := make(chan map[string]any, 1)
	pushSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var m map[string]any
		json.NewDecoder(r.Body).Decode(&m)
		select {
		case gotPush <- m:
		default:
		}
	}))
	defer pushSrv.Close()

	e := New(Config{SocketPath: f.SocketPath(), ListenAddr: "127.0.0.1:0", PollInterval: 5 * time.Second})
	e.setEndpoint(pushSrv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go e.pollLoop(ctx)

	time.Sleep(300 * time.Millisecond)

	f.SetPanes([]herdr.PaneInfo{{PaneID: "w6:p1", WorkspaceID: "w6", Agent: "claude", AgentStatus: "blocked"}})
	f.PushEvent("w6:p1", "blocked")

	select {
	case m := <-gotPush:
		if m["kind"] != "blocked" {
			t.Fatalf("want blocked push, got %v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("subscription did not trigger a fast poll/push within 2s (ticker is 5s)")
	}
}

func TestPollOncePopulatesWorkspacesAndTabs(t *testing.T) {
	f := newFakeHerdr(t)
	f.SetWorkspaces([]herdr.WorkspaceInfo{
		{WorkspaceID: "w7", Label: "omega3", Number: 4, PaneCount: 2, TabCount: 2},
	})
	f.SetTabs([]herdr.TabInfo{
		{TabID: "w7:t1", Label: "1", Number: 1, WorkspaceID: "w7"},
	})

	e := New(Config{SocketPath: f.SocketPath(), ListenAddr: "127.0.0.1:0"})
	e.pollOnce(context.Background())

	ws := e.store.Workspaces()
	if len(ws) != 1 || ws[0].Label != "omega3" {
		t.Fatalf("workspaces not populated: %+v", ws)
	}
	tabs := e.store.Tabs()
	if len(tabs) != 1 || tabs[0].TabID != "w7:t1" {
		t.Fatalf("tabs not populated: %+v", tabs)
	}
}

type stubHerdrRPCClient struct {
	mu               sync.Mutex
	calls            []string
	workspaces       []herdr.WorkspaceInfo
	panes            []herdr.PaneInfo
	splitPaneID      string
	agentStatus      string
	agentInteractive bool
	onPrompt         func(req herdr.PromptAgentRequest)
	lastStartReq     herdr.StartAgentRequest
	lastPromptReq    herdr.PromptAgentRequest
}

func (s *stubHerdrRPCClient) Call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, method)
	return nil, nil
}
func (s *stubHerdrRPCClient) ListPanes(ctx context.Context) ([]herdr.PaneInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "pane.list")
	return s.panes, nil
}
func (s *stubHerdrRPCClient) ListPanesInWorkspace(ctx context.Context, wsID string) ([]herdr.PaneInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "pane.list:"+wsID)
	var out []herdr.PaneInfo
	for _, p := range s.panes {
		if p.WorkspaceID == wsID {
			out = append(out, p)
		}
	}
	return out, nil
}
func (s *stubHerdrRPCClient) ListWorkspaces(ctx context.Context) ([]herdr.WorkspaceInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "workspace.list")
	return s.workspaces, nil
}
func (s *stubHerdrRPCClient) FindQServantWorkspace(ctx context.Context) (herdr.WorkspaceInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "workspace.find:Q Servant")
	return herdr.SelectWorkspaceByLabel(s.workspaces, herdr.QServantWorkspaceLabel)
}
func (s *stubHerdrRPCClient) SplitPane(ctx context.Context, target, ws, dir string) (string, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "pane.split:"+target)
	pid := s.splitPaneID
	if pid == "" {
		pid = "w1:pS"
	}
	return pid, "term_split", nil
}
func (s *stubHerdrRPCClient) StartAgentOnPane(ctx context.Context, req herdr.StartAgentRequest) (herdr.AgentInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "agent.start:"+req.Name)
	s.lastStartReq = req
	return herdr.AgentInfo{Name: req.Name, PaneID: req.PaneID, Agent: req.Kind, AgentStatus: "idle"}, nil
}
func (s *stubHerdrRPCClient) GetAgent(ctx context.Context, target string) (herdr.AgentInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "agent.get:"+target)
	st := s.agentStatus
	if st == "" {
		st = "idle"
	}
	return herdr.AgentInfo{Name: target, Agent: "codex", AgentStatus: st, InteractiveReady: s.agentInteractive}, nil
}
func (s *stubHerdrRPCClient) WaitAgent(ctx context.Context, target string, wait herdr.AgentWaitOptions) (herdr.AgentInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "agent.wait:"+target)
	st := s.agentStatus
	if st == "" {
		st = "idle"
	}
	return herdr.AgentInfo{Name: target, Agent: "codex", AgentStatus: st, InteractiveReady: true}, nil
}
func (s *stubHerdrRPCClient) PromptAgent(ctx context.Context, req herdr.PromptAgentRequest) (herdr.AgentInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "agent.prompt:"+req.Target)
	s.lastPromptReq = req
	if s.onPrompt != nil {
		s.onPrompt(req)
	}
	st := s.agentStatus
	if st == "" {
		st = "idle"
	}
	return herdr.AgentInfo{Name: req.Target, AgentStatus: st}, nil
}
func (s *stubHerdrRPCClient) InterruptAgent(ctx context.Context, target string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, "agent.send_keys:ctrl+c:"+target)
	return nil
}
func (s *stubHerdrRPCClient) Calls() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.calls))
	copy(out, s.calls)
	return out
}
func (s *stubHerdrRPCClient) LastStartReq() herdr.StartAgentRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastStartReq
}
func (s *stubHerdrRPCClient) LastPromptReq() herdr.PromptAgentRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastPromptReq
}

func TestHerdrRunnerProtocol20ExactContract(t *testing.T) {
	// 1. Missing workspace must fail immediately (no create/rename)
	stubMissing := &stubHerdrRPCClient{
		workspaces: []herdr.WorkspaceInfo{{WorkspaceID: "w2", Label: "Other Workspace"}},
	}
	rMissing := NewHerdrRunner(stubMissing)
	_, err := rMissing.Start(context.Background(), qservant.JobRequest{Model: "openai/gpt-5.6-sol"})
	if err == nil {
		t.Fatal("expected error when Q Servant workspace is missing")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not_found error, got %v", err)
	}
	if containsMethod(stubMissing.Calls(), "workspace.create") || containsMethod(stubMissing.Calls(), "workspace.rename") {
		t.Fatalf("must not create or rename workspace on missing Q Servant space: %v", stubMissing.Calls())
	}

	// 2. Ambiguous workspace must fail immediately
	stubAmb := &stubHerdrRPCClient{
		workspaces: []herdr.WorkspaceInfo{
			{WorkspaceID: "w1", Label: "Q Servant"},
			{WorkspaceID: "w2", Label: "Q Servant"},
		},
	}
	rAmb := NewHerdrRunner(stubAmb)
	_, err = rAmb.Start(context.Background(), qservant.JobRequest{Model: "openai/gpt-5.6-sol"})
	if err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected ambiguous error, got %v", err)
	}

	// 3. New agent startup via anchor split when no idle q-servant pane exists
	tempDir := t.TempDir()
	stubSplit := &stubHerdrRPCClient{
		workspaces:  []herdr.WorkspaceInfo{{WorkspaceID: "w1", Label: "Q Servant"}},
		panes:       []herdr.PaneInfo{{PaneID: "w1:p1", WorkspaceID: "w1"}},
		splitPaneID: "w1:p2",
		onPrompt: func(req herdr.PromptAgentRequest) {
			path := filepath.Join(tempDir, "report-"+req.Target+".json")
			r := qservant.RunnerReport{
				Request:      "write code",
				Work:         "done",
				Verification: "tests passed",
				Changes:      []string{"file.go"},
				Result:       "success",
				Success:      true,
			}
			b, _ := json.Marshal(r)
			_ = os.WriteFile(path, b, 0600)
		},
	}
	rSplit := NewHerdrRunnerWithDir(stubSplit, tempDir)

	handle, err := rSplit.Start(context.Background(), qservant.JobRequest{
		Model:  "openai/gpt-5.6-sol",
		Effort: "high",
		Prompt: "write code",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check StartAgentRequest protocol-20 args
	startReq := stubSplit.LastStartReq()
	if !strings.HasPrefix(startReq.Name, "q-servant-") || startReq.Kind != "codex" || startReq.PaneID != "w1:p2" {
		t.Fatalf("bad start request: %+v", startReq)
	}
	if startReq.TimeoutMS == nil || *startReq.TimeoutMS != 60000 {
		t.Fatalf("expected timeout_ms 60000, got %v", startReq.TimeoutMS)
	}
	args := startReq.Args
	if len(args) != 6 || args[0] != "-m" || args[1] != "openai/gpt-5.6-sol" || args[2] != "-c" || args[3] != "model_reasoning_effort=\"high\"" || args[4] != "-c" {
		t.Fatalf("bad protocol-20 args: %v", args)
	}
	wantReportPath := filepath.Join(tempDir, "report-"+startReq.Name+".json")
	if !strings.HasPrefix(args[5], "developer_instructions=") || !strings.Contains(args[5], wantReportPath) {
		t.Fatalf("report contract must be a developer instruction for %s: %v", wantReportPath, args)
	}

	// Check split was called on anchor pane
	if !containsMethod(stubSplit.Calls(), "pane.split") {
		t.Fatal("expected pane.split on anchor pane")
	}
	if !containsMethod(stubSplit.Calls(), "agent.wait:") {
		t.Fatal("expected agent.wait before prompting newly started agent")
	}

	// 4. Completion validation via handle.Wait
	res, err := handle.Wait(context.Background())
	if err != nil || res.State != "completed" || !res.Report.Valid() {
		t.Fatalf("expected completed valid report, got res=%+v err=%v", res, err)
	}

	// 5. Check prompt request carried the exact target agent name (not hardcoded q-servant)
	promptReq := stubSplit.LastPromptReq()
	if promptReq.Target != startReq.Name || promptReq.Text != "write code" {
		t.Fatalf("prompt must target exact owned agent name %s: %+v", startReq.Name, promptReq)
	}
	if strings.Contains(promptReq.Text, "INSTRUCTION") || strings.Contains(promptReq.Text, "report-") || strings.Contains(promptReq.Text, "schema") {
		t.Fatalf("user prompt must not expose internal report contract: %q", promptReq.Text)
	}
	if promptReq.Wait == nil || promptReq.Wait.TimeoutMS == nil || *promptReq.Wait.TimeoutMS != 600000 {
		t.Fatalf("expected nested wait object with timeout_ms 600000, got %+v", promptReq.Wait)
	}

	// 6. Test cancellation uses structural Ctrl+C on the exact target agent name
	hCancel, err := rSplit.Start(context.Background(), qservant.JobRequest{Model: "openai/gpt-5.6-sol", Effort: "high"})
	if err != nil {
		t.Fatal(err)
	}
	if err := hCancel.Cancel(); err != nil {
		t.Fatal(err)
	}
	cancelAgent := hCancel.(*herdrRunHandle).agentName
	wantCancelKey := "agent.send_keys:ctrl+c:" + cancelAgent
	if !containsMethod(stubSplit.Calls(), wantCancelKey) {
		t.Fatalf("cancel must issue structural ctrl+c to exact agent: %v (wanted %s)", stubSplit.Calls(), wantCancelKey)
	}
}

func containsMethod(calls []string, method string) bool {
	for _, c := range calls {
		if strings.HasPrefix(c, method) {
			return true
		}
	}
	return false
}

func TestHerdrRunnerReportValidationAndSessionReuse(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Missing report file fails completion
	stubMissingReport := &stubHerdrRPCClient{
		workspaces:  []herdr.WorkspaceInfo{{WorkspaceID: "w1", Label: "Q Servant"}},
		panes:       []herdr.PaneInfo{{PaneID: "w1:p1", WorkspaceID: "w1", Agent: "codex"}},
		agentStatus: "idle",
	}
	r1 := NewHerdrRunnerWithDir(stubMissingReport, tempDir)
	h1, err := r1.Start(context.Background(), qservant.JobRequest{Model: "openai/gpt"})
	if err != nil {
		t.Fatal(err)
	}
	// Do not write report file -> Wait must fail
	_, err = h1.Wait(context.Background())
	if err == nil || !strings.Contains(err.Error(), "missing report") {
		t.Fatalf("expected missing report error, got %v", err)
	}

	// 2. Invalid report JSON fails with ErrInvalidReport
	stubInvalidReport := &stubHerdrRPCClient{
		workspaces:  []herdr.WorkspaceInfo{{WorkspaceID: "w1", Label: "Q Servant"}},
		panes:       []herdr.PaneInfo{{PaneID: "w1:p1", WorkspaceID: "w1", Agent: "codex"}},
		agentStatus: "idle",
		onPrompt: func(req herdr.PromptAgentRequest) {
			path := filepath.Join(tempDir, "report-"+req.Target+".json")
			_ = os.WriteFile(path, []byte(`{"request": null}`), 0600)
		},
	}
	r2 := NewHerdrRunnerWithDir(stubInvalidReport, tempDir)
	h2, err := r2.Start(context.Background(), qservant.JobRequest{Model: "openai/gpt"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = h2.Wait(context.Background())
	if !errors.Is(err, qservant.ErrInvalidReport) {
		t.Fatalf("expected ErrInvalidReport, got %v", err)
	}

	// 3. Non-terminal agent state (e.g. blocked) fails
	stubBlocked := &stubHerdrRPCClient{
		workspaces:  []herdr.WorkspaceInfo{{WorkspaceID: "w1", Label: "Q Servant"}},
		panes:       []herdr.PaneInfo{{PaneID: "w1:p1", WorkspaceID: "w1", Agent: "codex"}},
		agentStatus: "blocked",
	}
	r3 := NewHerdrRunnerWithDir(stubBlocked, tempDir)
	h3, err := r3.Start(context.Background(), qservant.JobRequest{Model: "openai/gpt"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = h3.Wait(context.Background())
	if err == nil || !strings.Contains(err.Error(), "non-terminal") {
		t.Fatalf("expected non-terminal state error, got %v", err)
	}

	// 4. Reuse idle-safe session with SAME model and effort
	stubReuse := &stubHerdrRPCClient{
		workspaces:       []herdr.WorkspaceInfo{{WorkspaceID: "w1", Label: "Q Servant"}},
		panes:            []herdr.PaneInfo{{PaneID: "w1:p1", WorkspaceID: "w1"}},
		splitPaneID:      "w1:p2",
		agentStatus:      "idle",
		agentInteractive: true,
	}
	r4 := NewHerdrRunnerWithDir(stubReuse, tempDir)
	// First start creates session on w1:p2 with model=m1, effort=high
	h4a, err := r4.Start(context.Background(), qservant.JobRequest{Model: "m1", Effort: "high"})
	if err != nil {
		t.Fatal(err)
	}
	// Add w1:p2 to panes so it exists for reuse
	stubReuse.mu.Lock()
	stubReuse.panes = append(stubReuse.panes, herdr.PaneInfo{PaneID: "w1:p2", WorkspaceID: "w1", Agent: "codex"})
	stubReuse.calls = nil
	stubReuse.mu.Unlock()

	// Second start with SAME model and effort reuses existing session without split
	h4b, err := r4.Start(context.Background(), qservant.JobRequest{Model: "m1", Effort: "high"})
	if err != nil {
		t.Fatal(err)
	}
	if containsMethod(stubReuse.Calls(), "pane.split") {
		t.Fatalf("must not split pane when reusing same model/effort session: %v", stubReuse.Calls())
	}
	if h4b.(*herdrRunHandle).agentName != h4a.(*herdrRunHandle).agentName {
		t.Fatalf("expected same agent %s, got %s", h4a.(*herdrRunHandle).agentName, h4b.(*herdrRunHandle).agentName)
	}
	if h4b.(*herdrRunHandle).reportPath != h4a.(*herdrRunHandle).reportPath {
		t.Fatalf("reused session must retain its hidden report path: %s != %s", h4b.(*herdrRunHandle).reportPath, h4a.(*herdrRunHandle).reportPath)
	}

	// Third start with DIFFERENT effort creates a new owned agent/pane via split
	stubReuse.mu.Lock()
	stubReuse.splitPaneID = "w1:p3"
	stubReuse.mu.Unlock()
	h4c, err := r4.Start(context.Background(), qservant.JobRequest{Model: "m1", Effort: "low"})
	if err != nil {
		t.Fatal(err)
	}
	if !containsMethod(stubReuse.Calls(), "pane.split") {
		t.Fatal("expected pane.split when model/effort differs")
	}
	if h4c.(*herdrRunHandle).agentName == h4a.(*herdrRunHandle).agentName {
		t.Fatal("different effort must create new owned agent")
	}

	// 5. Working agent with interactive_ready=true must NOT be reused (idle/done only)
	stubReuse.mu.Lock()
	stubReuse.agentStatus = "working"
	stubReuse.agentInteractive = true
	stubReuse.splitPaneID = "w1:p4"
	stubReuse.calls = nil
	stubReuse.mu.Unlock()

	h4d, err := r4.Start(context.Background(), qservant.JobRequest{Model: "m1", Effort: "low"})
	if err != nil {
		t.Fatal(err)
	}
	if !containsMethod(stubReuse.Calls(), "pane.split") {
		t.Fatal("working agent with interactive_ready=true must not be reused and must split")
	}
	if h4d.(*herdrRunHandle).agentName == h4c.(*herdrRunHandle).agentName {
		t.Fatal("working agent must create a new owned agent")
	}
}

func TestCatalogMappingAndJoinedQuota(t *testing.T) {
	// Hermetic: NewCatalog derives the codex config path from $HOME.
	t.Setenv("HOME", t.TempDir())

	quotaClient := qservant.NewQuotaClient(&fakeCmdRunner{b: []byte(`{"reports":[{"provider":"openai","quota":{"fiveHourPercent":25}}]}`)}, time.Hour)

	modelsJSON := []byte(`[{"provider":"openai","id":"gpt-5.6-sol","namespaced":"openai/gpt-5.6-sol","reasoningEfforts":["low","high"],"defaultReasoningEffort":"high"}]`)
	cat := qservant.NewCatalog(&fakeCmdRunner{b: modelsJSON}, "/missing", nil)
	qm := NewQServantManager(cat, quotaClient, nil, nil, nil)

	info, err := qm.Catalog(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.DefaultModel != "openai/gpt-5.6-sol" || info.DefaultEffort != "high" {
		t.Fatalf("bad default: %+v", info)
	}
	models := info.Models.([]CatalogModelView)
	if len(models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(models))
	}
	m0 := models[0]
	if m0.ID != "openai/gpt-5.6-sol" || m0.Label != "gpt-5.6-sol" || len(m0.Efforts) != 2 || m0.DefaultEffort != "high" {
		t.Fatalf("bad mapped model: %+v", m0)
	}
	if m0.Quota == nil || m0.Quota.Used == nil || *m0.Quota.Used != 0.25 || m0.Quota.Label != "25.0% used (5h)" {
		t.Fatalf("expected joined quota on model with used=0.25, got %+v", m0.Quota)
	}
}

func TestSwiftSTTAdapter(t *testing.T) {
	fakeRunner := &fakeCmdRunner{b: []byte("안녕하세요 테스트입니다\n")}
	stt := NewSwiftSTTWithRunner(fakeRunner)
	txt, err := stt.Transcribe(context.Background(), "/path/to/audio.m4a", "ko-KR")
	// On darwin with fake runner, it succeeds; on non-darwin it returns ErrSTTUnavailable
	if err != nil && !errors.Is(err, qservant.ErrSTTUnavailable) {
		t.Fatalf("unexpected STT error: %v", err)
	}
	if err == nil && txt != "안녕하세요 테스트입니다" {
		t.Fatalf("unexpected transcribed text: %q", txt)
	}
}

func TestEngineQServantWiringAndBroadcast(t *testing.T) {
	dir := t.TempDir()
	report := qservant.RunnerReport{Request: map[string]any{}, Work: map[string]any{}, Verification: map[string]any{}, Changes: map[string]any{}, Result: map[string]any{}, Success: true}
	runner := &qservant.FakeRunner{Result: qservant.RunnerResult{State: "completed", Report: report}}

	f := newFakeHerdr(t)
	e := New(Config{SocketPath: f.SocketPath(), ListenAddr: "127.0.0.1:0", StateDir: dir})
	ctrl := qservant.NewJobController(dir, runner)
	cat := qservant.NewCatalog(&fakeCmdRunner{b: []byte(`[{"provider":"openai","id":"gpt","namespaced":"openai/gpt"}]`)}, "/missing", nil)
	qm := NewQServantManager(cat, nil, ctrl, &qservant.FakeSTT{Text: "transcribed audio"}, e.srv.Broadcast)
	e.SetQServant(qm)

	srv := httptest.NewServer(e.srv.Handler())
	defer srv.Close()
	ctx := context.Background()
	c := dialWS(t, srv)
	defer c.Close(websocket.StatusNormalClosure, "")

	// Drain initial welcome and snapshot
	_ = readFrame(t, c) // welcome
	_ = readFrame(t, c) // panes
	_ = readFrame(t, c) // workspaces
	_ = readFrame(t, c) // tabs

	// Submit job via WS
	submitFrame := []byte(`{"t":"qservant_submit","reqId":"rs1","model":"openai/gpt","audioMime":"audio/mp4","audioBase64":"YWFj"}`)
	c.Write(ctx, websocket.MessageText, submitFrame)

	// Expect direct submit response
	f1 := readFrame(t, c)
	if f1["t"] != "qservant_job" || f1["reqId"] != "rs1" {
		t.Fatalf("bad initial submit response: %+v", f1)
	}
	jobObj, ok := f1["job"].(map[string]any)
	if !ok || jobObj["transcript"] != "transcribed audio" {
		t.Fatalf("bad nested job object in submit response: %+v", f1)
	}

	// Expect broadcast frame for completion
	deadline := time.After(2 * time.Second)
	completed := false
	for !completed {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for completed broadcast frame")
		default:
			f := readFrame(t, c)
			if f["t"] == "qservant_job" {
				if bJob, ok := f["job"].(map[string]any); ok && bJob["state"] == "completed" {
					if f["reqId"] != nil {
						t.Fatalf("broadcast frame must omit reqId, got %v", f["reqId"])
					}
					completed = true
				}
			}
		}
	}
}

type fakeCmdRunner struct {
	b []byte
}

func (f *fakeCmdRunner) Run(context.Context, string, ...string) ([]byte, error) {
	return f.b, nil
}
