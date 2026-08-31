package herdr

import (
	"encoding/json"
	"fmt"
)

type RPCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *RPCError) Error() string { return e.Code + ": " + e.Message }

type Response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
}

type PaneInfo struct {
	PaneID      string `json:"pane_id"`
	WorkspaceID string `json:"workspace_id"`
	TabID       string `json:"tab_id"`
	CWD         string `json:"cwd"`
	Focused     bool   `json:"focused"`
	Agent       string `json:"agent"`
	AgentStatus string `json:"agent_status"`
	TerminalID  string `json:"terminal_id"`
}

type paneListResult struct {
	Type  string     `json:"type"`
	Panes []PaneInfo `json:"panes"`
}

type WorktreeInfo struct {
	RepoName         string `json:"repo_name"`
	RepoRoot         string `json:"repo_root"`
	CheckoutPath     string `json:"checkout_path"`
	IsLinkedWorktree bool   `json:"is_linked_worktree"`
}

type WorkspaceInfo struct {
	WorkspaceID string        `json:"workspace_id"`
	Label       string        `json:"label"`
	Number      int           `json:"number"`
	ActiveTabID string        `json:"active_tab_id"`
	AgentStatus string        `json:"agent_status"`
	Focused     bool          `json:"focused"`
	PaneCount   int           `json:"pane_count"`
	TabCount    int           `json:"tab_count"`
	Worktree    *WorktreeInfo `json:"worktree,omitempty"`
}

type workspaceListResult struct {
	Type       string          `json:"type"`
	Workspaces []WorkspaceInfo `json:"workspaces"`
}

type TabInfo struct {
	TabID       string `json:"tab_id"`
	Label       string `json:"label"`
	Number      int    `json:"number"`
	WorkspaceID string `json:"workspace_id"`
	AgentStatus string `json:"agent_status"`
	Focused     bool   `json:"focused"`
	PaneCount   int    `json:"pane_count"`
}

type tabListResult struct {
	Type string    `json:"type"`
	Tabs []TabInfo `json:"tabs"`
}

type paneReadResult struct {
	Type string `json:"type"`
	Read struct {
		PaneID string `json:"pane_id"`
		Source string `json:"source"`
		Text   string `json:"text"`
	} `json:"read"`
}

// Event is a frame pushed on a subscription connection.
type Event struct {
	Type        string `json:"type"`
	PaneID      string `json:"pane_id"`
	AgentStatus string `json:"agent_status"`
}

// paneRef is the {pane_id, terminal_id} carried by create/split/start results.
type paneRef struct {
	PaneID     string `json:"pane_id"`
	TerminalID string `json:"terminal_id"`
}
type paneInfoResult struct {
	Pane PaneInfo `json:"pane"`
}
type rootPaneResult struct {
	RootPane paneRef `json:"root_pane"`
}

// QServantWorkspaceLabel is the exact Herdr workspace label Q Servant targets.
// Selection is case-sensitive equality against workspace.list; there is no
// fallback to "mobile" or any other space.
const QServantWorkspaceLabel = "Q Servant"

// AgentKindCodex is the protocol-20 agent.start kind used by Q Servant.
const AgentKindCodex = "codex"

// InterruptKeyCtrlC is the structural key combo issued on cancel.
const InterruptKeyCtrlC = "ctrl+c"

// WorkspaceError is returned when exact-label workspace selection fails.
type WorkspaceError struct {
	Code    string // "not_found" or "ambiguous"
	Label   string
	Message string
}

func (e *WorkspaceError) Error() string { return e.Message }

func IsWorkspaceNotFound(err error) bool {
	we, ok := err.(*WorkspaceError)
	return ok && we.Code == "not_found"
}

func IsWorkspaceAmbiguous(err error) bool {
	we, ok := err.(*WorkspaceError)
	return ok && we.Code == "ambiguous"
}

