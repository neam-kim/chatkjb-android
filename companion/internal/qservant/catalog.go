package qservant

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var effortOrder = map[string]int{"low": 0, "medium": 1, "high": 2, "xhigh": 3, "max": 4, "ultra": 5}
var ErrModelNotFound = errors.New("model is not in catalog")
var ErrEffortNotFound = errors.New("reasoning effort is not available for model")

type CommandRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}
type ExecCommandRunner struct{}

func (ExecCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).Output()
}

type LiveModel struct {
	Provider               string   `json:"provider"`
	ID                     string   `json:"id"`
	Namespaced             string   `json:"namespaced"`
	ReasoningEfforts       []string `json:"reasoningEfforts,omitempty"`
	DefaultReasoningEffort string   `json:"defaultReasoningEffort,omitempty"`
	Disabled               bool     `json:"disabled,omitempty"`
}
type Model = LiveModel
type CatalogSnapshot struct {
	Models        []LiveModel `json:"models"`
	DefaultModel  string      `json:"defaultModel"`
	DefaultEffort string      `json:"defaultEffort,omitempty"`
	EffortCap     string      `json:"effortCap,omitempty"`
	Stale         bool        `json:"stale,omitempty"`
}

type Catalog struct {
	mu         sync.RWMutex
	runner     CommandRunner
	configPath string
	codexPath  string
	staticPath string
	fallback   []LiveModel
	current    CatalogSnapshot
}

func NewCatalog(r CommandRunner, configPath string, fallback []LiveModel) *Catalog {
	if r == nil {
		r = ExecCommandRunner{}
	}
	if configPath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configPath = filepath.Join(home, ".opencodex", "config.json")
		}
	}
	home, _ := os.UserHomeDir()
	return &Catalog{
		runner: r, configPath: configPath, fallback: fallback,
		codexPath:  filepath.Join(home, ".codex", "config.toml"),
		staticPath: filepath.Join(home, ".codex", "opencodex-catalog.json"),
	}
}
func (c *Catalog) Refresh(ctx context.Context) CatalogSnapshot {
	cap, disabled := readOCXConfig(c.configPath)
	b, err := c.runner.Run(ctx, ocxBinary(), "models", "live", "--json")
	if err != nil {
		return c.setFallback(cap, disabled)
	}
	models := normalizeLive(b)
	filtered := make([]LiveModel, 0, len(models))
	for _, m := range models {
		if m.Namespaced == "" || m.Disabled || disabled[m.Namespaced] || disabled[m.ID] || strings.HasPrefix(m.Namespaced, "oca-manager/") {
			continue
		}
		m.ReasoningEfforts = capEfforts(m.ReasoningEfforts, cap)
		if m.DefaultReasoningEffort != "" && !containsString(m.ReasoningEfforts, m.DefaultReasoningEffort) {
			m.DefaultReasoningEffort = ""
		}
		filtered = append(filtered, m)
	}
	if len(filtered) == 0 {
		return c.setFallback(cap, disabled)
	}
	def, effort := resolveDefaults(filtered, c.codexPath)
	s := CatalogSnapshot{Models: filtered, DefaultModel: def, DefaultEffort: effort, EffortCap: cap}
	c.mu.Lock()
	c.current = s
	c.mu.Unlock()
	return s
}

func (c *Catalog) setFallback(cap string, disabled map[string]bool) CatalogSnapshot {
	models := cloneModels(c.fallback)
	if len(models) == 0 {
		models = loadStaticCatalog(c.staticPath)
	}
	filtered := make([]LiveModel, 0, len(models))
	for _, m := range models {
		if m.Namespaced == "" || m.Disabled || disabled[m.Namespaced] || disabled[m.ID] || strings.HasPrefix(m.Namespaced, "oca-manager/") {
			continue
		}
		m.ReasoningEfforts = capEfforts(m.ReasoningEfforts, cap)
		if m.DefaultReasoningEffort != "" && !containsString(m.ReasoningEfforts, m.DefaultReasoningEffort) {
			m.DefaultReasoningEffort = ""
		}
		filtered = append(filtered, m)
	}
	def, effort := resolveDefaults(filtered, c.codexPath)
	s := CatalogSnapshot{Models: filtered, DefaultModel: def, DefaultEffort: effort, EffortCap: cap, Stale: true}
	c.mu.Lock()
	c.current = s
	c.mu.Unlock()
	return s
}
func (c *Catalog) Snapshot() CatalogSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	s := c.current
	s.Models = cloneModels(s.Models)
	return s
}
func (c *Catalog) Validate(model, effort string) error {
	s := c.Snapshot()
	for _, m := range s.Models {
		if m.Namespaced == model {
			if effort == "" {
				return nil
			}
			for _, e := range m.ReasoningEfforts {
				if e == effort {
					return nil
				}
			}
			return ErrEffortNotFound
		}
	}
	return ErrModelNotFound
}
func normalizeLive(b []byte) []LiveModel {
	var raw json.RawMessage
	if json.Unmarshal(b, &raw) != nil {
		return nil
	}
	var arr []struct {
		Provider, ID, Namespaced string
		ReasoningEfforts         []string `json:"reasoningEfforts"`
		DefaultReasoningEffort   string   `json:"defaultReasoningEffort"`
		Disabled                 bool     `json:"disabled"`
	}
	if len(raw) > 0 && raw[0] == '[' {
		_ = json.Unmarshal(raw, &arr)
	} else {
		var wrap struct {
			Models []json.RawMessage `json:"models"`
		}
		if json.Unmarshal(raw, &wrap) == nil {
			for _, x := range wrap.Models {
				var m struct {
					Provider, ID, Namespaced string
					ReasoningEfforts         []string `json:"reasoningEfforts"`
					DefaultReasoningEffort   string   `json:"defaultReasoningEffort"`
					Disabled                 bool     `json:"disabled"`
				}
				if json.Unmarshal(x, &m) == nil {
					arr = append(arr, m)
				}
			}
		}
	}
	out := make([]LiveModel, 0, len(arr))
	for _, m := range arr {
		out = append(out, LiveModel{Provider: m.Provider, ID: m.ID, Namespaced: m.Namespaced, ReasoningEfforts: m.ReasoningEfforts, DefaultReasoningEffort: m.DefaultReasoningEffort, Disabled: m.Disabled})
	}
	return out
}
func NormalizeLiveModels(b []byte) []LiveModel { return normalizeLive(b) }
func capEfforts(es []string, cap string) []string {
	if len(es) == 0 {
		return nil
	}
	max, ok := effortOrder[strings.ToLower(cap)]
	out := []string{}
	for _, e := range es {
		if n, yes := effortOrder[strings.ToLower(e)]; yes && (!ok || n <= max) {
			out = append(out, e)
		}
	}
	return out
}
func readOCXConfig(path string) (string, map[string]bool) {
	b, e := os.ReadFile(path)
	if e != nil {
		return "", nil
	}
	var v struct {
		EffortCap      string   `json:"effortCap"`
		DisabledModels []string `json:"disabledModels"`
	}
	if json.Unmarshal(b, &v) != nil {
		return "", nil
	}
	disabled := make(map[string]bool, len(v.DisabledModels))
	for _, id := range v.DisabledModels {
		disabled[id] = true
	}
	return v.EffortCap, disabled
}

