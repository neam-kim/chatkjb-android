package proto

import (
	"encoding/json"
	"testing"

	"github.com/mohamed-essam/herdr-mobile/companion/internal/state"
)

func TestParseClientMsg(t *testing.T) {
	m, err := ParseClient([]byte(`{"t":"send_text","reqId":"r2","paneId":"w6:p1","text":"y"}`))
	if err != nil {
		t.Fatal(err)
	}
	if m.T != "send_text" || m.PaneID != "w6:p1" || m.Text != "y" || m.ReqID != "r2" {
		t.Fatalf("bad parse: %+v", m)
	}
}

func TestPanesSnapshotFrame(t *testing.T) {
	b := PanesSnapshot([]state.Pane{{PaneID: "w6:p1", AgentStatus: "working"}})
	var got map[string]any
	json.Unmarshal(b, &got)
	if got["t"] != "panes" {
		t.Fatalf("want t=panes, got %v", got["t"])
	}
	panes := got["panes"].([]any)
	if len(panes) != 1 {
		t.Fatalf("want 1 pane in snapshot, got %d", len(panes))
	}
}

func TestErrorFrameCarriesReqID(t *testing.T) {
	b := ErrorFrame("r2", "not_found", "pane not found")
	var got map[string]any
	json.Unmarshal(b, &got)
	if got["t"] != "error" || got["reqId"] != "r2" || got["code"] != "not_found" {
		t.Fatalf("bad error frame: %v", got)
	}
}

func TestParseClientTermFields(t *testing.T) {
	m, err := ParseClient([]byte(`{"t":"term_open","reqId":"r9","target":"w6:p1","cols":80,"rows":24}`))
	if err != nil {
		t.Fatal(err)
	}
	if m.T != "term_open" || m.Target != "w6:p1" || m.Cols != 80 || m.Rows != 24 || m.ReqID != "r9" {
		t.Fatalf("bad term_open parse: %+v", m)
	}
	m2, _ := ParseClient([]byte(`{"t":"term_input","termId":"t1","data":"aGk="}`))
	if m2.TermID != "t1" || m2.Data != "aGk=" {
		t.Fatalf("bad term_input parse: %+v", m2)
	}
}

func TestTermFrames(t *testing.T) {
	var got map[string]any

	json.Unmarshal(TermOpened("r9", "t1"), &got)
	if got["t"] != "term_opened" || got["reqId"] != "r9" || got["termId"] != "t1" {
		t.Fatalf("bad term_opened: %v", got)
	}

	json.Unmarshal(TermData("t1", "aGk="), &got)
	if got["t"] != "term_data" || got["termId"] != "t1" || got["data"] != "aGk=" {
		t.Fatalf("bad term_data: %v", got)
	}

	json.Unmarshal(TermExit("t1", 3, "takeover"), &got)
	if got["t"] != "term_exit" || got["termId"] != "t1" || got["code"].(float64) != 3 || got["reason"] != "takeover" {
		t.Fatalf("bad term_exit: %v", got)
	}

	json.Unmarshal(TermError("r9", "t1", "boom"), &got)
	if got["t"] != "term_error" || got["message"] != "boom" {
		t.Fatalf("bad term_error: %v", got)
	}
}

func TestWelcomeAdvertisesProtocol7(t *testing.T) {
	var got map[string]any
	json.Unmarshal(Welcome("0.7.1", 14), &got)
	if got["companionProtocol"].(float64) != 8 {
		t.Fatalf("want companionProtocol 8, got %v", got["companionProtocol"])
	}
}

func TestQServantFrames(t *testing.T) {
	var got map[string]any
	json.Unmarshal(QServantCatalogResult("r1", nil, "m1", "high", "2026-08-31T00:00:00Z"), &got)
	if got["t"] != "qservant_catalog_result" || got["reqId"] != "r1" || got["defaultModel"] != "m1" || got["defaultEffort"] != "high" {
		t.Fatalf("bad catalog result: %v", got)
	}
	if got["quota"] != nil {
		t.Fatalf("top-level quota must not exist: %v", got)
	}

	got = map[string]any{}
	json.Unmarshal(QServantJob("r2", QServantJobPayload{JobID: "j1", State: "running", Transcript: "hello"}), &got)
	if got["t"] != "qservant_job" || got["reqId"] != "r2" {
		t.Fatalf("bad job frame header: %v", got)
	}
	jobObj := got["job"].(map[string]any)
	if jobObj["jobId"] != "j1" || jobObj["state"] != "running" || jobObj["transcript"] != "hello" {
		t.Fatalf("bad nested job obj: %v", jobObj)
	}

	got = map[string]any{}
	json.Unmarshal(QServantJob("", QServantJobPayload{JobID: "j1", State: "completed"}), &got)
	if got["t"] != "qservant_job" || got["reqId"] != nil {
		t.Fatalf("broadcast job frame should not have reqId: %v", got)
	}
	jobObj = got["job"].(map[string]any)
	if jobObj["jobId"] != "j1" || jobObj["state"] != "completed" {
		t.Fatalf("bad nested job obj in broadcast: %v", jobObj)
	}

	got = map[string]any{}
	json.Unmarshal(QServantError("r3", "j1", "not_found", "job not found"), &got)
	if got["t"] != "qservant_error" || got["reqId"] != "r3" || got["jobId"] != "j1" || got["code"] != "not_found" {
		t.Fatalf("bad error frame: %v", got)
	}
}

func TestParseQServantClientMsgs(t *testing.T) {
	m1, err := ParseClient([]byte(`{"t":"qservant_catalog","reqId":"r1"}`))
	if err != nil || m1.T != "qservant_catalog" || m1.ReqID != "r1" {
		t.Fatalf("bad catalog msg: %+v, err: %v", m1, err)
	}

	m2, err := ParseClient([]byte(`{"t":"qservant_submit","reqId":"r2","model":"openai/gpt","effort":"high","audioMime":"audio/mp4","audioBase64":"YWJj"}`))
	if err != nil || m2.T != "qservant_submit" || m2.Model != "openai/gpt" || m2.Effort != "high" || m2.AudioMIME != "audio/mp4" || m2.AudioBase64 != "YWJj" {
		t.Fatalf("bad submit msg: %+v, err: %v", m2, err)
	}

	m2Nested, err := ParseClient([]byte(`{"t":"qservant_submit","reqId":"r2","model":"openai/gpt","audio":{"mimeType":"audio/mp4","data":"YWJj"}}`))
	if err != nil || m2Nested.AudioMIME != "audio/mp4" || m2Nested.AudioBase64 != "YWJj" {
		t.Fatalf("bad nested audio submit msg: %+v", m2Nested)
	}

	m3, err := ParseClient([]byte(`{"t":"qservant_status","reqId":"r3","jobId":"j123"}`))
	if err != nil || m3.T != "qservant_status" || m3.ReqID != "r3" || m3.JobID != "j123" {
		t.Fatalf("bad status msg: %+v", m3)
	}

	m4, err := ParseClient([]byte(`{"t":"qservant_cancel","reqId":"r4","jobId":"j123"}`))
	if err != nil || m4.T != "qservant_cancel" || m4.ReqID != "r4" || m4.JobID != "j123" {
		t.Fatalf("bad cancel msg: %+v", m4)
	}
}
