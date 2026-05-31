# Inteligência Artificial Aplicada (IIA) — Documentação do Serviço de IA

Este documento descreve detalhadamente o serviço de Inteligência Artificial (`ai-service`), a sua arquitetura, funcionamento interno, métodos de comunicação e a sua posição global no sistema de controlo de acessos. A explicação está organizada em três dimensões estruturantes: Teórica, Técnica e Pedagógica.

---

## 1. Dimensão Teórica
A dimensão teórica aborda os princípios de Aprendizagem de Máquina (Machine Learning) supervisionada aplicados na deteção de anomalias e avaliação de riscos.

### O Problema Central
O objetivo do `ai-service` é classificar a **severidade do risco** de um evento de acesso físico a uma fechadura inteligente. O sistema analisa um histórico de tentativas e decide o nível de perigo com base em três variáveis observacionais (atributos/features):
1. **Número de falhas consecutivas (`fails`)**: Quantidade de vezes seguidas que um cartão inválido ou não autorizado foi lido.
2. **Distância de aproximação (`distance_cm`)**: A distância física (medida pelo sensor de ultrassons) do objeto à porta no momento da leitura.
3. **Indicador de negação de acesso (`is_denied`)**: Um valor binário (0 ou 1) que indica se o acesso foi recusado pelo sistema de controlo.

### Classificação Multiclasse e Mapeamento de Severidade
O modelo resolve um problema de **classificação supervisionada multiclasse**, categorizando a severidade em quatro níveis (classes):
*   **Classe 0 (SEVERITY_OK_UNSPECIFIED)**: Operação normal. Sem falhas de autenticação, distância regulamentar.
*   **Classe 1 (SEVERITY_IRREGULAR)**: Comportamento ligeiramente atípico (ex.: uma leitura falhada mas à distância normal).
*   **Classe 2 (SEVERITY_SUSPICIOUS)**: Comportamento suspeito (ex.: duas falhas consecutivas ou uma aproximação física extrema combinada com negação).
*   **Classe 3 (SEVERITY_CRITICAL)**: Ameaça crítica de segurança (ex.: 3 ou mais falhas consecutivas, indicando uma tentativa ativa de força bruta à curta distância).

### Arquitetura da Rede Neuronal (Feedforward Artificial Neural Network)
O modelo baseia-se numa rede neuronal artificial densamente ligada (Multi-Layer Perceptron):
*   **Camada de Entrada**: 3 neurónios, correspondentes aos atributos normalizados.
*   **Camadas Ocultas (Hidden Layers)**: Duas camadas densas com 16 neurónios cada, utilizando a função de ativação **ReLU (Rectified Linear Unit)** para introduzir não-linearidade:
    $$f(x) = \max(0, x)$$
*   **Camada de Saída**: 4 neurónios com função de ativação **Softmax** para calcular uma distribuição de probabilidade sobre as classes:
    $$\text{Softmax}(z_i) = \frac{e^{z_i}}{\sum_{j=1}^{4} e^{z_j}}$$
*   **Função de Perda (Loss Function)**: **Sparse Categorical Cross-Entropy**. É ideal para problemas de classificação multiclasse onde as classes alvo (`y`) são representadas por números inteiros (0 a 3) e não por vetores *one-hot*.
*   **Otimizador**: **Adam (Adaptive Moment Estimation)**, que calcula taxas de aprendizagem adaptativas para cada parâmetro da rede a partir de estimativas do primeiro e segundo momentos dos gradientes.

### Normalização dos Atributos
As redes neuronais convergem melhor quando as features de entrada estão na mesma escala (tipicamente $[0, 1]$). O serviço implementa uma normalização linear limitada (*clamping*):
*   **Fails**: Escalado por $5.0$ (qualquer valor $\ge 5$ é limitado a $1.0$).
*   **Distance**: Escalado por $150.0$ (qualquer distância $\ge 150\text{ cm}$ é limitada a $1.0$).
*   **Is Denied**: Já é binário ($0.0$ ou $1.0$).

### Sobreajuste (Overfitting) vs. Subajuste (Underfitting)
Durante o treino do modelo, monitorizam-se métricas em dados de treino e de validação (20% de partilha):
*   **Underfitting (Subajuste)**: Ocorre quando o modelo não consegue capturar a estrutura subjacente dos dados. É identificado se a precisão do treino (`train_accuracy`) e da validação (`val_accuracy`) forem ambas inferiores a $70\%$.
*   **Overfitting (Sobreajuste)**: Ocorre quando a rede decora o ruído dos dados de treino, perdendo capacidade de generalização. É detetado se a diferença de exatidão entre treino e validação for superior a $10\%$ ($\text{Acc}_{\text{train}} - \text{Acc}_{\text{val}} > 0.10$).

---

## 2. Dimensão Técnica
A dimensão técnica demonstra a implementação em código Python (`keras` e `tensorflow`) dos conceitos teóricos expostos.

### Ficheiros Relevantes:
*   [model.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/model.py): Define o grafo da rede, normalização e rotinas de treino.
*   [main.py](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/src/main.py): Implementa o servidor gRPC e o consumidor assíncrono RabbitMQ.

