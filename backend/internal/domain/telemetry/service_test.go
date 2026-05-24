package telemetry

import (
	"context"
	"testing"

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

func (f *fakeAIService) RetrainModel(ctx context.Context, epochs int32, datasetPath string) (bool, string, error) {
	return true, "Retrained", nil
}

func TestTelemetryIngest(t *testing.T) {
	telemetryRepo := &mockTelemetryRepository{}
	uRepo := &fakeUserRepo{users: make(map[string]*user.User)}
	uSvc := user.NewService(uRepo)
	aiSvc := &fakeAIService{}

	svc := NewService(telemetryRepo, uSvc, nil, nil, aiSvc)

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

	svc := NewService(telemetryRepo, uSvc, nil, nil, aiSvc)

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

