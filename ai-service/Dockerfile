FROM python:3.9-slim
ENV PYTHONPATH=src:src/gen

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy generated gRPC code and source
COPY src/ ./src/

# Set PYTHONPATH to include both src and generated code directories

# Expose gRPC port
EXPOSE 50051

CMD ["python", "src/main.py"]
