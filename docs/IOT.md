# Internet das Coisas (IoT) — Documentação da Fechadura Inteligente

Este documento detalha o subsistema de hardware e firmware da fechadura inteligente. Abrange a fase de simulação baseada em PlatformIO (`/firmware`), a implementação final adaptada para o Arduino IDE (`/arduino-ide`), as suas interfaces de comunicação física e de rede, bem como as decisões de design arquitetónico.

---

## 1. Dimensão Teórica
A dimensão teórica foca nos conceitos de computação de proximidade (Edge Computing), gestão de sensores/atuadores, e resiliência a falhas de comunicação de rede.

### O Paradigma Edge‑to‑Cloud no Controlo de Acessos
Tradicionalmente, fechaduras eletrónicas operavam de forma puramente local (tabelas estáticas na memória da fechadura) ou puramente ligada à nuvem (enviando a leitura do cartão para o servidor e aguardando resposta). Ambas têm limitações graves:
*   **Pure local**: Falta de gestão centralizada. Para aceitar ou bloquear utilizadores, o administrador tem de interagir fisicamente com cada fechadura.
*   **Pure cloud**: Falha catastrófica caso a ligação à Internet caia (ficando preso fora de casa/escritório).

Este sistema implementa uma abordagem híbrida **Offline‑First**:
*   A fechadura tenta validar e atualizar os dados do cartão em tempo real consultando a API central via chamadas HTTP.
*   Se a API responder com sucesso, o estado do cartão (`is_accepted` / `is_blocked`) é atualizado na fechadura e **persistido localmente** na memória não volátil (NVRAM).
*   Se o servidor não for alcançado (falha de WiFi ou API em baixo), a fechadura transita para um **fallback offline**, validando o cartão contra a cópia de segurança guardada na memória local.

### Sensores e Atuadores
O sistema é constituído por:
*   **Sensor RFID (MFRC522)**: Atua como interface de entrada primária de credenciais via barramento SPI.
*   **Sensor de Ultrassons (HC-SR04)**: Mede a distância física do objeto aproximado para evitar a leitura fantasma ou gasto excessivo de processamento, servindo como uma "porta lógica" física.
*   **Fotoresistor (LDR)**: Mede o nível de luminosidade ambiente.
*   **Atuador de Tranca**: Controla o fecho físico da fechadura (representado por um motor passo/servo ou tranca eletromagnética).

---

## 2. Dimensão Técnica
A dimensão técnica demonstra a implementação prática do firmware para o microcontrolador ESP32.

### Ficheiros Relevantes:
*   [main.cpp](file:///Users/rubenalves/Documents/repos/_school/iot/final/firmware/src/main.cpp): Código da fase de simulação (Wokwi/PlatformIO) recorrendo a servidores assíncronos.
*   [main.ino](file:///Users/rubenalves/Documents/repos/_school/iot/final/arduino-ide/main/main.ino): Código de produção do Arduino IDE com otimizações para compilação física e AP fallback.

### Transição Crítica: Servidor Assíncrono vs. Síncrono
*   No simulador `/firmware`, o código recorria à biblioteca `ESPAsyncWebServer`. Embora eficiente no processamento de pedidos simultâneos, introduzia incompatibilidades com versões recentes do core do ESP32 e causava colisões de semáforos no FreeRTOS ao misturar threads de sockets com comunicações de rede físicas.
*   No código final `/arduino-ide`, fez-se a transição para a biblioteca síncrona standard `WebServer.h`. Para evitar que o thread síncrono bloqueasse a resposta a eventos físicos e as rotinas de rede, o loop principal foi desenhado de forma não bloqueante, regulado por timers via `millis()`.

### Otimização da Leitura RFID e Gating Físico
No código original, a leitura de cartões ocorria em blocos lentos de 1 segundo. No firmware de produção (`main.ino`), o sensor de cartões é **polido continuamente** a cada ciclo de relógio do processador para garantir resposta instantânea. No entanto, este polimento contínuo via barramento SPI consome CPU e gera ruído.
Para otimizar, o polimento é gateado pelo sensor ultrassónico:

```cpp
// --- RFID Check (polido em cada ciclo apenas se houver aproximação física)
if (isClose) {
  if (rfid.readCard()) {
    // Prossiga com a verificação de autenticação
    ...
  }
}
```

### Mecanismo de Cache Local usando `Preferences`
O ESP32 fornece a API `Preferences` que mapeia pares chave-valor diretamente para partições NVS (Non-Volatile Storage) na Flash. Isto previne a perda de dados em caso de corte de energia elétrica.

```cpp
// Escrita de estado no cache
preferences.begin("cards", false);
preferences.putInt(uidStr.c_str(), newStatus); // 0=Pending, 1=Accepted, 2=Blocked
preferences.end();

// Leitura offline do cache
preferences.begin("cards", true);
int localStatus = preferences.getInt(uidStr.c_str(), -1);
preferences.end();
```

### Comunicação com Rede
O ESP32 comunica bidirecionalmente:
1.  **Outbound HTTP Client**: Envia pedidos GET para `https://smartlock-api.raiiaa.dev/api/users/{uid}` para verificar permissões e atualizar o cache local.
2.  **Inbound Web Server**: Escuta no porto 80 oferecendo endpoints locais como `/open` (abertura manual por painel local), `/status`, e `/wifi-save` (para reconfiguração).
3.  **MQTT Client**: Subscreve o tópico `lock/commands` e publica relatórios de telemetria no tópico `lock/telemetry`. Recebe o comando remoto de abertura `UNLOCK` a partir do Backend.

---

## 3. Dimensão Pedagógica
A dimensão pedagógica aborda os processos de validação prática, depuração e utilização da fechadura como ferramenta educativa.

### Ponto de Configuração Amigável: Modo de Ponto de Acesso (AP Mode)
Se o ESP32 não conseguir estabelecer ligação à rede local WiFi pré-configurada nos primeiros 10 segundos, ele reage inteligentemente:
1.  Inicia a sua própria rede WiFi aberta chamada `SmartLock-Setup` (palavra-passe padrão: `12345678`).
2.  Apresenta um painel de controlo Web no endereço IP estático `http://192.168.4.1`.
3.  O utilizador liga o telemóvel à rede, entra no browser e preenche as novas credenciais de rede, as quais são salvas na NVRAM, provocando o reinício automático do microcontrolador.

### Diagnósticos via Linha de Comandos e Porta Série
O firmware inclui logs detalhados enviados à porta de depuração série (`Serial` a 115200 baud). Isto permite aos estudantes inspecionarem o fluxo interno do sistema em tempo real:
*   `[Telemetry] Publishing to lock/telemetry:...` — Mostra o JSON que descreve os estados dos sensores no momento em que é transmitido.
*   `[RFID] API Request returned HTTP code: 200` — Confirma que o servidor respondeu corretamente.
*   `[RFID] Loaded cache for UID=DE:AD:BE:EF: status=1` — Indica a execução bem-sucedida do fallback offline.

### Exercício de Laboratório Recomendado:
1.  **Testar Validação Online**: Aproxime um cartão e verifique nos logs da consola série se a resposta HTTP é 200 e se o trincamento de fechadura foi efetuado.
2.  **Testar Resiliência Offline**:
    *   Desligue o router WiFi ou mande abaixo o backend local.
    *   Passe um cartão aceite que tenha sido previamente validado online.
    *   Verifique na consola série a mensagem `[RFID] API status not fetched. Querying local Preferences cache...` e confirme que a porta abre mesmo sem rede.
