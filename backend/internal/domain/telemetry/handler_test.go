package telemetry

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/user"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

func TestIngestTelemetryRequiresDeviceIDAndEvent(t *testing.T) {
	repo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	handler := NewHandler(NewService(repo, uSvc, nil, nil, nil, nil))

	req := httptest.NewRequest(http.MethodPost, "/api/telemetry", bytes.NewBufferString(`{"device_id":"","event":""}`))
	rec := httptest.NewRecorder()

	handler.IngestTelemetry(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestIngestTelemetryReturnsProcessedPayload(t *testing.T) {
	repo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	handler := NewHandler(NewService(repo, uSvc, nil, nil, nil, nil))

	req := httptest.NewRequest(http.MethodPost, "/api/telemetry", bytes.NewBufferString(`{"device_id":"lock_test","event":"access_denied","rfid_uid":"AA:BB:CC:DD"}`))
	rec := httptest.NewRecorder()

	handler.IngestTelemetry(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got models.SensorPayload
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response payload: %v", err)
	}

	if got.Event != "new_card" {
		t.Fatalf("expected event new_card, got %s", got.Event)
	}
	if got.Status != "Pending" {
		t.Fatalf("expected status Pending, got %s", got.Status)
	}
}
