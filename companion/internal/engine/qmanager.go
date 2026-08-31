package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/proto"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/qservant"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/wsserver"
)

type QuotaView struct {
	Used      *float64 `json:"used,omitempty"`
	Label     string   `json:"label,omitempty"`
	Remaining *float64 `json:"remaining,omitempty"`
	Limit     *float64 `json:"limit,omitempty"`
	Unit      string   `json:"unit,omitempty"`
}

type CatalogModelView struct {
	ID            string     `json:"id"`
	Label         string     `json:"label"`
	Efforts       []string   `json:"efforts"`
	DefaultEffort string     `json:"defaultEffort,omitempty"`
	Quota         *QuotaView `json:"quota,omitempty"`
}

func transformSingleQuota(q qservant.Quota) *QuotaView {
	if q.Remaining == nil && q.Limit == nil && q.Used == nil {
		return nil
	}
	qv := &QuotaView{
		Remaining: q.Remaining,
		Limit:     q.Limit,
		Unit:      q.Unit,
		Label:     strings.TrimSpace(q.Label),
	}
	if q.Limit != nil && *q.Limit > 0 {
		if q.Used != nil {
			u := *q.Used / *q.Limit
			if u < 0 {
				u = 0
			} else if u > 1 {
				u = 1
			}
			qv.Used = &u
		} else if q.Remaining != nil {
			u := (*q.Limit - *q.Remaining) / *q.Limit
			if u < 0 {
				u = 0
			} else if u > 1 {
				u = 1
			}
			qv.Used = &u
		}
	} else if q.Used != nil && *q.Used >= 0 && *q.Used <= 1 {
		u := *q.Used
		qv.Used = &u
	}
	if qv.Label != "" {
		return qv
	}
	if q.Remaining != nil && q.Limit != nil {
		qv.Label = fmt.Sprintf("%.0f/%.0f %s", *q.Remaining, *q.Limit, q.Unit)
	} else if q.Remaining != nil {
		qv.Label = fmt.Sprintf("%.0f %s remaining", *q.Remaining, q.Unit)
	} else if q.Used != nil && q.Limit != nil {
		qv.Label = fmt.Sprintf("%.0f/%.0f %s", *q.Used, *q.Limit, q.Unit)
	}
	qv.Label = strings.TrimSpace(qv.Label)
	return qv
}

func quotaForModel(quotas []qservant.Quota, model qservant.LiveModel) *QuotaView {
	provider := strings.ToLower(model.Provider)
	wantedFamily := ""
	if provider == "google-antigravity" {
		id := strings.ToLower(model.Namespaced + " " + model.ID)
		switch {
		case strings.Contains(id, "gemini"):
			wantedFamily = "gem"
		case strings.Contains(id, "claude"):
			wantedFamily = "cla"
		}
	}
	for _, q := range quotas {
		if !strings.EqualFold(q.Provider, provider) {
			continue
		}
		if wantedFamily != "" && !strings.EqualFold(q.Family, wantedFamily) {
			continue
		}
		if qv := transformSingleQuota(q); qv != nil {
			return qv
		}
	}
	return nil
}

