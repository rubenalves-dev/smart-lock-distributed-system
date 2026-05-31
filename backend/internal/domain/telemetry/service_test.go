package telemetry

import (
	"context"
	"fmt"
	"testing"

	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/mfa"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/domain/user"
	"github.com/rubenalves-dev/smart-lock-distributed-system/internal/models"
)

type mockTelemetryRepository struct {
	payloads []models.SensorPayload
}

func (m *mockTelemetryRepository) Save(ctx context.Context, p models.SensorPayload) error {
	m.payloads = append(m.payloads, p)
	return nil
}

func (m *mockTelemetryRepository) GetLatest(ctx context.Context, deviceID string) (*models.SensorPayload, error) {
	var latest *models.SensorPayload
	for i := len(m.payloads) - 1; i >= 0; i-- {
		if m.payloads[i].DeviceID == deviceID {
			latest = &m.payloads[i]
			break
		}
	}
	return latest, nil
}

type fakeUserRepo struct {
	users map[string]*user.User
}

func (f *fakeUserRepo) Create(ctx context.Context, u *user.User) error {
	f.users[u.RfidUID] = u
	return nil
}

func (f *fakeUserRepo) Update(ctx context.Context, u *user.User) error {
	f.users[u.RfidUID] = u
	return nil
}

func (f *fakeUserRepo) FindByUID(ctx context.Context, uid string) (*user.User, error) {
	u, ok := f.users[uid]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (f *fakeUserRepo) ListAll(ctx context.Context) ([]user.User, error) {
	return nil, nil
}

func (f *fakeUserRepo) ProcessMFAAuthentication(ctx context.Context, rfidUID string) (*user.User, error) {
	return f.FindByUID(ctx, rfidUID)
}

type fakeAIService struct {
	predictCalls int
}

func (f *fakeAIService) PredictSeverity(ctx context.Context, event models.SensorPayload) (int32, float32, string, error) {
	f.predictCalls++
	return 1, 0.95, "Keep monitor", nil
}

func (f *fakeAIService) RetrainModel(ctx context.Context, epochs int32, datasetPath string) (*models.RetrainResult, error) {
	return &models.RetrainResult{Success: true, Message: "Retrained"}, nil
}

func (f *fakeAIService) EvaluateModel(ctx context.Context, datasetPath string) (*models.EvaluationResult, error) {
	return &models.EvaluationResult{}, nil
}


func TestTelemetryIngest(t *testing.T) {
	telemetryRepo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	aiSvc := &fakeAIService{}

	svc := NewService(telemetryRepo, uSvc, nil, nil, aiSvc, nil)

	// Ingest access_granted telemetry
	payload := models.SensorPayload{
		DeviceID: "lock_test",
		Event:    "access_granted",
		RfidUID:  "99:88:77:66",
	}

	err := svc.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify telemetry was saved
	if len(telemetryRepo.payloads) != 1 {
		t.Errorf("expected 1 saved telemetry payload, got %d", len(telemetryRepo.payloads))
	}

	// Verify user tag was automatically registered
	registeredUser, err := uSvc.GetUserByUID(context.Background(), "99:88:77:66")
	if err != nil {
		t.Fatalf("failed checking user database: %v", err)
	}
	if registeredUser == nil {
		t.Error("expected user to be automatically registered, got nil")
	} else if registeredUser.RfidUID != "99:88:77:66" {
		t.Errorf("expected registered user UID 99:88:77:66, got %s", registeredUser.RfidUID)
	}
}

func TestTelemetryIngestHeartbeat(t *testing.T) {
	telemetryRepo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	aiSvc := &fakeAIService{}

	svc := NewService(telemetryRepo, uSvc, nil, nil, aiSvc, nil)

	// Ingest heartbeat telemetry
	payload := models.SensorPayload{
		DeviceID: "lock_test",
		Event:    "heartbeat",
	}

	err := svc.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify telemetry was NOT saved synchronously
	if len(telemetryRepo.payloads) != 0 {
		t.Errorf("expected 0 synchronously saved telemetry payloads for heartbeat, got %d", len(telemetryRepo.payloads))
	}

	// Verify AI predict was NOT called
	if aiSvc.predictCalls != 0 {
		t.Errorf("expected 0 AI predict calls for heartbeat, got %d", aiSvc.predictCalls)
	}
}

func TestTelemetryIngestRFIDAccessControl(t *testing.T) {
	telemetryRepo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	aiSvc := &fakeAIService{}

	svc := NewService(telemetryRepo, uSvc, nil, nil, aiSvc, nil)

	// 1. Scanned for the first time
	payload := models.SensorPayload{
		DeviceID: "lock_test",
		Event:    "access_denied", // firmware sent denied
		RfidUID:  "12:34:56:78",
	}

	err := svc.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be saved as new_card and pending
	if len(telemetryRepo.payloads) != 1 {
		t.Fatalf("expected 1 saved payload, got %d", len(telemetryRepo.payloads))
	}
	if telemetryRepo.payloads[0].Event != "new_card" || telemetryRepo.payloads[0].Status != "Pending" {
		t.Errorf("expected Event='new_card' and Status='Pending', got Event='%s' and Status='%s'", 
			telemetryRepo.payloads[0].Event, telemetryRepo.payloads[0].Status)
	}

	// Verify user is in DB as not accepted and not blocked
	u, err := uSvc.GetUserByUID(context.Background(), "12:34:56:78")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u == nil {
		t.Fatalf("expected user to be created")
	}
	if u.IsAccepted || u.IsBlocked {
		t.Errorf("expected is_accepted=false and is_blocked=false, got is_accepted=%v and is_blocked=%v", 
			u.IsAccepted, u.IsBlocked)
	}

	// 2. Scan again when still pending
	telemetryRepo.payloads = nil
	err = svc.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be saved as access_denied and pending
	if len(telemetryRepo.payloads) != 1 {
		t.Fatalf("expected 1 saved payload, got %d", len(telemetryRepo.payloads))
	}
	if telemetryRepo.payloads[0].Event != "access_denied" || telemetryRepo.payloads[0].Status != "Pending" {
		t.Errorf("expected Event='access_denied' and Status='Pending', got Event='%s' and Status='%s'", 
			telemetryRepo.payloads[0].Event, telemetryRepo.payloads[0].Status)
	}

	// 3. Accept user in DB
	isAccepted := true
	_, err = uSvc.UpdateUser(context.Background(), "12:34:56:78", nil, nil, &isAccepted, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Scan again when accepted
	telemetryRepo.payloads = nil
	err = svc.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be saved as access_granted and Success
	if len(telemetryRepo.payloads) != 1 {
		t.Fatalf("expected 1 saved payload, got %d", len(telemetryRepo.payloads))
	}
	if telemetryRepo.payloads[0].Event != "access_granted" || telemetryRepo.payloads[0].Status != "Success" {
		t.Errorf("expected Event='access_granted' and Status='Success', got Event='%s' and Status='%s'", 
			telemetryRepo.payloads[0].Event, telemetryRepo.payloads[0].Status)
	}

	// 4. Block user in DB
	isBlocked := true
	_, err = uSvc.UpdateUser(context.Background(), "12:34:56:78", nil, nil, nil, &isBlocked)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Scan again when blocked
	telemetryRepo.payloads = nil
	err = svc.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be saved as access_denied and Blocked
	if len(telemetryRepo.payloads) != 1 {
		t.Fatalf("expected 1 saved payload, got %d", len(telemetryRepo.payloads))
	}
	if telemetryRepo.payloads[0].Event != "access_denied" || telemetryRepo.payloads[0].Status != "Blocked" {
		t.Errorf("expected Event='access_denied' and Status='Blocked', got Event='%s' and Status='%s'", 
			telemetryRepo.payloads[0].Event, telemetryRepo.payloads[0].Status)
	}
}

type mockMFARepository struct {
	requests []mfa.MFARequest
}

func (m *mockMFARepository) Create(ctx context.Context, req *mfa.MFARequest) error {
	req.ID = len(m.requests) + 1
	m.requests = append(m.requests, *req)
	return nil
}

func (m *mockMFARepository) UpdateStatus(ctx context.Context, id int, status string) error {
	for i := range m.requests {
		if m.requests[i].ID == id {
			m.requests[i].Status = status
			return nil
		}
	}
	return fmt.Errorf("not found")
}

func (m *mockMFARepository) FindByID(ctx context.Context, id int) (*mfa.MFARequest, error) {
	for i := range m.requests {
		if m.requests[i].ID == id {
			return &m.requests[i], nil
		}
	}
	return nil, nil
}

func (m *mockMFARepository) ListAll(ctx context.Context) ([]mfa.MFARequest, error) {
	return m.requests, nil
}

type fakeAIServiceSuspicious struct {
	fakeAIService
}

func (f *fakeAIServiceSuspicious) PredictSeverity(ctx context.Context, event models.SensorPayload) (int32, float32, string, error) {
	return 2, 0.88, "AI predicts SUSPICIOUS access attempt.", nil
}

func TestTelemetryIngestRFIDAIMFA(t *testing.T) {
	telemetryRepo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	aiSvc := &fakeAIServiceSuspicious{}
	mfaRepo := &mockMFARepository{}
	mfaSvc := mfa.NewService(mfaRepo, uSvc, nil, nil)

	svc := NewService(telemetryRepo, uSvc, nil, nil, aiSvc, mfaSvc)

	// Create user in DB and accept it
	isAccepted := true
	_, err := uSvc.UpdateUser(context.Background(), "88:77:66:55", nil, nil, &isAccepted, nil)
	if err != nil {
		t.Fatalf("failed to setup user: %v", err)
	}

	payload := models.SensorPayload{
		DeviceID: "lock_test",
		Event:    "access_request",
		RfidUID:  "88:77:66:55",
	}

	err = svc.Ingest(context.Background(), payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be saved as mfa_pending and status Pending
	if len(telemetryRepo.payloads) != 1 {
		t.Fatalf("expected 1 saved payload, got %d", len(telemetryRepo.payloads))
	}
	if telemetryRepo.payloads[0].Event != "mfa_pending" || telemetryRepo.payloads[0].Status != "Pending" {
		t.Errorf("expected Event='mfa_pending' and Status='Pending', got Event='%s' and Status='%s'", 
			telemetryRepo.payloads[0].Event, telemetryRepo.payloads[0].Status)
	}

	// Verify MFA request was created in repository
	if len(mfaRepo.requests) != 1 {
		t.Fatalf("expected 1 MFA request in mock repository, got %d", len(mfaRepo.requests))
	}
	req := mfaRepo.requests[0]
	if req.RfidUID != "88:77:66:55" || req.Classification != 2 || req.Status != "pending" {
		t.Errorf("unexpected MFA request details: %+v", req)
	}
}

