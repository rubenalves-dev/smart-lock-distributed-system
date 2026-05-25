import csv
import json
import os
import threading
import time
from concurrent import futures

import gen.lock_pb2 as smartlock
import gen.lock_pb2_grpc as smartlock_grpc
import grpc
import model as model_lib
import numpy as np
import pika
import tensorflow as tf

# Paths
MODEL_PATH = "severity_model.keras"
DATASET_PATH = "data/sensor_events.csv"

class AIService(smartlock_grpc.AIServiceServicer):
    def __init__(self):
        self.model_lock = threading.Lock()
        self.dataset_lock = threading.Lock()

        # Ensure model exists on startup
        self.ensure_model_exists()

        print(f"Loading AI model from {MODEL_PATH}...")
        self.model = tf.keras.models.load_model(MODEL_PATH)
        print("AI model loaded successfully.")

    def ensure_model_exists(self):
        """Creates and trains a baseline model if it doesn't already exist."""
        if not os.path.exists(MODEL_PATH):
            print(f"Model file '{MODEL_PATH}' not found. Training a baseline model...")
            if not os.path.exists(DATASET_PATH):
                print(f"Baseline dataset '{DATASET_PATH}' not found. Generating synthetic data...")
                model_lib.generate_synthetic_data(DATASET_PATH, num_samples=1000)

            # Load baseline dataset
            X, y = model_lib.load_dataset(DATASET_PATH)

            # Create and train
            model = model_lib.create_model()
            print("Training baseline model...")
            model_lib.train_model(model, X, y, epochs=10)

            # Save
            model.save(MODEL_PATH)
            print(f"Baseline model trained and saved to {MODEL_PATH}.")

    def handle_async_event(self, event_data):
        """Asynchronously processes sensor events from RabbitMQ and adds them to the training dataset."""
        fails = int(event_data.get("fails", 0))
        distance = float(event_data.get("distance_cm", 150.0))

        # Determine if access was denied
        event = event_data.get("event", "").lower()
        status = event_data.get("status", "").lower()
        details = event_data.get("details", "").lower()

        is_denied = 0.0
        if "denied" in event or "fail" in event or "denied" in details or status == "failed":
            is_denied = 1.0

        # Heuristic rules for labelling new training data
        if fails >= 3:
            severity = 3  # Critical
        elif fails == 2:
            severity = 2  # Suspicious
        elif fails == 1:
            severity = 1  # Irregular
        elif distance < 15.0 and is_denied == 1.0:
            severity = 2  # Suspicious (close physical attempt)
        else:
            severity = 0  # OK

        # Thread-safe write to CSV
        with self.dataset_lock:
            os.makedirs(os.path.dirname(DATASET_PATH), exist_ok=True)
            file_exists = os.path.exists(DATASET_PATH)
            with open(DATASET_PATH, mode='a', newline='') as f:
                writer = csv.writer(f)
                if not file_exists:
                    writer.writerow(['fails', 'distance_cm', 'is_denied', 'severity'])
                writer.writerow([fails, distance, is_denied, severity])
        print(f"Async recorded event to dataset: fails={fails}, dist={distance:.1f}, is_denied={is_denied} -> severity={severity}")

    def PredictSeverity(self, request, context):
        if not request.events:
            return smartlock.PredictSeverityResponse(
                classification=smartlock.Severity.SEVERITY_OK_UNSPECIFIED,
                confidence=1.0,
                recommendation="No events provided."
            )

        last_event = request.events[-1]

        # Feature extraction
        fails = float(last_event.fails)
        distance = float(last_event.distance_cm)

        event_name = last_event.event.lower() if last_event.event else ""
        status = last_event.status.lower() if last_event.status else ""
        detail = last_event.detail.lower() if last_event.detail else ""

        is_denied = 0.0
        if "denied" in event_name or "fail" in event_name or "denied" in detail or status == "failed":
            is_denied = 1.0

        # Normalize features
        norm_features = model_lib.normalize_features(fails, distance, is_denied)
        input_data = np.array([norm_features])

        # Thread-safe prediction
        with self.model_lock:
            predictions = self.model.predict(input_data)

        class_idx = int(np.argmax(predictions[0]))
        confidence = float(np.max(predictions[0]))

        recommendations = {
            0: "Status normal. Access allowed.",
            1: "Minor irregular patterns. Monitor access closely.",
            2: "Suspicious activity detected. Notify administrator.",
            3: "Critical threat: Multiple authentication failures. Deny access and trigger security protocols."
        }

        recommendation = recommendations.get(class_idx, "Unknown risk classification.")

        # Ensure we return valid enum values
        severity_mapping = {
            0: smartlock.Severity.SEVERITY_OK_UNSPECIFIED,
            1: smartlock.Severity.SEVERITY_IRREGULAR,
            2: smartlock.Severity.SEVERITY_SUSPICIOUS,
            3: smartlock.Severity.SEVERITY_CRITICAL
        }

        return smartlock.PredictSeverityResponse(
            classification=severity_mapping.get(class_idx, smartlock.Severity.SEVERITY_OK_UNSPECIFIED),
            confidence=confidence,
            recommendation=recommendation
        )

    def RetrainModel(self, request, context):
        epochs = request.epochs if request.epochs > 0 else 10
        dataset_path = request.dataset_path if request.dataset_path else DATASET_PATH

        if not os.path.exists(dataset_path):
            return smartlock.RetrainModelResponse(
                success=False,
                message=f"Dataset path '{dataset_path}' does not exist."
            )

        print(f"Retraining model on dataset: {dataset_path} for {epochs} epochs...")

        try:
            # Load the dataset
            X, y = model_lib.load_dataset(dataset_path)

            # Thread-safe retraining and swap
            with self.model_lock:
                # Compile a new model to train fresh or load and fit
                model = model_lib.create_model()
                model_lib.train_model(model, X, y, epochs=epochs)
                model.save(MODEL_PATH)
                self.model = tf.keras.models.load_model(MODEL_PATH)

            msg = f"Model retrained successfully on {len(X)} samples and updated in memory."
            print(msg)
            return smartlock.RetrainModelResponse(success=True, message=msg)

        except Exception as e:
            error_msg = f"Retraining failed: {str(e)}"
            print(error_msg)
            return smartlock.RetrainModelResponse(success=False, message=error_msg)

    def EvaluateModel(self, request, context):
        return super().EvaluateModel(request, context)

