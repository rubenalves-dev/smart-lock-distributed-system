FROM python:3.9-slim
WORKDIR /app

# Como o ficheiro está ao lado do Dockerfile, não usamos "src/"
COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt

# Copia tudo o resto (incluindo a pasta src com o teu código)
COPY . .

# Comando para correr o main dentro da pasta src
CMD ["python", "src/main.py"]