package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agro-supply-erp/internal/response"
)

func TestHealthLive(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	handler.Live(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body response.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", body.Status)
	}

	dataMap, ok := body.Data.(map[string]interface{})
	if !ok || dataMap["status"] != "alive" {
		t.Errorf("expected data.status 'alive', got %v", dataMap["status"])
	}
}

func TestHealthLiveMethodNotAllowed(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/health/live", nil)
	w := httptest.NewRecorder()

	handler.Live(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", resp.StatusCode)
	}
}

func TestHealthReadyUnhealthyWhenDBNil(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	handler.Ready(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 when DB is nil, got %d", resp.StatusCode)
	}

	var body response.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Status != "error" {
		t.Errorf("expected status 'error', got '%s'", body.Status)
	}

	if body.Error.Code != "DATABASE_UNAVAILABLE" {
		t.Errorf("expected error code 'DATABASE_UNAVAILABLE', got '%s'", body.Error.Code)
	}
}
