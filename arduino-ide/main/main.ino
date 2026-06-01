#include <Arduino.h>
#include <ArduinoJson.h>
#include <ESPmDNS.h>
#include <Preferences.h>
#include <WebServer.h>
#include <HTTPClient.h>
#include <WiFi.h>

#include <INDEX.h>
#include "WIFI_HTML.h"

#include "Timer.h"
#include "RFID.h"
#include "Ultrassonic.h"
#include "Photoresistor.h"
#include "MQTT.h"

void updateLockState(bool lock, String user = "system");

// --- RFID
#define RFID_SDA_PIN 5
#define RFID_RST_PIN 22
#define RFID_SCK_PIN 18
#define RFID_MISO_PIN 19
#define RFID_MOSI_PIN 23
RFID rfid(RFID_SDA_PIN, RFID_RST_PIN);

// --- Ultrasonic
#define TRIGGER_PIN 16
#define ECHO_PIN 17
#define DISTANCE_THRESHOLD_CM 100
Ultrassonic ultrassonic(TRIGGER_PIN, ECHO_PIN, DISTANCE_THRESHOLD_CM);

// --- Light Sensor (LDR)
#define LDR_PIN 4
Photoresistor ldr(LDR_PIN, 5000);

// --- MQTT
const char *MQTT_SERVER = "mqtt.raiiaa.dev";
const uint16_t MQTT_PORT = 1883;
MQTT mqtt(MQTT_SERVER, MQTT_PORT);

// --- Authorized UID
byte authorizedUID[4] = { 0xDE, 0xAD, 0xBE, 0xEF };  // Example UID, replace with actual

// --- Servers Config
const char *SSID = "Wokwi-GUEST";
const char *PASSWORD = "";
WebServer server(80);

// --- State
String lastUser = "none";
int failCount = 0;
bool isLocked = true;

// --- Servers
Preferences preferences;

void sendTelemetry(String eventType, String details, String rfidUid = "");

// --- Timers
Timer autoCloseTimer(
  5000, []() {
    updateLockState(true, "auto_close");
  },
  Timer::Mode::OneShot);
Timer telemetryTimer(
  5000, []() {
    sendTelemetry("heartbeat", "Periodic status update");
  },
  Timer::Mode::Periodic);

// --- Functions
void sendTelemetry(String eventType, String details, String rfidUid) {
  if (WiFi.status() != WL_CONNECTED) {
    Serial.println("[Telemetry] WiFi not connected. Telemetry not sent.");
    return;
  }
  StaticJsonDocument<300> doc;
  doc["device_id"] = "smartlock_esp32";
  doc["event"] = eventType;
  doc["details"] = details;

  if (rfidUid != "") {
    doc["rfid_uid"] = rfidUid;
  }

  doc["status"] = "";
  doc["distance_cm"] = ultrassonic.distance();
  doc["light_level"] = ldr.lightLevel();
  doc["fails"] = rfid.failCount();
  if (rfid.failCount() > 0) {
    doc["user"] = lastUser;
  }

  doc["rssi"] = WiFi.RSSI();
  doc["uptime"] = millis() / 1000;  // Uptime in seconds

  char buffer[300];
  serializeJson(doc, buffer);
  Serial.print("[Telemetry] Publishing to lock/telemetry: ");
  Serial.println(buffer);
  mqtt.publish("lock/telemetry", buffer);
}

void updateLockState(bool lock, String user) {
  isLocked = lock;
  Serial.print("[Lock] State updated to: ");
  Serial.print(isLocked ? "LOCKED" : "UNLOCKED");
  Serial.print(" by: ");
  Serial.println(user);
  sendTelemetry("status_change", isLocked ? "LOCKED" : "UNLOCKED", user);
}

void callback(char *topic, byte *payload, unsigned int length) {
  String msg = "";
  for (int i = 0; i < length; i++) {
    msg += (char)payload[i];
  }
  Serial.print("[MQTT] Received message on topic ");
  Serial.print(topic);
  Serial.print(": ");
  Serial.println(msg);

  if (msg == "UNLOCK") {
    Serial.println("[MQTT] AI/Backend authorized door UNLOCK");
    updateLockState(false, "mqtt_command");
  } else if (msg == "LOCK") {
    Serial.println("[MQTT] AI/Backend command: LOCK");
    updateLockState(true, "mqtt_command");
  }
}

