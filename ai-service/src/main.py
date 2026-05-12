from concurrent import futures
import grpc
import numpy as np
import tensorflow as tf
import pika
import json
import threading

import gen.lock_pb2 as smartlock
import gen.lock_pb2_grpc as smartlock_grpc

class AIService(smartlock_grpc.AIServiceServicer):
    def __init__(self):
        self.model = tf.keras.models.load_model("severity_model.keras")
        self.entry_map = {"vibration": 0, "nfc": 1, "fingerprint": 2}

    def PredictSeverity(self, request, context):
        last_event = request.events[-1]
        return self._process_prediction(
            last_event.vibration_intensity, 
            last_event.entry_method, 
            last_event.success
        )

    def _process_prediction(self, vibration, method_name, success):
        method = self.entry_map.get(method_name.lower(), 0)
        success_val = 1.0 if success else 0.0

        input_data = np.array([[vibration, method, success_val]])
        predictions = self.model.predict(input_data)

        class_idx = np.argmax(predictions[0])
        confidence = float(np.max(predictions[0]))

        recommendations = {
            0: "Everything is fine.",
            1: "Monitor closely",
            2: "Notify user",
            3: "Request MFA immediately."
        }

        print(f"🤖 AI Prediction: Class {class_idx} ({confidence*100:.2f}%) - {method_name}")

        return smartlock.PredictSeverityResponse(
            classification=class_idx,
            confidence=confidence,
            recommendation=recommendations.get(int(class_idx), "Unknown classification")
        )

def consume_rabbit(ai_service):
    try:
        connection = pika.BlockingConnection(pika.ConnectionParameters(host='rabbitmq_broker'))
        channel = connection.channel()
        channel.queue_declare(queue='sensor_events', durable=True)

        def callback(ch, method, properties, body):
            data = json.loads(body)
            print(f"📥 RabbitMQ message received: {data}")
            ai_service._process_prediction(data['Value'], data['EventType'], True)

        channel.basic_consume(queue='sensor_events', on_message_callback=callback, auto_ack=True)
        channel.start_consuming()
    except Exception as e:
        print(f"❌ RabbitMQ Error: {e}")

def serve():
    ai_instance = AIService()

    rabbit_thread = threading.Thread(target=consume_rabbit, args=(ai_instance,), daemon=True)
    rabbit_thread.start()

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    smartlock_grpc.add_AIServiceServicer_to_server(ai_instance, server)
    server.add_insecure_port('[::]:50051')
    
    server.start()
    server.wait_for_termination()

if __name__ == "__main__":
    serve()