func resolveDefaults(models []LiveModel, codexPath string) (string, string) {
	model, effort := readCodexDefaults(codexPath)
	if !containsModel(models, model) && model != "" {
		for _, m := range models {
			if m.ID == model || strings.HasSuffix(m.Namespaced, "/"+model) {
				model = m.Namespaced
				break
			}
		}
	}
	if !containsModel(models, model) {
		model = deterministicDefault(models)
	}
	for _, m := range models {
		if m.Namespaced != model {
			continue
		}
		if containsString(m.ReasoningEfforts, effort) {
			return model, effort
		}
		if containsString(m.ReasoningEfforts, m.DefaultReasoningEffort) {
			return model, m.DefaultReasoningEffort
		}
		if len(m.ReasoningEfforts) > 0 {
			return model, m.ReasoningEfforts[0]
		}
	}
	return model, ""
}

func readCodexDefaults(path string) (string, string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()
	var model, effort string
	inSection := false
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inSection = true
			continue
		}
		if inSection {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		value, err := strconv.Unquote(strings.TrimSpace(strings.SplitN(parts[1], "#", 2)[0]))
		if err != nil {
			continue
		}
		switch strings.TrimSpace(parts[0]) {
		case "model":
			model = value
		case "model_reasoning_effort":
			effort = value
		}
	}
	return model, effort
}

func loadStaticCatalog(path string) []LiveModel {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var root struct {
		Models []struct {
			Slug          string `json:"slug"`
			DisplayName   string `json:"display_name"`
			DefaultEffort string `json:"default_reasoning_level"`
			Levels        []struct {
				Effort string `json:"effort"`
			} `json:"supported_reasoning_levels"`
		} `json:"models"`
	}
	if json.Unmarshal(b, &root) != nil {
		return nil
	}
	out := make([]LiveModel, 0, len(root.Models))
	for _, item := range root.Models {
		provider := "openai"
		if before, _, ok := strings.Cut(item.Slug, "/"); ok {
			provider = before
		}
		efforts := make([]string, 0, len(item.Levels))
		for _, level := range item.Levels {
			efforts = append(efforts, level.Effort)
		}
		out = append(out, LiveModel{Provider: provider, ID: item.Slug, Namespaced: item.Slug, ReasoningEfforts: efforts, DefaultReasoningEffort: item.DefaultEffort})
	}
	return out
}

func ocxBinary() string {
	if v := strings.TrimSpace(os.Getenv("OCX_BIN")); v != "" {
		return v
	}
	if p, err := exec.LookPath("ocx"); err == nil {
		return p
	}
	for _, p := range []string{"/opt/homebrew/bin/ocx", "/usr/local/bin/ocx"} {
		if st, err := os.Stat(p); err == nil && st.Mode()&0111 != 0 {
			return p
		}
	}
	return "ocx"
}
func deterministicDefault(ms []LiveModel) string {
	if len(ms) == 0 {
		return ""
	}
	cp := cloneModels(ms)
	sort.Slice(cp, func(i, j int) bool { return cp[i].Namespaced < cp[j].Namespaced })
	return cp[0].Namespaced
}
func containsModel(ms []LiveModel, n string) bool {
	for _, m := range ms {
		if m.Namespaced == n {
			return true
		}
	}
	return false
}
func containsString(xs []string, n string) bool {
	for _, x := range xs {
		if x == n {
			return true
		}
	}
	return false
}
func cloneModels(ms []LiveModel) []LiveModel {
	o := make([]LiveModel, len(ms))
	copy(o, ms)
	for i := range o {
		o[i].ReasoningEfforts = append([]string(nil), o[i].ReasoningEfforts...)
	}
	return o
}