void setupWebServer() {
  server.on("/", HTTP_GET, []() {
    server.send(200, "text/html", INDEX_HTML);
  });
  server.on("/wifi", HTTP_GET, []() {
    server.send(200, "text/html", WIFI_HTML);
  });
  server.on("/open", HTTP_GET, []() {
    updateLockState(false, "web_button");
    autoCloseTimer.reset();
    autoCloseTimer.start();
    server.send(200, "text/plain", "PORTA ABERTA");
  });
  server.on("/status", HTTP_GET, []() {
    server.send(200, "text/plain", isLocked ? "LOCKED" : "UNLOCKED");
  });
  server.on("/wifi-save", HTTP_POST, []() {
    if (server.hasArg("ssid") && server.hasArg("password")) {
      String ssid = server.arg("ssid");
      String password = server.arg("password");

      preferences.begin("wifi", false);
      preferences.putString("ssid", ssid);
      preferences.putString("password", password);
      preferences.end();

      server.send(200, "text/plain", "OK");

      delay(1000);
      ESP.restart();
    } else {
      server.send(400, "text/plain", "Missing ssid or password");
    }
  });
  server.on("/wifi-info", HTTP_GET, []() {
    if (WiFi.status() == WL_CONNECTED) {
      server.send(200, "text/plain", "IP: " + WiFi.localIP().toString());
    } else {
      server.send(200, "text/plain", "IP: " + WiFi.softAPIP().toString() + " (AP Mode)");
    }
  });
  server.on("/users", HTTP_GET, []() {
    if (WiFi.status() != WL_CONNECTED) {
      server.send(500, "text/plain", "[]");
      return;
    }
    HTTPClient http;
    http.begin("https://smartlock-api.raiiaa.dev/api/users");
    int httpCode = http.GET();
    if (httpCode > 0) {
      String payload = http.getString();
      server.send(200, "application/json", payload);
    } else {
      server.send(500, "text/plain", "[]");
    }
    http.end();
  });
  server.on("/user-details", HTTP_GET, []() {
    if (WiFi.status() != WL_CONNECTED) {
      server.send(500, "text/plain", "{}");
      return;
    }
    if (!server.hasArg("uid")) {
      server.send(400, "text/plain", "Missing uid");
      return;
    }
    String uid = server.arg("uid");
    HTTPClient http;
    http.begin("https://smartlock-api.raiiaa.dev/api/users/" + uid);
    int httpCode = http.GET();
    if (httpCode > 0) {
      String payload = http.getString();
      server.send(200, "application/json", payload);
    } else {
      server.send(500, "text/plain", "{}");
    }
    http.end();
  });
  server.on("/check-services", HTTP_GET, []() {
    String backendStatus = "{}";
    if (WiFi.status() == WL_CONNECTED) {
      HTTPClient http;
      http.begin("https://smartlock-api.raiiaa.dev/api/health");
      int httpCode = http.GET();
      if (httpCode > 0) {
        backendStatus = http.getString();
      }
      http.end();
    }

    String response = "{\"local_mqtt\":" + String(mqtt.connected() ? "true" : "false") + ",\"backend_services\":" + backendStatus + "}";
    server.send(200, "application/json", response);
  });
  server.begin();
}

