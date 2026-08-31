package qservant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JobState string

const (
	StateRecorded     JobState = "recorded"
	StateUploading    JobState = "uploading"
	StateTranscribing JobState = "transcribing"
	StateQueued       JobState = "queued"
	StateRunning      JobState = "running"
	StateCompleted    JobState = "completed"
	StateFailed       JobState = "failed"
	StateCancelled    JobState = "cancelled"
)

type JobRequest struct {
	Model     string `json:"model"`
	Effort    string `json:"effort,omitempty"`
	Prompt    string `json:"prompt,omitempty"`
	AudioPath string `json:"audioPath,omitempty"`
}
type RunnerReport struct {
	Request      any  `json:"request"`
	Work         any  `json:"work"`
	Verification any  `json:"verification"`
	Changes      any  `json:"changes"`
	Commit       any  `json:"commit,omitempty"`
	Result       any  `json:"result"`
	Success      bool `json:"success"`
}

func (r RunnerReport) Valid() bool {
	return r.Request != nil && r.Work != nil && r.Verification != nil && r.Changes != nil && r.Result != nil
}

type RunnerResult struct {
	State  string
	Report RunnerReport
}
type RunHandle interface {
	Wait(context.Context) (RunnerResult, error)
	Cancel() error
}
type Runner interface {
	Start(context.Context, JobRequest) (RunHandle, error)
}

var ErrInvalidReport = errors.New("runner report missing required fields")
var ErrRunnerState = errors.New("runner returned non-terminal state")

