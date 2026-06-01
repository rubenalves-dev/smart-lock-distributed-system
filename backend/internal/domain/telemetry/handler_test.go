package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/mfa"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/user"
)

func TestIngestTelemetryRejectsMissingRequiredFields(t *testing.T) {
	telemetryRepo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	svc := NewService(telemetryRepo, uSvc, nil, nil, &fakeAIService{}, nil)
	handler := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/telemetry", bytes.NewBufferString(`{"details":"x"}`))
	rec := httptest.NewRecorder()

	handler.IngestTelemetry(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestIngestTelemetryReturnsProcessedTelemetry(t *testing.T) {
	telemetryRepo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	aiSvc := &fakeAIServiceSuspicious{}
	mfaRepo := &mockMFARepository{}
	mfaSvc := mfa.NewService(mfaRepo, uSvc, nil, nil)
	svc := NewService(telemetryRepo, uSvc, nil, nil, aiSvc, mfaSvc)
	handler := NewHandler(svc)

	isAccepted := true
	if _, err := uSvc.UpdateUser(context.Background(), "88:77:66:55", nil, nil, &isAccepted, nil); err != nil {
		t.Fatalf("failed to setup user: %v", err)
	}

	reqBody := `{"device_id":"lock_test","event":"access_request","rfid_uid":"88:77:66:55"}`
	req := httptest.NewRequest(http.MethodPost, "/api/telemetry", bytes.NewBufferString(reqBody))
	rec := httptest.NewRecorder()

	handler.IngestTelemetry(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["status"] != "accepted" {
		t.Fatalf("expected status=accepted, got %v", body["status"])
	}

	telemetryBody, ok := body["telemetry"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected telemetry object in response")
	}

	if telemetryBody["event"] != "mfa_pending" {
		t.Fatalf("expected telemetry.event=mfa_pending, got %v", telemetryBody["event"])
	}
	if telemetryBody["status"] != "Pending" {
		t.Fatalf("expected telemetry.status=Pending, got %v", telemetryBody["status"])
	}
}