void setup() {
  Serial.begin(115200);

  // Components
  ultrassonic.setup();
  ldr.setup();
  rfid.setup(RFID_SCK_PIN, RFID_MISO_PIN, RFID_MOSI_PIN);

  // Wifi
  preferences.begin("wifi", true);
  String storedSSID = preferences.getString("ssid", SSID);
  String storedPassword = preferences.getString("password", PASSWORD);
  preferences.end();

  int attempts = 0;

  WiFi.begin(storedSSID.c_str(), storedPassword.c_str());
  while (WiFi.status() != WL_CONNECTED && attempts < 20) {
    delay(500);
    Serial.print(".");
    attempts++;
  }
  if (WiFi.status() != WL_CONNECTED) {
    Serial.println("\nWiFi connection failed! Starting Access Point (AP mode)...");
    WiFi.softAP("SmartLock-Setup", "12345678");
    Serial.println("AP Mode started. SSID: SmartLock-Setup, Password: 12345678");
    Serial.print("Access the setup page at: http://");
    Serial.println(WiFi.softAPIP());
  } else {
    Serial.println("\nWifi connected, IP: " + WiFi.localIP().toString());
  }

  // Servers
  mqtt.subscribe("lock/commands", callback);
  mqtt.subscribe("lock/ai/response", callback);
  mqtt.setup();

  // Timers
  telemetryTimer.start();

  // mDNS
  if (!MDNS.begin("smartlock")) {
    Serial.println("Error setting up MDNS responder!");
  }

  // Web Server
  setupWebServer();
}

unsigned long lastSensorCheck = 0;

