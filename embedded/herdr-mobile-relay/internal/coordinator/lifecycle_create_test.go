package coordinator

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/0cv/herdr-mobile-relay/internal/herdr"
	"github.com/0cv/herdr-mobile-relay/internal/profiles"
)

func TestLifecycleStartsAgentInNestedWorkspaceCreateRootPane(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	cwd := filepath.Join(home, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("create project directory: %v", err)
	}
	record := filepath.Join(dir, "invocations.log")
	bin := writeScript(t, dir, "herdr", "#!/bin/sh\n"+
		"printf '%s\\n' \"$*\" >> \""+record+"\"\n"+
		"case \"$1 $2\" in\n"+
		"  'pane list') printf '%s\\n' '{\"result\":{\"panes\":[]}}' ;;\n"+
		"  'workspace list') printf '%s\\n' '{\"result\":{\"workspaces\":[]}}' ;;\n"+
		"  'workspace create') printf '%s\\n' '{\"result\":{\"type\":\"workspace_created\",\"workspace\":{\"workspace_id\":\"workspace-new\",\"label\":\"project\"},\"tab\":{\"tab_id\":\"tab-new\",\"workspace_id\":\"workspace-new\",\"label\":\"project\"},\"root_pane\":{\"pane_id\":\"pane-new\",\"tab_id\":\"tab-new\",\"workspace_id\":\"workspace-new\"}}}' ;;\n"+
		"  'tab rename') printf '%s\\n' '{\"result\":{\"tab_id\":\"tab-new\"}}' ;;\n"+
		"  'agent start') printf '%s\\n' '{\"result\":{\"type\":\"agent_started\",\"agent\":{\"pane_id\":\"pane-new\",\"agent\":\"codex\",\"name\":\"project-codex\"}}}' ;;\n"+
		"  *) exit 2 ;;\n"+
		"esac\n")

	resolver := profiles.NewResolver(filepath.Join(dir, "config"), nil)
	lifecycle := &Lifecycle{
		herdr:    herdr.NewClient(bin, filepath.Join(dir, "herdr.sock")),
		profiles: resolver,
		home:     home,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := lifecycle.Start(ctx, profiles.Profile{ID: "codex", Kind: "codex"}, StartRequest{
		ProfileID: "codex",
		Name:      "project-codex",
		Cwd:       cwd,
	})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if result.PaneID != "pane-new" {
		t.Fatalf("Start() pane_id = %q, want pane-new", result.PaneID)
	}

	invocations, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("read invocations: %v", err)
	}
	if !strings.Contains(string(invocations), "agent start project-codex --kind codex --pane pane-new") {
		t.Fatalf("agent was not started in created root pane:\n%s", invocations)
	}
}

