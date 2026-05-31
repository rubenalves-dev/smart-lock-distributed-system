# Programação Concorrente e Distribuída (PCD) — Documentação da Arquitetura de Comunicação e do Backend

Este documento descreve a infraestrutura distribuída do sistema, os padrões de concorrência aplicados no Go Backend e os fluxos de mensagens através dos múltiplos protocolos de comunicação (REST, gRPC, MQTT e RabbitMQ).

---

## 1. Dimensão Teórica
A dimensão teórica aborda a taxonomia de sistemas distribuídos, os padrões de concorrência baseados em partilha de memória e troca de mensagens, e a separação de responsabilidades no armazenamento de dados.

### Sincronismo vs. Assincronismo nos Protocolos de Comunicação
O sistema foi concebido para desacoplar as tarefas críticas imediatas das tarefas analíticas ou de monitorização de estado secundárias, recorrendo a diferentes paradigmas de comunicação:

| Protocolo | Tipo de Fluxo | Síncrono / Assíncrono | Justificação e Caso de Uso |
| :--- | :--- | :--- | :--- |
| **REST (HTTP/HTTPS)** | Request-Response | Síncrono | Adequado para operações transacionais instantâneas, como o registo de novos utilizadores ou verificação de permissões do cartão RFID na fechadura. |
| **gRPC (HTTP/2)** | RPC | Síncrono | Protocolo de alto desempenho orientado a serviços, utilizado para a inferência e re-treino de IA em tempo real entre o Go Backend e o `ai-service`. |
| **MQTT** | Pub-Sub | Assíncrono | Protocolo leve ideal para telemetria IoT. Permite à fechadura publicar dados e receber comandos remotos de desbloqueio em canais dedicados sem manter ligações HTTP persistentes e pesadas. |
| **RabbitMQ (AMQP)** | Message Queue | Assíncrono | Fila de mensagens durável para desacoplamento de tarefas. Envia eventos de telemetria e pulsações de estado para processamento em segundo plano por serviços parceiros, garantindo tolerância a picos de carga. |

### Separação de Dados: Relacional vs. Séries Temporais
*   **PostgreSQL (Relacional)**: Armazena o estado transacional e estruturado (entidades de utilizadores, cartões associados, permissões lógicas e registos persistentes de acessos). A consistência ACID é essencial para garantir a fiabilidade no controlo de acessos.
*   **InfluxDB (Séries Temporais)**: Otimizado para alta frequência de escritas e consultas por intervalos temporais. Armazena os dados de telemetria de saúde de todo o ecossistema distribuído (disponibilidade dos serviços e latências medidas a cada 10 segundos).

---

## 2. Dimensão Técnica
A dimensão técnica demonstra como a linguagem Go e as suas primitivas de concorrência (Goroutines, Canais e Trincos) implementam a comunicação assíncrona tolerante a falhas.