void loop() {
  server.handleClient();
  if (WiFi.status() == WL_CONNECTED) {
    mqtt.update();
  }
  telemetryTimer.update();
  autoCloseTimer.update();

  // --- Proximity State Transition Logging (non-flooding)
  static bool wasObjectClose = false;
  bool isClose = ultrassonic.isObjectClose();
  if (isClose != wasObjectClose)
  {
    wasObjectClose = isClose;
    Serial.print("[Sensor] Proximity state changed. Object is close: ");
    Serial.print(isClose ? "YES" : "NO");
    Serial.print(" (Distance: ");
    Serial.print(ultrassonic.distance());
    Serial.println(" cm)");
  }

  // --- 3. RFID Check (polled every 100ms for stable RF field and responsive scan detection)
  static unsigned long lastRFIDCheck = 0;
  if (millis() - lastRFIDCheck >= 100)
  {
    lastRFIDCheck = millis();
    if (rfid.readCard())
    {
      // Format UID as string (e.g. "DE:AD:BE:EF")
      char uidStrBuf[20];
      sprintf(uidStrBuf, "%02X:%02X:%02X:%02X", rfid.buffer()[0], rfid.buffer()[1], rfid.buffer()[2], rfid.buffer()[3]);
      String uidStr = String(uidStrBuf);

      Serial.print("[RFID] Card scanned! UID: ");
      Serial.println(uidStr);

      bool shouldUnlock = false;
      int newStatus = 0; // default to pending if we register it
      bool statusFetched = false;

      // Try to communicate with server if WiFi is connected
      if (WiFi.status() == WL_CONNECTED)
      {
        Serial.print("[RFID] WiFi connected. Sending API request to check card UID: ");
        Serial.println(uidStr);

        HTTPClient http;
        String url = "https://smartlock-api.raiiaa.dev/api/users/" + uidStr;
        http.begin(url);
        int httpCode = http.GET();

        Serial.print("[RFID] API Request returned HTTP code: ");
        Serial.println(httpCode);

        if (httpCode == 200)
        {
          String response = http.getString();
          Serial.print("[RFID] API Response received: ");
          Serial.println(response);

          StaticJsonDocument<300> doc;
          DeserializationError error = deserializeJson(doc, response);
          if (!error)
          {
            bool isAccepted = doc["is_accepted"] | false;
            bool isBlocked = doc["is_blocked"] | false;
            Serial.printf("[RFID] Parsed states - is_accepted: %s, is_blocked: %s\n", isAccepted ? "true" : "false", isBlocked ? "true" : "false");

            if (isBlocked)
            {
              newStatus = 2; // Blocked
              shouldUnlock = false;
            }
            else if (isAccepted)
            {
              newStatus = 1; // Accepted
              shouldUnlock = true;
            }
            else
            {
              newStatus = 0; // Pending
              shouldUnlock = false;
            }
            statusFetched = true;
          }
          else
          {
            Serial.print("[RFID] JSON deserialization failed: ");
            Serial.println(error.c_str());
          }
        }
        else if (httpCode == 404)
        {
          Serial.println("[RFID] Card not found on backend (404). Initiating auto-registration telemetry...");
          // Card not in database yet. Ingest event to trigger auto-registration!
          sendTelemetry("access_denied", "New card detected on reader", uidStr);
          newStatus = 0; // Pending
          shouldUnlock = false;
          statusFetched = true;
        }
        else
        {
          Serial.println("[RFID] Server check failed with unexpected HTTP code. Falling back to cache.");
        }
        http.end();
      }
      else
      {
        Serial.println("[RFID] WiFi not connected. Skipping API server check.");
      }

      if (statusFetched)
      {
        Serial.printf("[RFID] Saving card status to Preferences cache 'cards': UID=%s, status=%d\n", uidStr.c_str(), newStatus);
        // Update local preferences cache
        preferences.begin("cards", false);
        preferences.putInt(uidStr.c_str(), newStatus);
        preferences.end();

        if (shouldUnlock)
        {
          lastUser = uidStr;
          // Online: do NOT unlock immediately. Request access telemetry, and await MQTT command.
          Serial.println("[RFID] Authorized card scanned. Delegating access authorization to AI/MFA...");
          sendTelemetry("access_request", "Valid RFID (online) - awaiting authorization", uidStr);
        }
        else
        {
          lastUser = "unknown";
          if (newStatus == 2) {
            sendTelemetry("access_denied", "Blocked RFID card (online)", uidStr);
          } else {
            sendTelemetry("access_denied", "Pending RFID card (online)", uidStr);
          }
        }
      }
      else
      {
        Serial.println("[RFID] API status not fetched. Querying local Preferences cache...");
        // FALLBACK: Offline or server error. Read from local preferences.
        preferences.begin("cards", true);
        int localStatus = preferences.getInt(uidStr.c_str(), -1);
        preferences.end();
        Serial.printf("[RFID] Loaded cache for UID=%s: status=%d\n", uidStr.c_str(), localStatus);

        if (localStatus == -1)
        {
          Serial.println("[RFID] Card never seen offline. Saving local status as Pending (0).");
          // First time this card is seen and we are offline. Store it as pending.
          preferences.begin("cards", false);
          preferences.putInt(uidStr.c_str(), 0); // Store as pending
          preferences.end();

          Serial.println("Offline: New card detected. Cached locally as pending.");
          shouldUnlock = false;
        }
        else if (localStatus == 1)
        {
          // Accepted card from local cache
          Serial.println("Offline: Accepted card matched from local cache.");
          shouldUnlock = true;
        }
        else
        {
          // Blocked or Pending card
          Serial.print("Offline: Card not authorized. Local status: ");
          Serial.println(localStatus);
          shouldUnlock = false;
        }

        if (shouldUnlock)
        {
          lastUser = uidStr;
          updateLockState(false, lastUser);
          sendTelemetry("access_granted", "Valid RFID (offline cache)", uidStr);
        }
        else
        {
          lastUser = "unknown";
          sendTelemetry("access_denied", "Unauthorized RFID (offline cache)", uidStr);
        }
      }
    }
  }

  unsigned long currentMillis = millis();
  if (currentMillis - lastSensorCheck >= 1000) {
    lastSensorCheck = currentMillis;

    // Components
    ultrassonic.update();
    ldr.update();
    rfid.update();

    Serial.print("[Status] Uptime: ");
    Serial.print(millis() / 1000);
    Serial.print("s | Distance: ");
    Serial.print(ultrassonic.distance());
    Serial.print(" cm | Light: ");
    Serial.print(ldr.lightLevel());
    Serial.print(" | RFID: ");
    if (rfid.isConnected()) {
      Serial.print("CONNECTED (v0x");
      Serial.print(rfid.getVersion(), HEX);
      Serial.print(")");
    } else {
      Serial.print("DISCONNECTED");
    }
    Serial.print(" | Lock: ");
    Serial.println(isLocked ? "LOCKED" : "UNLOCKED");
  }

  delay(1);
}
