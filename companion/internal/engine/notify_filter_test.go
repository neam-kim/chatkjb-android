package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/herdr"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/notify"
	"github.com/mohamed-essam/herdr-mobile/companion/internal/state"
)

// engineWithSpaces builds an engine whose store knows the live space labels,
// mirroring the local Herdr layout: General plus the OCA spaces.
func engineWithSpaces(t *testing.T, allowed ...string) *Engine {
	t.Helper()
	e := New(Config{NotifySpaces: allowed})
	e.store.ApplyWorkspaces([]herdr.WorkspaceInfo{
		{WorkspaceID: "w1C", Label: "General"},
		{WorkspaceID: "w1E", Label: "OCA Manager"},
		{WorkspaceID: "w1D", Label: "OCA worker"},
	})
	return e
}

func TestSpaceFilterAllowsOnlyGeneral(t *testing.T) {
	e := engineWithSpaces(t, "General")
	if !e.spaceNotifies("w1C") {
		t.Fatal("General space must notify")
	}
	for _, muted := range []string{"w1E", "w1D"} {
		if e.spaceNotifies(muted) {
			t.Fatalf("OCA space %s must stay silent", muted)
		}
	}
}

// A space Herdr has not reported yet (a freshly spawned OCA space) must not
// notify, so new subagent spaces are silent by default.
func TestSpaceFilterRejectsUnknownWorkspace(t *testing.T) {
	e := engineWithSpaces(t, "General")
	if e.spaceNotifies("w9Z") {
		t.Fatal("unknown workspace must not notify while a filter is configured")
	}
}

func TestSpaceFilterLabelMatchIsCaseInsensitive(t *testing.T) {
	e := engineWithSpaces(t, "general")
	if !e.spaceNotifies("w1C") {
		t.Fatal("label match should ignore case")
	}
}

// An empty filter keeps the previous behaviour of notifying for every space.
func TestEmptyFilterNotifiesEverySpace(t *testing.T) {
	e := engineWithSpaces(t)
	for _, id := range []string{"w1C", "w1E", "w9Z"} {
		if !e.spaceNotifies(id) {
			t.Fatalf("empty filter must notify for %s", id)
		}
	}
}

// "clear" retracts an already-delivered notification, so handleTransition must
// still deliver it for a muted space; only blocked/finished are filtered.
func TestHandleTransitionDeliversClearForMutedSpace(t *testing.T) {
	e, pushes, done := engineWithPushCapture(t, "General")
	defer done()

	e.handleTransition(context.Background(),
		state.Transition{PaneID: "w1E:p1", WorkspaceID: "w1E", From: "blocked", To: "working"})

	select {
	case p := <-pushes:
		if p.Kind != "clear" || p.WorkspaceID != "w1E" {
			t.Fatalf("want clear push for the muted space, got %+v", p)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("clear push for a muted space was dropped")
	}
}

// A blocked transition in an OCA space must raise nothing at all.
func TestHandleTransitionDropsBlockedForMutedSpace(t *testing.T) {
	e, pushes, done := engineWithPushCapture(t, "General")
	defer done()

	e.handleTransition(context.Background(),
		state.Transition{PaneID: "w1D:p1", WorkspaceID: "w1D", From: "working", To: "blocked"})

	select {
	case p := <-pushes:
		t.Fatalf("muted space must not push, got %+v", p)
	case <-time.After(300 * time.Millisecond):
	}
}

// engineWithPushCapture wires the engine's push endpoint to a test server so
// tests observe exactly what the phone would receive.
func engineWithPushCapture(t *testing.T, allowed ...string) (*Engine, <-chan notify.Push, func()) {
	t.Helper()
	e := engineWithSpaces(t, allowed...)
	got := make(chan notify.Push, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p notify.Push
		json.NewDecoder(r.Body).Decode(&p)
		got <- p
	}))
	e.setEndpoint(srv.URL)
	return e, got, srv.Close
}
