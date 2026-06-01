package models

import "time"

type SensorPayload struct {
	DeviceID string `json:"device_id"`
	Event    string `json:"event"`
	Details  string `json:"details"`

	Status     string  `json:"status"`
	DistanceCm float32 `json:"distance_cm"`
	LightLevel int     `json:"light_level"`

	Fails int    `json:"fails"`
	User  string `json:"user,omitempty"`

	RSSI   int     `json:"rssi"`
	Uptime float32 `json:"uptime"`

	RfidUID string `json:"rfid_uid,omitempty"`

	Timestamp time.Time `json:"timestamp,omitempty"`
	IsOnline  bool      `json:"is_online"`
}


type HealthPoint struct {
	Timestamp string `json:"ts"`
	Status    *int   `json:"status"`
}

type ServiceHealthSeries struct {
	Service string        `json:"service"`
	Points  []HealthPoint `json:"points"`
}

type EvaluationMetrics struct {
	Accuracy       float32 `json:"accuracy"`
	PrecisionMacro float32 `json:"precision_macro"`
	RecallMacro    float32 `json:"recall_macro"`
	F1Macro        float32 `json:"f1_macro"`
}

type BinaryEvaluationMetrics struct {
	Accuracy  float32 `json:"accuracy"`
	Precision float32 `json:"precision"`
	Recall    float32 `json:"recall"`
	F1        float32 `json:"f1"`
}

type EvaluationResult struct {
	ConfusionMatrix [][]int32               `json:"confusion_matrix"`
	Metrics         EvaluationMetrics       `json:"metrics"`
	BinaryMetrics   BinaryEvaluationMetrics `json:"binary_metrics"`
}

type TrainingDiagnostics struct {
	TrainAccuracy        float32 `json:"train_accuracy"`
	ValidationAccuracy   float32 `json:"validation_accuracy"`
	TrainLoss            float32 `json:"train_loss"`
	ValidationLoss       float32 `json:"validation_loss"`
	UnderfittingDetected bool    `json:"underfitting_detected"`
	OverfittingDetected  bool    `json:"overfitting_detected"`
}

type RetrainResult struct {
	Success     bool                 `json:"success"`
	Message     string               `json:"message"`
	Diagnostics *TrainingDiagnostics `json:"diagnostics,omitempty"`
}