func TransformQuotas(quotas []qservant.Quota) []QuotaView {
	if len(quotas) == 0 {
		return nil
	}
	out := make([]QuotaView, 0, len(quotas))
	for _, q := range quotas {
		if qv := transformSingleQuota(q); qv != nil {
			out = append(out, *qv)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

type QServantManager struct {
	catalog    *qservant.Catalog
	quota      *qservant.QuotaClient
	controller *qservant.JobController
	stt        qservant.STT
	broadcast  func([]byte)
	mu         sync.Mutex
	watching   map[string]bool
}

func NewQServantManager(cat *qservant.Catalog, quota *qservant.QuotaClient, ctrl *qservant.JobController, stt qservant.STT, broadcast func([]byte)) *QServantManager {
	if broadcast == nil {
		broadcast = func([]byte) {}
	}
	return &QServantManager{
		catalog:    cat,
		quota:      quota,
		controller: ctrl,
		stt:        stt,
		broadcast:  broadcast,
		watching:   map[string]bool{},
	}
}

func (m *QServantManager) Catalog(ctx context.Context) (wsserver.QServantCatalogInfo, error) {
	if m.catalog == nil {
		return wsserver.QServantCatalogInfo{}, errors.New("catalog unavailable")
	}
	snap := m.catalog.Refresh(ctx)
	defaultEffort := snap.DefaultEffort
	var qList []qservant.Quota
	if m.quota != nil {
		qList = m.quota.Fetch(ctx)
	}

	mappedModels := make([]CatalogModelView, 0, len(snap.Models))
	for _, mdl := range snap.Models {
		efforts := mdl.ReasoningEfforts
		if efforts == nil {
			efforts = []string{}
		}
		label := mdl.ID
		if label == "" {
			label = mdl.Namespaced
		}
		id := mdl.Namespaced
		if id == "" {
			id = mdl.ID
		}
		qv := quotaForModel(qList, mdl)
		mappedModels = append(mappedModels, CatalogModelView{
			ID:            id,
			Label:         label,
			Efforts:       efforts,
			DefaultEffort: mdl.DefaultReasoningEffort,
			Quota:         qv,
		})
	}

	return wsserver.QServantCatalogInfo{
		Models:        mappedModels,
		DefaultModel:  snap.DefaultModel,
		DefaultEffort: defaultEffort,
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (m *QServantManager) Submit(ctx context.Context, model, effort, audioMime, audioBase64 string) (proto.QServantJobPayload, error) {
	if m.catalog == nil || m.controller == nil {
		return proto.QServantJobPayload{}, errors.New("qservant service unavailable")
	}
	if len(m.catalog.Snapshot().Models) == 0 {
		m.catalog.Refresh(ctx)
	}
	if err := m.catalog.Validate(model, effort); err != nil {
		return proto.QServantJobPayload{}, err
	}
	audioJSON, _ := json.Marshal(map[string]any{
		"v":        1,
		"mimeType": audioMime,
		"data":     audioBase64,
	})
	audioIn, err := qservant.DecodeAudioJSON(audioJSON)
	if err != nil {
		return proto.QServantJobPayload{}, err
	}
	transcript, err := qservant.TranscribeAudio(ctx, audioIn, m.stt)
	if err != nil {
		return proto.QServantJobPayload{}, err
	}
	jobID, err := m.controller.Submit(ctx, qservant.JobRequest{
		Model:  model,
		Effort: effort,
		Prompt: transcript,
	})
	if err != nil {
		return proto.QServantJobPayload{}, err
	}
	payload := proto.QServantJobPayload{
		JobID:      jobID,
		State:      string(qservant.StateRecorded),
		Transcript: transcript,
	}
	m.watchJob(jobID, transcript)
	return payload, nil
}

func (m *QServantManager) Status(ctx context.Context, jobID string) (proto.QServantJobPayload, bool) {
	if m.controller == nil {
		return proto.QServantJobPayload{}, false
	}
	job, ok := m.controller.Status(jobID)
	if !ok {
		return proto.QServantJobPayload{}, false
	}
	payload := proto.QServantJobPayload{
		JobID:      job.ID,
		State:      string(job.State),
		Transcript: job.Request.Prompt,
		Error:      job.Error,
	}
	if job.Report != nil {
		payload.Report = job.Report
	}
	return payload, true
}

func (m *QServantManager) Cancel(ctx context.Context, jobID string) (proto.QServantJobPayload, error) {
	if m.controller == nil {
		return proto.QServantJobPayload{}, errors.New("controller unavailable")
	}
	if err := m.controller.Cancel(jobID); err != nil {
		return proto.QServantJobPayload{}, err
	}
	job, ok := m.controller.Status(jobID)
	if !ok {
		return proto.QServantJobPayload{
			JobID: jobID,
			State: string(qservant.StateCancelled),
		}, nil
	}
	payload := proto.QServantJobPayload{
		JobID:      job.ID,
		State:      string(job.State),
		Transcript: job.Request.Prompt,
		Error:      job.Error,
	}
	if job.Report != nil {
		payload.Report = job.Report
	}
	return payload, nil
}

func (m *QServantManager) watchJob(jobID, transcript string) {
	m.mu.Lock()
	if m.watching[jobID] {
		m.mu.Unlock()
		return
	}
	m.watching[jobID] = true
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.watching, jobID)
			m.mu.Unlock()
		}()

		var lastState qservant.JobState
		var lastErr string
		ticker := time.NewTicker(25 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			job, ok := m.controller.Status(jobID)
			if !ok {
				return
			}
			if job.State != lastState || job.Error != lastErr {
				lastState = job.State
				lastErr = job.Error
				p := proto.QServantJobPayload{
					JobID:      job.ID,
					State:      string(job.State),
					Transcript: transcript,
					Error:      job.Error,
				}
				if job.Report != nil {
					p.Report = job.Report
				}
				m.broadcast(proto.QServantJob("", p))
			}
			if job.State == qservant.StateCompleted || job.State == qservant.StateFailed || job.State == qservant.StateCancelled {
				return
			}
		}
	}()
}