type Job struct {
	ID        string        `json:"id"`
	Request   JobRequest    `json:"request"`
	State     JobState      `json:"state"`
	Error     string        `json:"error,omitempty"`
	Report    *RunnerReport `json:"report,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}
type JobController struct {
	mu      sync.RWMutex
	dir     string
	runner  Runner
	jobs    map[string]*Job
	handles map[string]RunHandle
	seq     uint64
}

func NewJobController(dir string, r Runner) *JobController {
	if dir == "" {
		dir = os.TempDir()
	}
	_ = os.MkdirAll(dir, 0700)
	c := &JobController{dir: dir, runner: r, jobs: map[string]*Job{}, handles: map[string]RunHandle{}}
	c.load()
	return c
}
func (c *JobController) Submit(ctx context.Context, req JobRequest) (string, error) {
	id := fmt.Sprintf("job-%d", time.Now().UnixNano())
	now := time.Now().UTC()
	j := &Job{ID: id, Request: req, State: StateRecorded, CreatedAt: now, UpdatedAt: now}
	c.mu.Lock()
	c.jobs[id] = j
	c.persistLocked(j)
	c.mu.Unlock()
	// A WebSocket request context ends when the submit handler returns or the
	// phone disconnects. The persisted job must continue independently.
	go c.run(context.Background(), id)
	return id, nil
}
func (c *JobController) run(ctx context.Context, id string) {
	c.mu.RLock()
	if j := c.jobs[id]; j == nil || j.State == StateCancelled {
		c.mu.RUnlock()
		return
	}
	c.mu.RUnlock()
	c.setState(id, StateUploading, "")
	c.setState(id, StateTranscribing, "")
	c.setState(id, StateQueued, "")
	c.mu.RLock()
	j := c.jobs[id]
	req := j.Request
	c.mu.RUnlock()
	if c.runner == nil {
		c.setState(id, StateFailed, "runner unavailable")
		return
	}
	h, e := c.runner.Start(ctx, req)
	if e != nil {
		c.setState(id, StateFailed, e.Error())
		return
	}
	c.mu.Lock()
	c.handles[id] = h
	if j := c.jobs[id]; j != nil && j.State == StateCancelled {
		delete(c.handles, id)
		c.mu.Unlock()
		_ = h.Cancel()
		return
	}
	c.mu.Unlock()
	c.setState(id, StateRunning, "")
	rr, e := h.Wait(ctx)
	c.mu.Lock()
	delete(c.handles, id)
	c.mu.Unlock()
	if e != nil {
		if errors.Is(e, context.Canceled) {
			c.setState(id, StateCancelled, e.Error())
		} else {
			c.setState(id, StateFailed, e.Error())
		}
		return
	}
	if rr.State == "cancelled" {
		c.setState(id, StateCancelled, "")
		return
	}
	if rr.State != "completed" && rr.State != "failed" {
		c.setState(id, StateFailed, ErrRunnerState.Error())
		return
	}
	if !rr.Report.Valid() {
		c.setState(id, StateFailed, ErrInvalidReport.Error())
		return
	}
	c.mu.Lock()
	if current := c.jobs[id]; current != nil && current.State == StateCancelled {
		c.mu.Unlock()
		return
	}
	if rr.State == "completed" && rr.Report.Success {
		c.jobs[id].State = StateCompleted
	} else if rr.State == "failed" {
		c.jobs[id].State = StateFailed
	} else {
		c.jobs[id].State = StateFailed
	}
	c.jobs[id].Report = &rr.Report
	c.jobs[id].UpdatedAt = time.Now().UTC()
	c.persistLocked(c.jobs[id])
	c.mu.Unlock()
}
func (c *JobController) setState(id string, s JobState, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if j := c.jobs[id]; j != nil {
		if (j.State == StateCancelled || j.State == StateCompleted || j.State == StateFailed) && s != StateCancelled && s != StateCompleted && s != StateFailed {
			return
		}
		j.State = s
		j.Error = msg
		j.UpdatedAt = time.Now().UTC()
		c.persistLocked(j)
	}
}
func (c *JobController) Status(id string) (Job, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	j, ok := c.jobs[id]
	if !ok {
		return Job{}, false
	}
	v := *j
	return v, true
}
func (c *JobController) GetStatus(id string) (Job, bool) { return c.Status(id) }
func (c *JobController) Cancel(id string) error {
	c.mu.Lock()
	j, ok := c.jobs[id]
	if !ok {
		c.mu.Unlock()
		return os.ErrNotExist
	}
	if j.State == StateCompleted || j.State == StateFailed || j.State == StateCancelled {
		c.mu.Unlock()
		return nil
	}
	h := c.handles[id]
	j.State = StateCancelled
	j.UpdatedAt = time.Now().UTC()
	c.persistLocked(j)
	c.mu.Unlock()
	if h != nil {
		return h.Cancel()
	}
	if x, ok := c.runner.(interface{ Cancel(string) error }); ok {
		return x.Cancel(id)
	}
	return nil
}
func (c *JobController) CancelJob(id string) error { return c.Cancel(id) }
func (c *JobController) persistLocked(j *Job) {
	b, _ := json.MarshalIndent(j, "", "  ")
	tmp := filepath.Join(c.dir, j.ID+".tmp")
	_ = os.WriteFile(tmp, b, 0600)
	_ = os.Rename(tmp, filepath.Join(c.dir, j.ID+".json"))
}
func (c *JobController) load() {
	ents, _ := os.ReadDir(c.dir)
	for _, e := range ents {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(c.dir, e.Name()))
		if err != nil {
			continue
		}
		var j Job
		if json.Unmarshal(b, &j) == nil && j.ID != "" {
			switch j.State {
			case StateRecorded, StateUploading, StateTranscribing, StateQueued, StateRunning:
				j.State = StateFailed
				j.Error = "companion restarted before job completion"
				j.UpdatedAt = time.Now().UTC()
			}
			c.jobs[j.ID] = &j
			c.persistLocked(&j)
		}
	}
}

type FakeRunner struct {
	Result RunnerResult
	Err    error
	Handle *FakeHandle
}

func (f *FakeRunner) Start(context.Context, JobRequest) (RunHandle, error) {
	if f.Err != nil {
		return nil, f.Err
	}
	if f.Handle == nil {
		f.Handle = &FakeHandle{Result: f.Result}
	}
	return f.Handle, nil
}

type FakeHandle struct {
	Result    RunnerResult
	Cancelled bool
	Err       error
}

func (f *FakeHandle) Wait(ctx context.Context) (RunnerResult, error) {
	if f.Err != nil {
		return RunnerResult{}, f.Err
	}
	if f.Cancelled {
		return RunnerResult{State: "cancelled"}, nil
	}
	return f.Result, nil
}
func (f *FakeHandle) Cancel() error { f.Cancelled = true; return nil }
