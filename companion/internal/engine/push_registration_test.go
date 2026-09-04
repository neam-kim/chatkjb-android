package engine

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestPushRegistrationRequiresTokenAndPersistsEndpoint(t *testing.T) {
	dir := t.TempDir()
	endpointPath := filepath.Join(dir, "push-endpoint")
	tokenPath := filepath.Join(dir, "push-registration-token")
	e := New(Config{
		PushEndpointPath:          endpointPath,
		PushRegistrationTokenPath: tokenPath,
		PushEndpointOrigin:        "http://100.94.190.97:2586",
	})

	bad := httptest.NewRequest(http.MethodPost, "/notify/register",
		bytes.NewBufferString(`{"endpoint":"http://100.94.190.97:2586/up-new"}`))
	bad.Header.Set("Authorization", "Bearer wrong")
	badResult := httptest.NewRecorder()
	e.handlePushRegistration(badResult, bad)
	if badResult.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token status = %d, want 401", badResult.Code)
	}

	good := httptest.NewRequest(http.MethodPost, "/notify/register",
		bytes.NewBufferString(`{"endpoint":"http://100.94.190.97:2586/up-new"}`))
	good.Header.Set("Authorization", "Bearer "+e.pushRegistrationToken)
	goodResult := httptest.NewRecorder()
	e.handlePushRegistration(goodResult, good)
	if goodResult.Code != http.StatusNoContent {
		t.Fatalf("valid registration status = %d, want 204", goodResult.Code)
	}
	if got, err := loadPushEndpoint(endpointPath); err != nil || got != "http://100.94.190.97:2586/up-new" {
		t.Fatalf("persisted endpoint = %q, err=%v", got, err)
	}
	if info, err := os.Stat(tokenPath); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("token state mode is not 0600: info=%v err=%v", info, err)
	}
}

func TestPushRegistrationRejectsDifferentOrigin(t *testing.T) {
	e := New(Config{
		PushRegistrationTokenPath: filepath.Join(t.TempDir(), "token"),
		PushEndpointOrigin:        "http://100.94.190.97:2586",
	})
	req := httptest.NewRequest(http.MethodPost, "/notify/register",
		bytes.NewBufferString(`{"endpoint":"https://attacker.example/topic"}`))
	req.Header.Set("Authorization", "Bearer "+e.pushRegistrationToken)
	result := httptest.NewRecorder()
	e.handlePushRegistration(result, req)
	if result.Code != http.StatusBadRequest {
		t.Fatalf("foreign origin status = %d, want 400", result.Code)
	}
}