class RabbitMQConsumer(threading.Thread):
    def __init__(self, ai_service):
        super().__init__()
        self.ai_service = ai_service
        self.daemon = True
        self.running = True

    def run(self):
        while self.running:
            try:
                print("Connecting to RabbitMQ broker (ai-service side)...")
                connection = pika.BlockingConnection(
                    pika.ConnectionParameters(
                        host='rabbitmq',
                        port=5672,
                        credentials=pika.PlainCredentials('guest', 'guest'),
                        heartbeat=600
                    )
                )
                channel = connection.channel()
                channel.queue_declare(queue='sensor_events', durable=True)

                def callback(ch, method, properties, body):
                    try:
                        event_data = json.loads(body)
                        self.ai_service.handle_async_event(event_data)
                    except Exception as e:
                        print(f"Error processing RabbitMQ message: {e}")
                    ch.basic_ack(delivery_tag=method.delivery_tag)

                channel.basic_consume(queue='sensor_events', on_message_callback=callback)
                print("RabbitMQ consumer connected and listening on 'sensor_events' queue.")
                channel.start_consuming()
            except pika.exceptions.AMQPConnectionError as e:
                print(f"RabbitMQ connection error: {e}. Retrying in 5 seconds...")
                time.sleep(5)
            except Exception as e:
                print(f"RabbitMQ error: {e}. Retrying in 5 seconds...")
                time.sleep(5)

def serve():
    ai_service = AIService()

    # Start the RabbitMQ consumer thread
    consumer = RabbitMQConsumer(ai_service)
    consumer.start()

    # Start the gRPC server
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    smartlock_grpc.add_AIServiceServicer_to_server(ai_service, server)
    server.add_insecure_port('[::]:50051')
    print("AI Service is running on port 50051...")
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
