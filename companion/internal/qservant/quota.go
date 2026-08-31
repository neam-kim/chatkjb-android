package qservant

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

type Quota struct {
	Provider  string   `json:"provider"`
	Family    string   `json:"family,omitempty"`
	Label     string   `json:"label,omitempty"`
	Remaining *float64 `json:"remaining,omitempty"`
	Limit     *float64 `json:"limit,omitempty"`
	Used      *float64 `json:"used,omitempty"`
	Unit      string   `json:"unit,omitempty"`
}

type QuotaRunner interface {
	Run(context.Context, string, ...string) ([]byte, error)
}

type QuotaClient struct {
	runner QuotaRunner
	ttl    time.Duration
	mu     sync.Mutex
	at     time.Time
	value  []Quota
}

func NewQuotaClient(r QuotaRunner, ttl time.Duration) *QuotaClient {
	if ttl <= 0 {
		ttl = 15 * time.Second
	}
	return &QuotaClient{runner: r, ttl: ttl}
}

func (q *QuotaClient) Fetch(ctx context.Context) []Quota {
	q.mu.Lock()
	if time.Since(q.at) < q.ttl {
		v := cloneQuota(q.value)
		q.mu.Unlock()
		return v
	}
	q.mu.Unlock()
	if q.runner == nil {
		return nil
	}
	b, err := q.runner.Run(ctx, ocxBinary(), "provider", "quota", "--json")
	if err != nil {
		return nil
	}
	v := normalizeQuota(b)
	q.mu.Lock()
	q.at = time.Now()
	q.value = v
	q.mu.Unlock()
	return cloneQuota(v)
}

// normalizeQuota follows the live `ocx provider quota --json` reports schema.
// Percent values are normalized to the [0,1] range consumed by Android. A
// missing or malformed provider window deliberately produces no quota object.
func normalizeQuota(b []byte) []Quota {
	var root struct {
		Reports []struct {
			Provider string `json:"provider"`
			Quota    struct {
				FiveHourPercent *float64 `json:"fiveHourPercent"`
				WeeklyPercent   *float64 `json:"weeklyPercent"`
				CustomWindows   []struct {
					Label   string   `json:"label"`
					Percent *float64 `json:"percent"`
				} `json:"customWindows"`
			} `json:"quota"`
		} `json:"reports"`
	}
	if json.Unmarshal(b, &root) != nil {
		return nil
	}
	out := make([]Quota, 0, len(root.Reports))
	for _, report := range root.Reports {
		provider := strings.TrimSpace(report.Provider)
		if provider == "" {
			continue
		}
		if len(report.Quota.CustomWindows) > 0 {
			for _, window := range report.Quota.CustomWindows {
				if used, ok := normalizedPercent(window.Percent); ok {
					family := strings.ToLower(strings.TrimSpace(window.Label))
					out = append(out, Quota{
						Provider: provider,
						Family:   family,
						Label:    fmt.Sprintf("%.1f%% used (%s)", used*100, strings.TrimSpace(window.Label)),
						Used:     floatPtr(used),
						Unit:     "percent",
					})
				}
			}
			continue
		}
		percent, window := report.Quota.FiveHourPercent, "5h"
		if percent == nil {
			percent, window = report.Quota.WeeklyPercent, "weekly"
		}
		if used, ok := normalizedPercent(percent); ok {
			out = append(out, Quota{
				Provider: provider,
				Label:    fmt.Sprintf("%.1f%% used (%s)", used*100, window),
				Used:     floatPtr(used),
				Unit:     "percent",
			})
		}
	}
	return out
}

func NormalizeQuota(b []byte) []Quota { return normalizeQuota(b) }

func normalizedPercent(v *float64) (float64, bool) {
	if v == nil || math.IsNaN(*v) || math.IsInf(*v, 0) || *v < 0 || *v > 100 {
		return 0, false
	}
	return *v / 100, true
}

func floatPtr(v float64) *float64 { return &v }

func cloneQuota(v []Quota) []Quota {
	o := make([]Quota, len(v))
	for i, q := range v {
		o[i] = q
		if q.Remaining != nil {
			x := *q.Remaining
			o[i].Remaining = &x
		}
		if q.Limit != nil {
			x := *q.Limit
			o[i].Limit = &x
		}
		if q.Used != nil {
			x := *q.Used
			o[i].Used = &x
		}
	}
	return o
}

func quotaProvider(v []Quota, p string) *Quota {
	for i := range v {
		if strings.EqualFold(v[i].Provider, p) {
			return &v[i]
		}
	}
	return nil
}
