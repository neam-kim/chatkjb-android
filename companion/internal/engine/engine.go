package engine

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/herdr"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/notify"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/proto"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/state"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/wsserver"
)

type Config struct {
	SocketPath       string
	ListenAddr       string
	PollInterval     time.Duration
	DebounceFinished time.Duration
	// PushEndpointPath persists the phone's UnifiedPush endpoint across daemon
	// restarts. Empty keeps the old in-memory-only behaviour, which is useful
	// for isolated tests.
	PushEndpointPath string
	// PushRegistrationTokenPath stores the bearer token accepted by the narrow
	// HTTPS endpoint used by the current ChatKJB app to register UnifiedPush.
	PushRegistrationTokenPath string
	// PushEndpointOrigin optionally restricts registered endpoints to one
	// trusted UnifiedPush distributor origin.
	PushEndpointOrigin string
	// NotifySpaces restricts push notifications to the named Herdr spaces
	// (workspace labels), case-insensitively. Empty means notify for every
	// space. OCA/subagent spaces are excluded by listing only "General".
	NotifySpaces []string
}

type Engine struct {
	cfg    Config
	client *herdr.Client
	store  *state.Store
	srv    *wsserver.Server

	mu                    sync.Mutex
	endpoint              string
	pushRegistrationToken string

	trigger chan struct{}
	subs    map[string]context.CancelFunc
}

func New(cfg Config) *Engine {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 1500 * time.Millisecond
	}
	if cfg.DebounceFinished == 0 {
		cfg.DebounceFinished = 4 * time.Second
	}
	c := herdr.New(cfg.SocketPath)
	e := &Engine{cfg: cfg, client: c, store: state.NewStore()}
	if endpoint, err := loadPushEndpoint(cfg.PushEndpointPath); err != nil {
		log.Printf("herdr-mobiled: load push endpoint: %v", err)
	} else {
		e.endpoint = endpoint
	}
	if token, err := loadOrCreatePushRegistrationToken(cfg.PushRegistrationTokenPath); err != nil {
		log.Printf("herdr-mobiled: load push registration token: %v", err)
	} else {
		e.pushRegistrationToken = token
	}
	e.srv = wsserver.NewServer(wsserver.AllowAll{}, c)
	e.srv.SetInitialSnapshot(e.store.Snapshot)
	e.srv.SetWorkspaceSnapshot(e.store.Workspaces)
	e.srv.SetTabSnapshot(e.store.Tabs)
	e.srv.SetPushEndpoint(e.setEndpoint)
	e.srv.SetPoke(e.Poke)
	e.trigger = make(chan struct{}, 1)
	e.subs = map[string]context.CancelFunc{}
	return e
}

func (e *Engine) setEndpoint(ep string) {
	ep = strings.TrimSpace(ep)
	e.mu.Lock()
	e.endpoint = ep
	e.mu.Unlock()
	if err := savePushEndpoint(e.cfg.PushEndpointPath, ep); err != nil {
		log.Printf("herdr-mobiled: save push endpoint: %v", err)
	}
}

func loadPushEndpoint(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func savePushEndpoint(path, endpoint string) error {
	if path == "" {
		return nil
	}
	if endpoint == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}

	return savePrivateValue(path, endpoint)
}