// SelectWorkspaceByLabel returns the unique workspace whose Label equals
// label exactly. Zero matches is not_found; two or more is ambiguous.
func SelectWorkspaceByLabel(workspaces []WorkspaceInfo, label string) (WorkspaceInfo, error) {
	var matches []WorkspaceInfo
	for _, ws := range workspaces {
		if ws.Label == label {
			matches = append(matches, ws)
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return WorkspaceInfo{}, &WorkspaceError{
			Code:    "not_found",
			Label:   label,
			Message: fmt.Sprintf("herdr space '%s' not found", label),
		}
	default:
		return WorkspaceInfo{}, &WorkspaceError{
			Code:    "ambiguous",
			Label:   label,
			Message: fmt.Sprintf("herdr space '%s' is ambiguous", label),
		}
	}
}

// AgentInfo is the protocol-20 agent object returned by agent.start/get/prompt/list.
type AgentInfo struct {
	Name             string `json:"name"`
	PaneID           string `json:"pane_id"`
	WorkspaceID      string `json:"workspace_id"`
	TabID            string `json:"tab_id"`
	TerminalID       string `json:"terminal_id"`
	Agent            string `json:"agent"`
	AgentStatus      string `json:"agent_status"`
	Focused          bool   `json:"focused"`
	CWD              string `json:"cwd"`
	InteractiveReady bool   `json:"interactive_ready"`
}

// StartAgentRequest is the protocol-20 agent.start params object.
// Required: Name, Kind, PaneID. Args is always encoded as an array.
// TimeoutMS is encoded as timeout_ms (not timeout).
type StartAgentRequest struct {
	Name      string   `json:"name"`
	Kind      string   `json:"kind"`
	PaneID    string   `json:"pane_id"`
	Args      []string `json:"args"`
	TimeoutMS *uint64  `json:"timeout_ms,omitempty"`
}

func (r StartAgentRequest) params() map[string]any {
	args := r.Args
	if args == nil {
		args = []string{}
	}
	p := map[string]any{
		"name":    r.Name,
		"kind":    r.Kind,
		"pane_id": r.PaneID,
		"args":    args,
	}
	if r.TimeoutMS != nil {
		p["timeout_ms"] = *r.TimeoutMS
	}
	return p
}

// TimeoutMillis returns a pointer suitable for StartAgentRequest.TimeoutMS
// and AgentWaitOptions.TimeoutMS (encoded as timeout_ms).
func TimeoutMillis(ms uint64) *uint64 {
	v := ms
	return &v
}

// AgentWaitOptions is the protocol-20 wait object on agent.prompt / agent.wait.
type AgentWaitOptions struct {
	TimeoutMS *uint64  `json:"timeout_ms,omitempty"`
	Until     []string `json:"until,omitempty"`
}

// PromptAgentRequest is the protocol-20 agent.prompt params object.
type PromptAgentRequest struct {
	Target string            `json:"target"`
	Text   string            `json:"text"`
	Wait   *AgentWaitOptions `json:"wait,omitempty"`
}

func (r PromptAgentRequest) params() map[string]any {
	p := map[string]any{
		"target": r.Target,
		"text":   r.Text,
	}
	if r.Wait != nil {
		p["wait"] = r.Wait
	}
	return p
}

type agentStartedResult struct {
	Type  string    `json:"type"`
	Agent AgentInfo `json:"agent"`
	Argv  []string  `json:"argv"`
}
type agentInfoResult struct {
	Type  string    `json:"type"`
	Agent AgentInfo `json:"agent"`
}
type agentListResult struct {
	Type   string      `json:"type"`
	Agents []AgentInfo `json:"agents"`
}
type agentReadResult struct {
	Type string `json:"type"`
	Read struct {
		PaneID string `json:"pane_id"`
		Source string `json:"source"`
		Text   string `json:"text"`
	} `json:"read"`
}
type agentManifestsResult struct {
	Manifests []struct {
		Agent string `json:"agent"`
	} `json:"manifests"`
}

// WorktreeEntry is one entry from herdr's worktree.list (scoped to one repo).
// Branch and OpenWorkspaceID are Option in herdr and arrive absent (→ "")
// when None. Label is the repo name (same for every entry), so it is NOT a
// per-sibling label — resolve sibling labels from the workspace snapshot.
type WorktreeEntry struct {
	Path             string `json:"path"`
	Branch           string `json:"branch"`
	IsLinkedWorktree bool   `json:"is_linked_worktree"`
	OpenWorkspaceID  string `json:"open_workspace_id"`
	Label            string `json:"label"`
}

type worktreeListResult struct {
	Type      string          `json:"type"`
	Worktrees []WorktreeEntry `json:"worktrees"`
}