### Código de Definição do Modelo (`model.py`)
```python
def create_model():
    model = keras.Sequential([
        # Primeira camada densa de 16 neurónios com ativação ReLU
        layers.Dense(16, activation="relu", input_shape=(3,)),
        # Segunda camada densa de 16 neurónios com ativação ReLU
        layers.Dense(16, activation="relu"),
        # Camada de saída com ativação Softmax para distribuição de probabilidades
        layers.Dense(4, activation="softmax")
    ])

    model.compile(
        optimizer="adam",
        loss="sparse_categorical_crossentropy",
        metrics=["accuracy"]
    )
    return model
```

### Comunicação Distribúida e Concorrência (`main.py`)
O serviço de IA atua de forma dupla na rede distribuída:

#### 1. Ingestão Assíncrona via RabbitMQ (`RabbitMQConsumer`)
Um thread paralelo de execução em segundo plano liga-se ao broker RabbitMQ na fila `sensor_events`. Cada mensagem contendo telemetria é recolhida e processada:
*   Uma heurística rotula o evento dinamicamente (para fins de compilação contínua do dataset).
*   O evento é persistido no ficheiro CSV local [sensor_events.csv](file:///Users/rubenalves/Documents/repos/_school/iot/final/ai-service/data/sensor_events.csv).
*   **Mecanismo de Concorrência**: Para evitar condições de corrida (*race conditions*) quando o modelo é treinado ou avaliado ao mesmo tempo que eventos assíncronos chegam, são usados trincos de exclusão mútua (`threading.Lock`):
    *   `self.model_lock`: Bloqueia o modelo durante a inferência ou re-carregamento.
    *   `self.dataset_lock`: Protege as escritas simultâneas no ficheiro CSV.

```python
with self.dataset_lock:
    # Apenas uma escrita por thread de cada vez no ficheiro CSV
    with open(DATASET_PATH, mode='a', newline='') as f:
        writer = csv.writer(f)
        writer.writerow([fails, distance, is_denied, severity])
```

#### 2. Servidor de Chamadas de Procedimento Remoto (gRPC Server)
O Go Backend comunica com o `ai-service` via gRPC no porto `50051`. A interface protobuf (`lock.proto`) expõe três métodos críticos:
*   `PredictSeverity`: Recebe uma lista de eventos, extrai features do último evento, normaliza e corre a inferência em tempo real:
    ```python
    with self.model_lock:
        predictions = self.model.predict(input_data)
    class_idx = int(np.argmax(predictions[0]))
    confidence = float(np.max(predictions[0]))
    ```
*   `RetrainModel`: Disparado sob demanda pelo administrador. Cria uma nova rede neuronal, treina-a sobre o dataset acumulado, avalia o subajuste/sobreajuste e guarda o ficheiro do modelo atualizado (`.keras`), atualizando a referência em memória de forma segura.
*   `EvaluateModel`: Calcula métricas estatísticas sobre um conjunto de dados enviado na mensagem: **Matriz de Confusão**, **Exatidão**, **Precisão**, **Sensibilidade (Recall)** e **F1-Score**.

---

## 3. Dimensão Pedagógica
A dimensão pedagógica detalha como estudantes, investigadores e engenheiros de sistemas podem interagir com esta tecnologia para fins de ensino e aprendizagem prática.

### Como Ensinar e Aprender com este Serviço:
1.  **Monitorização de Métricas**: Ao navegar no Vue Frontend na página de Treino, o utilizador pode visualizar curvas de perda (*loss*) e exatidão (*accuracy*) para ambos os dados (treino vs. validação). Isso ensina de forma interativa a interpretação física de métricas matemáticas.
2.  **Experimentação com Hiperparâmetros**: Os estudantes podem efetuar pedidos HTTP à API REST `/api/ai/retrain` modificando o número de épocas (`epochs`) de treino ou apontando para diferentes caminhos de ficheiros CSV. O retorno diagnóstico ajuda a compreender o impacto de ciclos de treino mais longos.
3.  **Simulação de Ataques de Segurança (Análise Dinâmica)**:
    Ao interagir com o ESP32 simulado no Wokwi ou no hardware real, pode simular-se um atacante físico:
    *   *Caso Prático*: Aproximar o objeto a menos de 10cm da porta e introduzir cartões inválidos 3 vezes consecutivas.
    *   *Comportamento Esperado*: O monitor de diagnóstico do backend irá receber a classificação `SEVERITY_CRITICAL` e disparar imediatamente um comando de autenticação multifator (MFA) ou fecho total do sistema.
    *   *Aprendizagem*: Demonstra na prática como um modelo de Machine Learning serve de gatilho para políticas reativas de cibersegurança num sistema de automação.

### Passos Rápidos para Execução em Laboratório:
1.  **Criar ou Iniciar o Ambiente**:
    O serviço é executado automaticamente com o Docker Compose na diretoria raiz:
    ```bash
    docker compose up --build ai-service
    ```
2.  **Verificar Logs do Modelo**:
    Para examinar a inicialização do modelo base e o treino inicial de 10 épocas, consulte os logs:
    ```bash
    docker logs -f ai_service
    ```
3.  **Teste de Avaliação de Dados Personalizados**:
    Pode enviar dados arbitrários em formato CSV para ver o modelo classificar e retornar a Matriz de Confusão através do endpoint POST `/api/ai/evaluate` do backend Go.