func loadOrCreatePushRegistrationToken(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	if token, err := loadPushEndpoint(path); err != nil {
		return "", err
	} else if token != "" {
		return token, nil
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := fmt.Sprintf("%x", raw)
	if err := savePrivateValue(path, token); err != nil {
		return "", err
	}
	return token, nil
}

func savePrivateValue(path, value string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".push-endpoint-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := fmt.Fprintln(tmp, value); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (e *Engine) handlePushRegistration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if e.pushRegistrationToken == "" {
		http.Error(w, "push registration unavailable", http.StatusServiceUnavailable)
		return
	}
	presented := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if len(presented) != len(e.pushRegistrationToken) ||
		subtle.ConstantTimeCompare([]byte(presented), []byte(e.pushRegistrationToken)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	var body struct {
		Endpoint string `json:"endpoint"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil || !e.allowedPushEndpoint(body.Endpoint) {
		http.Error(w, "invalid endpoint", http.StatusBadRequest)
		return
	}
	e.setEndpoint(body.Endpoint)
	w.WriteHeader(http.StatusNoContent)
}

func (e *Engine) allowedPushEndpoint(endpoint string) bool {
	candidate, err := url.ParseRequestURI(strings.TrimSpace(endpoint))
	if err != nil || candidate.User != nil || candidate.Fragment != "" || candidate.Host == "" {
		return false
	}
	if candidate.Scheme != "http" && candidate.Scheme != "https" {
		return false
	}
	if e.cfg.PushEndpointOrigin == "" {
		return true
	}
	allowed, err := url.ParseRequestURI(e.cfg.PushEndpointOrigin)
	return err == nil && strings.EqualFold(candidate.Scheme, allowed.Scheme) &&
		strings.EqualFold(candidate.Host, allowed.Host)
}

// Poke requests an immediate poll (coalesced with the ticker). Used by the
// wsserver after a successful structural action so the tree refreshes fast.
func (e *Engine) Poke() {
	select {
	case e.trigger <- struct{}{}:
	default:
	}
}

func (e *Engine) Run(ctx context.Context) error {
	// probe herdr version for the welcome frame (best-effort)
	if raw, err := e.client.Call(ctx, "ping", nil); err == nil {
		var pong struct {
			Version  string `json:"version"`
			Protocol int    `json:"protocol"`
		}
		_ = json.Unmarshal(raw, &pong)
		e.srv.SetHerdrInfo(pong.Version, pong.Protocol)
	}

	go e.pollLoop(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/notify/register", e.handlePushRegistration)
	mux.HandleFunc("/register", e.handlePushRegistration)
	mux.Handle("/", e.srv.Handler())
	httpSrv := &http.Server{Addr: e.cfg.ListenAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		httpSrv.Shutdown(shutdownCtx)
	}()
	err := httpSrv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (e *Engine) pollLoop(ctx context.Context) {
	t := time.NewTicker(e.cfg.PollInterval)
	defer t.Stop()
	defer e.cancelAllSubs()
	// immediate first poll + subscribe so we don't wait a full interval
	e.pollOnce(ctx)
	e.reconcileSubs(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			e.pollOnce(ctx)
			e.reconcileSubs(ctx)
		case <-e.trigger:
			e.pollOnce(ctx)
			e.reconcileSubs(ctx)
		}
	}
}

// reconcileSubs opens a per-pane agent_status_changed subscription for every
// agent-bearing pane and cancels subscriptions for panes that are gone. Runs
// only on the poll goroutine, so e.subs needs no lock. Subscription events
// poke e.trigger (coalesced), causing an immediate pollOnce - the store stays
// the single transition source, so notifications never double-fire.
func (e *Engine) reconcileSubs(ctx context.Context) {
	desired := map[string]bool{}
	for _, p := range e.store.Snapshot() {
		if p.Agent != "" {
			desired[p.PaneID] = true
		}
	}
	for id := range desired {
		if _, ok := e.subs[id]; ok {
			continue
		}
		cctx, cancel := context.WithCancel(ctx)
		ch, err := e.client.Subscribe(cctx, id, "pane.agent_status_changed")
		if err != nil {
			cancel()
			continue
		}
		e.subs[id] = cancel
		go e.drainSub(ch)
	}
	for id, cancel := range e.subs {
		if !desired[id] {
			cancel()
			delete(e.subs, id)
		}
	}
}

func (e *Engine) drainSub(ch <-chan herdr.Event) {
	for range ch {
		select {
		case e.trigger <- struct{}{}:
		default:
		}
	}
}

func (e *Engine) cancelAllSubs() {
	for id, cancel := range e.subs {
		cancel()
		delete(e.subs, id)
	}
}

func (e *Engine) pollOnce(ctx context.Context) {
	panes, err := e.client.ListPanes(ctx)
	if err != nil {
		return
	}
	changes, transitions := e.store.Apply(panes)
	for _, ch := range changes {
		if ch.Kind == "removed" {
			e.srv.Broadcast(proto.PaneRemoved(ch.PaneID))
		} else {
			e.srv.Broadcast(proto.PaneUpdate(ch.Pane))
		}
	}
	for _, tr := range transitions {
		e.handleTransition(ctx, tr)
	}

	if ws, err := e.client.ListWorkspaces(ctx); err == nil {
		if e.store.ApplyWorkspaces(ws) {
			e.srv.Broadcast(proto.WorkspacesSnapshot(e.store.Workspaces()))
		}
	}
	if tabs, err := e.client.ListTabs(ctx); err == nil {
		if e.store.ApplyTabs(tabs) {
			e.srv.Broadcast(proto.TabsSnapshot(e.store.Tabs()))
		}
	}
}

func (e *Engine) handleTransition(ctx context.Context, tr state.Transition) {
	push, ok := notify.ShouldNotify(tr, e.displayName(tr.PaneID), "")
	if !ok {
		return
	}
	// "clear" carries no text and only dismisses an existing notification, so
	// it must still reach the phone even for a muted space; otherwise a stale
	// blocked notification would never be retracted.
	if push.Kind != "clear" && !e.spaceNotifies(tr.WorkspaceID) {
		return
	}
	if tr.To == "blocked" {
		if txt, err := e.client.ReadPane(ctx, tr.PaneID, "detection", 40); err == nil {
			push.Body = lastNonEmptyLine(txt)
		}
	}
	if push.Kind == "finished" {
		// debounce: only fire if still not working after the window
		go func() {
			select {
			case <-ctx.Done():
			case <-time.After(e.cfg.DebounceFinished):
				for _, p := range e.store.Snapshot() {
					if p.PaneID == tr.PaneID && p.AgentStatus == "working" {
						return // resumed; suppress
					}
				}
				// Summarize after the debounce so the body reflects the
				// settled final output rather than a mid-run frame. Use the
				// unwrapped snapshot: the wrapped view splits a sentence
				// across rows, which would surface a mid-sentence fragment.
				// The socket API spells this "recent_unwrapped"; the CLI's
				// hyphenated spelling is rejected over the wire.
				push.Body = e.summarizePane(ctx, tr.PaneID)
				e.fire(ctx, push)
			}
		}()
		return
	}
	e.fire(ctx, push)
}

// summarizePane builds the short "what finished" line for a completed pane.
// It prefers the unwrapped snapshot so hard-wrapped prose is not cut
// mid-sentence, and falls back to the wrapped view when a herdr build does
// not offer that source.
func (e *Engine) summarizePane(ctx context.Context, paneID string) string {
	for _, source := range []string{"recent_unwrapped", "recent", "detection"} {
		txt, err := e.client.ReadPane(ctx, paneID, source, 60)
		if err != nil {
			continue
		}
		if s := notify.Summarize(txt); s != "" {
			return s
		}
	}
	return ""
}

// spaceNotifies reports whether a workspace is allowed to raise notifications.
// The match is on the space label from herdr (e.g. "General"), because
// workspace ids like "w1C" are unstable across restarts. A workspace missing
// from the snapshot is treated as not allowed once a filter is configured, so
// newly spawned OCA spaces stay silent until they are identified.
func (e *Engine) spaceNotifies(workspaceID string) bool {
	if len(e.cfg.NotifySpaces) == 0 {
		return true
	}
	label := ""
	for _, w := range e.store.Workspaces() {
		if w.WorkspaceID == workspaceID {
			label = w.Label
			break
		}
	}
	if label == "" {
		return false
	}
	for _, allowed := range e.cfg.NotifySpaces {
		if strings.EqualFold(strings.TrimSpace(allowed), label) {
			return true
		}
	}
	return false
}

// displayName returns the friendly pane name for notification titles: the cwd
// basename (the project folder, e.g. "omega3") if present, else the workspace
// id. Read from the store snapshot since a Transition carries only the id.
func (e *Engine) displayName(paneID string) string {
	for _, p := range e.store.Snapshot() {
		if p.PaneID == paneID {
			if b := filepath.Base(p.CWD); b != "" && b != "." && b != string(filepath.Separator) {
				return b
			}
			return p.WorkspaceID
		}
	}
	return ""
}

func (e *Engine) fire(ctx context.Context, p notify.Push) {
	e.mu.Lock()
	ep := e.endpoint
	e.mu.Unlock()
	if ep == "" {
		return
	}
	n := notify.NewHTTPNotifier(ep, http.DefaultClient)
	_ = n.Notify(ctx, p)
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			return strings.TrimSpace(lines[i])
		}
	}
	return ""
}
