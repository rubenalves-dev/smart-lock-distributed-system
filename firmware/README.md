# 🔌 ESP32 Firmware

Embedded C++ code for the physical lock hardware.

### 📡 Connectivity

- **Primary**: Connects to Go Backend via MQTT.
- **Secondary (Rescue)**: Hosts a local **Wi-Fi Access Point** with an HTML dashboard.

### 🔧 Components

- **MFRC522**: RFID/NFC Reader.
- **AS608**: Fingerprint Sensor.
- **SW-420**: Vibration/Tilt Sensor.

### ⚠️ Note

Update `config.h` with your local Wi-Fi credentials and the Docker Host IP address before flashing.
