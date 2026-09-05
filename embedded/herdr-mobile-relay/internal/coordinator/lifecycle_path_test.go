package coordinator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveCwdReturnsCanonicalSymlinkTarget(t *testing.T) {
	home := t.TempDir()
	target := filepath.Join(home, "projects", "app")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(home, "app-link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	resolved, err := (&Lifecycle{home: home}).ResolveCwd(link)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != target {
		t.Fatalf("resolved cwd = %q, want %q", resolved, target)
	}
}

func TestResolveShellCwdAllowsHomeDirectory(t *testing.T) {
	home := t.TempDir()
	resolved, err := (&Lifecycle{home: home}).resolveShellCwd(home)
	if err != nil {
		t.Fatalf("resolveShellCwd() error = %v", err)
	}
	want, err := filepath.EvalSymlinks(home)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != want {
		t.Fatalf("resolveShellCwd() = %q, want %q", resolved, want)
	}
}
