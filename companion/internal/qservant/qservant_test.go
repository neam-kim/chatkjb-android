package qservant

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeCmd struct {
	b     []byte
	err   error
	calls int
}

func (f *fakeCmd) Run(context.Context, string, ...string) ([]byte, error) {
	f.calls++
	return f.b, f.err
}

func TestCatalogNamespacedCapAndValidation(t *testing.T) {
	d := t.TempDir()
	if err := os.WriteFile(filepath.Join(d, "config.json"), []byte(`{"effortCap":"high","defaultModel":"openai/gpt"}`), 0600); err != nil {
		t.Fatal(err)
	}
	r := &fakeCmd{b: []byte(`[{"provider":"openai","id":"gpt-5.6-sol","namespaced":"gpt-5.6-sol","reasoningEfforts":["low","high","ultra"],"defaultReasoningEffort":"high"},{"provider":"x","id":"m","namespaced":"oca-manager/m","reasoningEfforts":["low"]}]`)}
	c := NewCatalog(r, filepath.Join(d, "config.json"), nil)
	s := c.Refresh(context.Background())
	if len(s.Models) != 1 || s.Models[0].Namespaced != "gpt-5.6-sol" {
		t.Fatalf("models=%+v", s.Models)
	}
	if len(s.Models[0].ReasoningEfforts) != 2 {
		t.Fatalf("efforts=%v", s.Models[0].ReasoningEfforts)
	}
	if err := c.Validate("gpt-5.6-sol", "high"); err != nil {
		t.Fatal(err)
	}
	if err := c.Validate("gpt-5.6-sol", "ultra"); err == nil {
		t.Fatal("expected effort rejection")
	}
}

func TestCatalogExcludesDisabledAndUsesCodexDefaults(t *testing.T) {
	d := t.TempDir()
	configPath := filepath.Join(d, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"effortCap":"high","disabledModels":["xai/off"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	codexPath := filepath.Join(d, "config.toml")
	if err := os.WriteFile(codexPath, []byte("model = \"gpt-5.6-sol\"\nmodel_reasoning_effort = \"high\"\n[projects]\nmodel = \"ignored\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	r := &fakeCmd{b: []byte(`[
		{"provider":"openai","id":"old","namespaced":"old","disabled":true,"reasoningEfforts":["low"]},
		{"provider":"xai","id":"off","namespaced":"xai/off","reasoningEfforts":["low"]},
		{"provider":"openai","id":"gpt-5.6-sol","namespaced":"gpt-5.6-sol","reasoningEfforts":["low","high","max"],"defaultReasoningEffort":"low"}
	]`)}
	c := NewCatalog(r, configPath, nil)
	c.codexPath = codexPath
	s := c.Refresh(context.Background())
	if len(s.Models) != 1 || s.DefaultModel != "gpt-5.6-sol" || s.DefaultEffort != "high" {
		t.Fatalf("snapshot=%+v", s)
	}
	if got := s.Models[0].ReasoningEfforts; len(got) != 2 || got[1] != "high" {
		t.Fatalf("capped efforts=%v", got)
	}
}

func TestQuotaNormalizesObservedOCXSchema(t *testing.T) {
	got := NormalizeQuota([]byte(`{"reports":[
		{"provider":"openai","quota":{"fiveHourPercent":45.5,"weeklyPercent":25}},
		{"provider":"xai","quota":{"weeklyPercent":29}},
		{"provider":"google-antigravity","quota":{"customWindows":[{"label":"Gem","percent":43.7},{"label":"Cla","percent":77.5}]}},
		{"provider":"bad","quota":{"weeklyPercent":101}}
	]}`))
	if len(got) != 4 {
		t.Fatalf("quota=%+v", got)
	}
	if got[0].Used == nil || *got[0].Used != .455 || got[0].Label != "45.5% used (5h)" {
		t.Fatalf("openai=%+v", got[0])
	}
	if got[2].Family != "gem" || got[3].Family != "cla" {
		t.Fatalf("families=%+v", got[2:])
	}
}
func TestCatalogFallbackAndQuotaNull(t *testing.T) {
	fb := []LiveModel{{Namespaced: "z"}}
	c := NewCatalog(&fakeCmd{err: os.ErrNotExist}, "/missing", fb)
	s := c.Refresh(context.Background())
	if !s.Stale || s.DefaultModel != "z" {
		t.Fatalf("fallback=%+v", s)
	}
	q := NewQuotaClient(&fakeCmd{b: []byte(`{"provider":"x","remaining":null,"limit":"bad"}`)}, time.Millisecond)
	if got := q.Fetch(context.Background()); len(got) != 0 {
		t.Fatalf("quota=%v", got)
	}
}
func TestAudioCleanup(t *testing.T) {
	raw := base64.StdEncoding.EncodeToString([]byte("aac"))
	in, e := DecodeAudioJSON([]byte(`{"v":1,"mimeType":"audio/mp4","data":"` + raw + `"}`))
	if e != nil {
		t.Fatal(e)
	}
	f, e := MaterializeAudio(in)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = os.Stat(f.Path); e != nil {
		t.Fatal(e)
	}
	p := f.Path
	f.Cleanup()
	if _, e = os.Stat(p); !os.IsNotExist(e) {
		t.Fatal("temp file retained")
	}
	fake := &FakeSTT{Text: "안녕하세요"}
	if text, e := TranscribeAudio(context.Background(), in, fake); e != nil || text != "안녕하세요" {
		t.Fatalf("transcribe=%q err=%v", text, e)
	}
	if _, e := os.Stat(fake.LastPath); !os.IsNotExist(e) {
		t.Fatal("transcription temp retained")
	}
}
func TestJobLifecyclePersistenceAndReport(t *testing.T) {
	report := RunnerReport{Request: map[string]any{}, Work: map[string]any{}, Verification: map[string]any{}, Changes: map[string]any{}, Result: map[string]any{}, Success: true}
	r := &FakeRunner{Result: RunnerResult{State: "completed", Report: report}}
	c := NewJobController(t.TempDir(), r)
	id, _ := c.Submit(context.Background(), JobRequest{Model: "m"})
	for i := 0; i < 100; i++ {
		j, _ := c.Status(id)
		if j.State == StateCompleted {
			if j.Report == nil {
				t.Fatal("missing report")
			}
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("job did not complete")
}

func TestJobOutlivesSubmitContext(t *testing.T) {
	report := RunnerReport{Request: "r", Work: "w", Verification: "v", Changes: []any{}, Result: "ok", Success: true}
	c := NewJobController(t.TempDir(), &FakeRunner{Result: RunnerResult{State: "completed", Report: report}})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	id, err := c.Submit(ctx, JobRequest{Model: "m"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 100; i++ {
		if job, _ := c.Status(id); job.State == StateCompleted {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("job was cancelled with submit context")
}

func TestRestartMarksNonterminalJobFailed(t *testing.T) {
	d := t.TempDir()
	job := Job{ID: "job-running", State: StateRunning, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	b, _ := json.Marshal(job)
	if err := os.WriteFile(filepath.Join(d, job.ID+".json"), b, 0600); err != nil {
		t.Fatal(err)
	}
	c := NewJobController(d, nil)
	loaded, ok := c.Status(job.ID)
	if !ok || loaded.State != StateFailed || loaded.Error == "" {
		t.Fatalf("loaded=%+v ok=%v", loaded, ok)
	}
}