func TestLifecycleStartsAgentInExplicitWorkspace(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	cwd := filepath.Join(home, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(dir, "invocations.log")
	bin := writeScript(t, dir, "herdr", "#!/bin/sh\n"+
		"printf '%s\\n' \"$*\" >> \""+record+"\"\n"+
		"case \"$1 $2\" in\n"+
		"  'pane list') printf '%s\\n' '{\"result\":{\"panes\":[]}}' ;;\n"+
		"  'workspace list') printf '%s\\n' '{\"result\":{\"workspaces\":[{\"workspace_id\":\"workspace-existing\",\"label\":\"Project\"}]}}' ;;\n"+
		"  'tab create') printf '%s\\n' '{\"result\":{\"type\":\"tab_created\",\"tab\":{\"tab_id\":\"tab-new\",\"workspace_id\":\"workspace-existing\"},\"root_pane\":{\"pane_id\":\"pane-new\",\"tab_id\":\"tab-new\",\"workspace_id\":\"workspace-existing\"}}}' ;;\n"+
		"  'agent start') printf '%s\\n' '{\"result\":{\"type\":\"agent_started\",\"agent\":{\"pane_id\":\"pane-new\",\"agent\":\"codex\",\"name\":\"project-codex\"}}}' ;;\n"+
		"  *) exit 2 ;;\n"+
		"esac\n")

	lifecycle := &Lifecycle{
		herdr:    herdr.NewClient(bin, filepath.Join(dir, "herdr.sock")),
		profiles: profiles.NewResolver(filepath.Join(dir, "config"), nil),
		home:     home,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := lifecycle.Start(ctx, profiles.Profile{ID: "codex", Kind: "codex"}, StartRequest{
		ProfileID:   "codex",
		WorkspaceID: "workspace-existing",
		Name:        "project-codex",
		Cwd:         cwd,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.WorkspaceID != "workspace-existing" || result.PaneID != "pane-new" {
		t.Fatalf("result = %+v", result)
	}
	invocations, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(invocations), "tab create --workspace workspace-existing") {
		t.Fatalf("explicit workspace was not used:\n%s", invocations)
	}
	if strings.Contains(string(invocations), "workspace create") {
		t.Fatalf("unexpected workspace create:\n%s", invocations)
	}
}

func TestLifecycleStartsShellInExplicitWorkspace(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	cwd := filepath.Join(home, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(dir, "invocations.log")
	bin := writeScript(t, dir, "herdr", "#!/bin/sh\n"+
		"printf '%s\\n' \"$*\" >> \""+record+"\"\n"+
		"case \"$1 $2\" in\n"+
		"  'pane list') printf '%s\\n' '{\"result\":{\"panes\":[]}}' ;;\n"+
		"  'workspace list') printf '%s\\n' '{\"result\":{\"workspaces\":[{\"workspace_id\":\"workspace-existing\",\"label\":\"Project\"}]}}' ;;\n"+
		"  'tab create') printf '%s\\n' '{\"result\":{\"type\":\"tab_created\",\"tab\":{\"tab_id\":\"tab-shell\",\"workspace_id\":\"workspace-existing\"},\"root_pane\":{\"pane_id\":\"pane-shell\",\"tab_id\":\"tab-shell\",\"workspace_id\":\"workspace-existing\"}}}' ;;\n"+
		"  *) exit 2 ;;\n"+
		"esac\n")
	lifecycle := &Lifecycle{
		herdr: herdr.NewClient(bin, filepath.Join(dir, "herdr.sock")),
		home:  home,
	}

	result, err := lifecycle.StartShell(context.Background(), "workspace-existing", "project-shell", cwd)
	if err != nil {
		t.Fatal(err)
	}
	if result.PaneID != "pane-shell" || result.WorkspaceID != "workspace-existing" {
		t.Fatalf("result = %+v", result)
	}
	invocations, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(invocations), "tab create --workspace workspace-existing") {
		t.Fatalf("shell tab was not created in target workspace:\n%s", invocations)
	}
	if strings.Contains(string(invocations), "agent start") {
		t.Fatalf("shell creation unexpectedly started an agent:\n%s", invocations)
	}
}

func TestSentinelStartsSolHighInExplicitWorkspace(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	cwd := filepath.Join(home, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	record := filepath.Join(dir, "invocations.log")
	bin := writeScript(t, dir, "herdr", "#!/bin/sh\n"+
		"printf '%s\\n' \"$*\" >> \""+record+"\"\n"+
		"case \"$1 $2\" in\n"+
		"  'pane list') printf '%s\\n' '{\"result\":{\"panes\":[]}}' ;;\n"+
		"  'workspace list') printf '%s\\n' '{\"result\":{\"workspaces\":[{\"workspace_id\":\"workspace-existing\",\"label\":\"Project\"}]}}' ;;\n"+
		"  'tab create') printf '%s\\n' '{\"result\":{\"type\":\"tab_created\",\"tab\":{\"tab_id\":\"tab-new\",\"workspace_id\":\"workspace-existing\"},\"root_pane\":{\"pane_id\":\"pane-new\",\"tab_id\":\"tab-new\",\"workspace_id\":\"workspace-existing\"}}}' ;;\n"+
		"  'agent start') printf '%s\\n' '{\"result\":{\"type\":\"agent_started\",\"agent\":{\"pane_id\":\"pane-new\",\"agent\":\"codex\",\"name\":\"project-codex\"}}}' ;;\n"+
		"  *) exit 2 ;;\n"+
		"esac\n")

	lifecycle := &Lifecycle{
		herdr:    herdr.NewClient(bin, filepath.Join(dir, "herdr.sock")),
		profiles: profiles.NewResolver(filepath.Join(dir, "config"), nil),
		home:     home,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := lifecycle.Start(ctx, profiles.Profile{ID: "codex", Kind: "codex"}, StartRequest{
		Preset:      "sentinel-sol-high",
		ProfileID:   "codex",
		WorkspaceID: "workspace-existing",
		Name:        "project-codex",
		Cwd:         cwd,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.WorkspaceID != "workspace-existing" || result.PaneID != "pane-new" {
		t.Fatalf("result = %+v", result)
	}
	invocations, err := os.ReadFile(record)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(invocations), "tab create --workspace workspace-existing") {
		t.Fatalf("explicit workspace was not used:\n%s", invocations)
	}
	if strings.Contains(string(invocations), "workspace create") {
		t.Fatalf("unexpected workspace create:\n%s", invocations)
	}
	if !strings.Contains(string(invocations), `-- --model gpt-5.6-sol -c model_reasoning_effort="high"`) {
		t.Fatalf("Sentinel did not preserve Sol/high argv: %s", invocations)
	}

}

func TestSentinelPresetRejectsWrongTargetBeforeProfileLookup(t *testing.T) {
	lifecycle := &Lifecycle{}
	for _, request := range []StartRequest{
		{Preset: "unknown"},
		{Preset: "sentinel-sol-high", ProfileID: "claude", WorkspaceID: "w1", Cwd: "/Volumes/NEAM_SSD/security-sentinel"},
		{Preset: "sentinel-sol-high", ProfileID: "codex", Cwd: "/Volumes/NEAM_SSD/security-sentinel"},
		{Preset: "sentinel-sol-high", ProfileID: "codex", WorkspaceID: "w1", Cwd: "/Users/neam"},
	} {
		if _, _, err := lifecycle.ValidateStart(request); err == nil {
			t.Fatalf("accepted invalid preset: %+v", request)
		}
	}
}

func TestSentinelValidateCanonicalProject(t *testing.T) {
	const project = "/Volumes/NEAM_SSD/security-sentinel"
	if _, err := os.Stat(project); err != nil {
		t.Skip("canonical Sentinel project is not mounted on this host")
	}
	dir := t.TempDir()
	writeScript(t, dir, "codex", "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", dir)
	lifecycle := &Lifecycle{profiles: profiles.NewResolver(dir, nil), home: t.TempDir()}
	profile, request, err := lifecycle.ValidateStart(StartRequest{
		Preset: "sentinel-sol-high", ProfileID: "codex", WorkspaceID: "w1",
		Name: "sentinel-test", Cwd: project, Prompt: "Read-only test",
	})
	if err != nil || profile.ID != "codex" || request.Cwd != project {
		t.Fatalf("canonical preset admission failed: profile=%+v request=%+v err=%v", profile, request, err)
	}
	if _, err := lifecycle.ResolveCwd(project); err == nil {
		t.Fatal("generic launcher incorrectly inherited the preset exception")
	}
}
