FROM python:3.9-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy generated gRPC code and source
COPY src/ ./src/

# Expose gRPC port
EXPOSE 50051

CMD ["python", "src/main.py"]