### Ficheiros Relevantes:
*   [main.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/cmd/api/main.go): Ponto de entrada do backend, inicia goroutines paralelas e as rotas HTTP do `go-chi`.
*   [rabbitmq.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/core/rabbitmq.go): Interface de ligação à fila AMQP.
*   [monitor.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/monitor/monitor.go): Monitor de integridade do sistema que corre concorrentemente.
*   [service.go](file:///Users/rubenalves/Documents/repos/_school/iot/final/backend/internal/domain/telemetry/service.go): Ingestão assíncrona e orquestração.

### Concorrência no Go Backend: Primitivas Implementadas

#### 1. Consumo Não-Bloqueante de Canais (Goroutines)
No ficheiro `main.go`, a telemetria recebida através do MQTT é empurrada para um canal em memória Go chamado `telemetryChan`. Uma goroutine paralela consome continuamente este canal de forma assíncrona, libertando o cliente MQTT para processar novos pacotes de rede sem bloqueios físicos:

```go
// main.go
telemetryChan := make(chan models.SensorPayload, 100)

// Inicia um thread Go para processar mensagens recebidas no canal
go func() {
    for event := range telemetryChan {
        _ = telemetryService.Ingest(context.Background(), event)
    }
}()
```

#### 2. Monitorização e Escrita Paralela no InfluxDB
O serviço `monitor.go` executa uma tarefa recorrente a cada 10 segundos usando um `time.Ticker`. As escritas no InfluxDB são enviadas de forma concorrente em background usando uma goroutine anónima para evitar latências no ciclo principal do monitor:

```go
// monitor.go
func (m *Monitor) writeStatusesToInflux(statuses map[string]ServiceStatus) {
    writeAPI := m.influxClient.WriteAPI(m.influxOrg, m.influxBucket)

    // Goroutine paralela para ler e registar erros da API InfluxDB
    go func() {
        for err := range writeAPI.Errors() {
            log.Printf("InfluxDB Write error: %v", err)
        }
    }()
    ...
    writeAPI.Flush()
}
```

#### 3. Proteção Concorrente de Conexões (Mutexes)
Uma vez que múltiplos pedidos HTTP e tarefas assíncronas podem aceder simultaneamente ao cliente RabbitMQ (`RabbitMQClient`), é utilizado um trinco de leitura/escrita (`sync.RWMutex`) para garantir a integridade dos ponteiros de rede de canal e de ligação AMQP:

```go
// rabbitmq.go
type RabbitMQClient struct {
	mu             sync.RWMutex
	conn           *amqp.Connection
	channel        *amqp.Channel
	sensorQueue    amqp.Queue
	heartbeatQueue amqp.Queue
}

func (r *RabbitMQClient) PublishSensorEvent(body []byte) error {
	r.mu.RLock() // Bloqueio de leitura concorrente
	defer r.mu.RUnlock()

	if r.channel == nil || r.conn == nil || r.conn.IsClosed() {
		return amqp.ErrClosed
	}
    ...
}
```

---

## 3. Dimensão Pedagógica
A dimensão pedagógica fornece diretrizes para mapear visualmente a trajetória de dados no sistema distribuído e validar as suas capacidades concorrentes.

### A Trajetória de Dados (End‑to‑End)
1.  **O Evento**: O utilizador passa o cartão RFID na fechadura ESP32.
2.  **Edge Query**: O ESP32 executa um GET HTTP síncrono para o Backend.
3.  **Transação Postgres**: O Backend valida o cartão na BD PostgreSQL e devolve o estado à fechadura (que atualiza o seu cache local).
4.  **Comando de Abertura**: Se válido, o Backend publica uma mensagem `UNLOCK` no broker MQTT. O ESP32, subscritor deste tópico, recebe a ordem e tranca/destranca a porta fisicamente.
5.  **Telemetria Assíncrona**: O ESP32 publica no canal MQTT de telemetria `lock/telemetry`. O Backend apanha o evento, insere-o no Postgres e empurra o evento JSON para a fila `sensor_events` no RabbitMQ de forma concorrente.
6.  **Inferência Inteligente**: O `ai-service`, que consome a fila `sensor_events`, analisa o comportamento e devolve o veredito de severidade de risco ao Backend por meio de uma chamada gRPC.

### Atividades Práticas Recomendadas:
1.  **Explorar o RabbitMQ Management**:
    Aceda a `http://localhost:15672` (credenciais: `guest`/`guest`). Ao passar cartões na fechadura, verifique o gráfico de débito de mensagens nas filas `sensor_events` e `heartbeat_events`.
2.  **Verificar a Tolerância a Falhas**:
    *   Mande abaixo o contentor do `ai-service`.
    *   Continue a passar cartões na fechadura inteligente.
    *   Verifique se o sistema continua operacional (a fechadura continua a abrir e a registar logs de acessos no PostgreSQL) e se o erro de ligação gRPC é devidamente capturado no Go Backend sem deitar abaixo o servidor inteiro.
    *   Ao reiniciar o `ai-service`, comprove que os dados acumulados na fila RabbitMQ são consumidos normalmente.
