from concurrent import futures

import grpc
import numpy as np
import tensorflow as tf

import gen.lock_pb2 as smartlock
import gen.lock_pb2_grpc as smartlock_grpc


class AIService(smartlock_grpc.AIServiceServicer):
    def __init__(self):
        self.model = tf.keras.models.load_model("severity_model.keras")
        self.entry_map = {"vibration": 0, "nfc": 1, "fingerprint": 2}

    def PredictSeverity(self, request, context):
        last_event = request.events[-1]

        vibration = last_event.vibration_intensity
        method = self.entry_map.get(last_event.entry_method.lower(), 0)
        success = 1.0 if last_event.success else 0.0

        input_data = np.array([[vibration, method, success]])
        predictions = self.model.predict(input_data)

        class_idx = np.argmax(predictions[0])
        confidence = float(np.max(predictions[0]))

        recommendations = {
            0: "Everything is fine.",
            1: "Monitor closely",
            2: "Notify user",
            3: "Request MFA immediately."
        }

        return smartlock.PredictSeverityResponse(
            classification=class_idx,
            confidence=confidence,
            recommendation=recommendations.get(int(class_idx), "Unknown classification")
        )

    def RetrainModel(self, request, context):
        return super().RetrainModel(request, context)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    smartlock_grpc.add_AIServiceServicer_to_server(AIService, server)
    server.add_insecure_port('[::]:50051')
    print("AI Service is running on port 50051...")
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()
