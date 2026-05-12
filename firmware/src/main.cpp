#include <Arduino.h>
#include <WiFi.h>
#include <AsyncTCP.h>
#include <ESPAsyncWebServer.h>
#include <PubSubClient.h>
#include <pb_encode.h>
#include "lock.pb.h" 

// Configurações
const char* ssid = "Wokwi-GUEST";
const char* password = "";
const char* mqtt_server = "host.wokwi.internal";

WiFiClient espClient;
PubSubClient mqttClient(espClient);
AsyncWebServer server(80);

const int LED_PIN = 2; 
bool isLocked = true;

// Interface HTML 
const char index_html[] PROGMEM = R"rawliteral(
<!DOCTYPE HTML><html><head>
  <meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1">
  <style>
    body { font-family: sans-serif; background-color: #f0f2f5; display: flex; justify-content: center; padding-top: 50px; }
    .card { background: white; padding: 20px; border-radius: 15px; box-shadow: 0 4px 10px rgba(0,0,0,0.1); width: 350px; text-align: center; }
    .status { font-weight: bold; font-size: 24px; margin-bottom: 20px; }
    .status.open { color: #4CAF50; }
    .status.locked { color: #f44336; }
    .btn { display: block; width: 100%; padding: 12px; margin: 10px 0; border: none; border-radius: 8px; font-weight: bold; cursor: pointer; color: white; text-transform: uppercase; }
    .btn-unlock { background-color: #3498db; }
    .btn-wifi { background-color: #95a5a6; }
    .btn-add { background-color: #2ecc71; }
    .btn-check { background-color: #7f8c8d; }
  </style>
</head><body>
  <div class="card">
    <div id="lockState" class="status locked">LOCKED</div>
    <button class="btn btn-unlock" onclick="sendAction('/toggle')">REMOTELY UNLOCK / LOCK</button>
    <button class="btn btn-wifi" onclick="sendAction('/wifi-info')">WIFI LOGIN</button>
    <button class="btn btn-add" onclick="sendAction('/add-user')">ADD USERS</button>
    <button class="btn btn-check" onclick="sendAction('/check-services')">VERIFICAR SERVIÇOS</button>
  </div>
  <script>
    function sendAction(path) {
      fetch(path).then(r => r.text()).then(txt => {
        if(path === '/toggle') {
           const s = document.getElementById('lockState');
           if(txt.includes('ABERTA')) { s.innerText = 'OPEN'; s.className = 'status open'; }
           else { s.innerText = 'LOCKED'; s.className = 'status locked'; }
        } else { alert(txt); }
      });
    }
  </script>
</body></html>)rawliteral";

void setup() {
    Serial.begin(115200);
    pinMode(LED_PIN, OUTPUT);
    digitalWrite(LED_PIN, LOW);
    
    WiFi.begin(ssid, password);
    while (WiFi.status() != WL_CONNECTED) delay(500);

    mqttClient.setServer(mqtt_server, 1883);

    server.on("/", HTTP_GET, [](AsyncWebServerRequest *request){
        request->send_P(200, "text/html", index_html);
    });

    server.on("/toggle", HTTP_GET, [](AsyncWebServerRequest *request){
        isLocked = !isLocked;
        digitalWrite(LED_PIN, isLocked ? LOW : HIGH);
        request->send(200, "text/plain", isLocked ? "PORTA FECHADA" : "PORTA ABERTA");
    });

    server.on("/add-user", HTTP_GET, [](AsyncWebServerRequest *request){
    // Usamos R"raw(...)raw" para o C++ ignorar as aspas lá dentro
    String json = R"raw({"device_id":"esp32_hugo","event_type":"vibration","value":9.8})raw";
    
    if (mqttClient.connected()) {
        if (mqttClient.publish("lock/events", json.c_str())) {
            request->send(200, "text/plain", "JSON enviado com sucesso!");
        } else {
            request->send(500, "text/plain", "Falha no envio MQTT");
        }
    } else {
        request->send(500, "text/plain", "MQTT nao conectado");
    }
});

    server.on("/wifi-info", HTTP_GET, [](AsyncWebServerRequest *request){
        request->send(200, "text/plain", "IP: " + WiFi.localIP().toString());
    });

    server.on("/check-services", HTTP_GET, [](AsyncWebServerRequest *request){
        request->send(200, "text/plain", mqttClient.connected() ? "Servidores Online" : "Erro de Ligação");
    });

    server.begin();
}

void loop() {
    if (!mqttClient.connected()) {
        mqttClient.connect("ESP32_Hugo");
    }
    mqttClient.loop();
}