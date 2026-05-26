const char INDEX_HTML[] = R"rawliteral(
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Biometric Smart Lock Dashboard</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-grad-start: #0f172a;
            --bg-grad-end: #1e293b;
            --glass-bg: rgba(30, 41, 59, 0.7);
            --glass-border: rgba(255, 255, 255, 0.08);
            --text-primary: #f8fafc;
            --text-secondary: #94a3b8;
            --color-locked: #ef4444;
            --color-unlocked: #10b981;
            --color-primary: #3b82f6;
            --color-primary-hover: #2563eb;
            --glow-locked: rgba(239, 68, 68, 0.4);
            --glow-unlocked: rgba(16, 185, 129, 0.4);
            --glow-primary: rgba(59, 130, 246, 0.3);
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
            font-family: 'Plus Jakarta Sans', sans-serif;
            -webkit-tap-highlight-color: transparent;
        }

        body {
            background: linear-gradient(135deg, var(--bg-grad-start), var(--bg-grad-end));
            color: var(--text-primary);
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 20px;
            overflow-x: hidden;
        }

        .container {
            width: 100%;
            max-width: 440px;
            display: flex;
            flex-direction: column;
            gap: 20px;
        }

        .card {
            background: var(--glass-bg);
            backdrop-filter: blur(16px);
            -webkit-backdrop-filter: blur(16px);
            border: 1px solid var(--glass-border);
            border-radius: 24px;
            padding: 24px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
            text-align: center;
            transition: transform 0.3s ease, box-shadow 0.3s ease;
        }

        .header-section h1 {
            font-size: 1.5rem;
            font-weight: 700;
            letter-spacing: -0.5px;
            background: linear-gradient(to right, #60a5fa, #a78bfa);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 4px;
        }

        .header-section p {
            font-size: 0.85rem;
            color: var(--text-secondary);
        }

        .status-container {
            margin: 20px 0;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 12px;
        }

        .status-ring {
            width: 140px;
            height: 140px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            position: relative;
            background: rgba(15, 23, 42, 0.4);
            border: 4px solid var(--glass-border);
            transition: border-color 0.5s ease, box-shadow 0.5s ease;
        }

        .status-ring::before {
            content: '';
            position: absolute;
            width: calc(100% + 12px);
            height: calc(100% + 12px);
            border-radius: 50%;
            border: 2px dashed rgba(255, 255, 255, 0.15);
            animation: rotate 20s linear infinite;
        }

        @keyframes rotate {
            from { transform: rotate(0deg); }
            to { transform: rotate(360deg); }
        }

        .status-icon {
            font-size: 3rem;
            transition: transform 0.3s ease;
        }

        .LOCKED .status-ring {
            border-color: var(--color-locked);
            box-shadow: 0 0 25px var(--glow-locked);
        }

        .UNLOCKED .status-ring {
            border-color: var(--color-unlocked);
            box-shadow: 0 0 25px var(--glow-unlocked);
        }

        .status-badge {
            font-size: 1.25rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 1px;
            padding: 6px 16px;
            border-radius: 50px;
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid var(--glass-border);
            transition: color 0.3s ease;
        }

        .LOCKED .status-badge {
            color: var(--color-locked);
        }

        .UNLOCKED .status-badge {
            color: var(--color-unlocked);
        }

        .btn {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 100%;
            padding: 16px;
            border-radius: 16px;
            font-size: 1rem;
            font-weight: 600;
            border: none;
            cursor: pointer;
            transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
            color: var(--text-primary);
        }

        .btn-open {
            background: linear-gradient(135deg, #3b82f6, #6366f1);
            box-shadow: 0 8px 20px var(--glow-primary);
        }

        .btn-open:hover {
            transform: translateY(-2px);
            box-shadow: 0 12px 24px rgba(99, 102, 241, 0.45);
        }

        .btn-open:active {
            transform: translateY(1px);
        }

        .btn-open:disabled {
            background: rgba(255, 255, 255, 0.08);
            color: var(--text-secondary);
            cursor: not-allowed;
            box-shadow: none;
            transform: none;
        }

        .panel-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;
            border-bottom: 1px solid var(--glass-border);
            padding-bottom: 10px;
        }

        .panel-header h3 {
            font-size: 0.95rem;
            font-weight: 600;
            color: var(--text-primary);
        }

        .btn-text {
            background: none;
            border: none;
            color: var(--color-primary);
            font-size: 0.8rem;
            font-weight: 600;
            cursor: pointer;
            transition: color 0.2s ease;
        }

        .btn-text:hover {
            color: var(--color-primary-hover);
        }

        .service-list, .user-list {
            display: flex;
            flex-direction: column;
            gap: 10px;
            text-align: left;
            max-height: 160px;
            overflow-y: auto;
            padding-right: 4px;
        }

        .service-list::-webkit-scrollbar, .user-list::-webkit-scrollbar {
            width: 4px;
        }

        .service-list::-webkit-scrollbar-thumb, .user-list::-webkit-scrollbar-thumb {
            background: var(--glass-border);
            border-radius: 4px;
        }

        .service-item, .user-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 10px 14px;
            border-radius: 12px;
            background: rgba(15, 23, 42, 0.2);
            border: 1px solid rgba(255, 255, 255, 0.03);
            font-size: 0.85rem;
        }

        .user-item {
            cursor: pointer;
            transition: background-color 0.2s ease, transform 0.2s ease;
        }

        .user-item:hover {
            background: rgba(255, 255, 255, 0.05);
            transform: translateX(2px);
        }

        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            display: inline-block;
        }

        .status-dot.online {
            background-color: var(--color-unlocked);
            box-shadow: 0 0 8px var(--color-unlocked);
        }

        .status-dot.offline {
            background-color: var(--color-locked);
            box-shadow: 0 0 8px var(--color-locked);
        }

        .status-value {
            font-size: 0.75rem;
            color: var(--text-secondary);
        }

        .footer-link {
            font-size: 0.85rem;
            color: var(--text-secondary);
            text-decoration: none;
            transition: color 0.2s ease;
            margin-top: 10px;
            display: inline-block;
        }

        .footer-link:hover {
            color: var(--color-primary);
            text-decoration: underline;
        }

        /* Modal styling */
        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(15, 23, 42, 0.8);
            backdrop-filter: blur(8px);
            -webkit-backdrop-filter: blur(8px);
            align-items: center;
            justify-content: center;
            z-index: 100;
            padding: 20px;
        }

        .modal-content {
            background: var(--bg-grad-end);
            border: 1px solid var(--glass-border);
            border-radius: 20px;
            padding: 24px;
            width: 100%;
            max-width: 360px;
            box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
            text-align: left;
        }

        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;
            border-bottom: 1px solid var(--glass-border);
            padding-bottom: 10px;
        }

        .modal-title {
            font-size: 1.1rem;
            font-weight: 700;
        }

        .modal-close {
            background: none;
            border: none;
            color: var(--text-secondary);
            font-size: 1.5rem;
            cursor: pointer;
        }

        .modal-close:hover {
            color: var(--text-primary);
        }

        .details-group {
            margin-bottom: 14px;
        }

        .details-label {
            font-size: 0.75rem;
            color: var(--text-secondary);
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-bottom: 4px;
        }

        .details-value {
            font-size: 0.95rem;
            font-weight: 500;
            color: var(--text-primary);
        }
    </style>
</head>
<body class="LOCKED" id="bodyState">

    <div class="container">

        <!-- Header -->
        <div style="text-align: center; margin-bottom: 8px;">
            <div class="header-section">
                <h1>Smart Lock Hub</h1>
                <p>Distributed IoT Controller</p>
            </div>
        </div>

        <!-- Lock Control Card -->
        <div class="card">
            <div class="status-container">
                <div class="status-ring">
                    <span class="status-icon" id="lockIcon">🔒</span>
                </div>
                <div class="status-badge" id="statusLabel">LOCKED</div>
            </div>

            <button class="btn btn-open" id="openBtn" onclick="openDoor()">OPEN LOCK</button>
        </div>

        <!-- System Health Card -->
        <div class="card">
            <div class="panel-header">
                <h3>System Health</h3>
                <button class="btn-text" onclick="checkServices()">Verify Services</button>
            </div>
            <div class="service-list" id="servicesContainer">
                <div class="service-item">
                    <span>Local Controller (ESP32)</span>
                    <span class="status-dot online"></span>
                </div>
                <div class="service-item">
                    <span>Local MQTT connection</span>
                    <span class="status-dot offline" id="dot-local_mqtt"></span>
                </div>
                <div class="service-item">
                    <span>PostgreSQL Database</span>
                    <span class="status-dot offline" id="dot-postgres"></span>
                </div>
                <div class="service-item">
                    <span>RabbitMQ Message Broker</span>
                    <span class="status-dot offline" id="dot-rabbitmq"></span>
                </div>
                <div class="service-item">
                    <span>InfluxDB Telemetry DB</span>
                    <span class="status-dot offline" id="dot-influxdb"></span>
                </div>
            </div>
        </div>

        <!-- Registered Cards Card -->
        <div class="card">
            <div class="panel-header">
                <h3>Registered UIDs</h3>
                <button class="btn-text" onclick="loadUsers()">Refresh List</button>
            </div>
            <div class="user-list" id="usersContainer">
                <div style="color: var(--text-secondary); text-align: center; font-size: 0.8rem; width: 100%; padding: 10px 0;">
                    Click refresh to load registered cards.
                </div>
            </div>
        </div>

        <!-- Settings Info / Link -->
        <div style="text-align: center;">
            <a href="/wifi" class="footer-link">Configure WiFi Connection</a>
        </div>

    </div>

    <!-- Details Modal -->
    <div class="modal" id="detailsModal">
        <div class="modal-content">
            <div class="modal-header">
                <span class="modal-title">Card Details</span>
                <button class="modal-close" onclick="closeModal()">&times;</button>
            </div>
            <div id="modalBody">
                <div class="details-group">
                    <div class="details-label">RFID UID</div>
                    <div class="details-value" id="detail-uid">-</div>
                </div>
                <div class="details-group">
                    <div class="details-label">Owner Name</div>
                    <div class="details-value" id="detail-name">-</div>
                </div>
                <div class="details-group">
                    <div class="details-label">Owner Email</div>
                    <div class="details-value" id="detail-email">-</div>
                </div>
                <div class="details-group">
                    <div class="details-label">Registered At</div>
                    <div class="details-value" id="detail-date">-</div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // Poll status periodically
        function pollStatus() {
            fetch('/status')
                .then(r => r.text())
                .then(state => {
                    updateUIState(state);
                })
                .catch(err => console.error("Error polling status:", err));
        }

        function updateUIState(state) {
            const body = document.getElementById('bodyState');
            const icon = document.getElementById('lockIcon');
            const label = document.getElementById('statusLabel');
            const btn = document.getElementById('openBtn');

            if (state.trim() === 'UNLOCKED') {
                body.className = 'UNLOCKED';
                icon.innerText = '🔓';
                label.innerText = 'OPEN';
                btn.disabled = true;
                btn.innerText = 'OPENING...';
            } else {
                body.className = 'LOCKED';
                icon.innerText = '🔒';
                label.innerText = 'LOCKED';
                btn.disabled = false;
                btn.innerText = 'OPEN LOCK';
            }
        }

        function openDoor() {
            const btn = document.getElementById('openBtn');
            btn.disabled = true;
            btn.innerText = 'SENDING...';

            fetch('/open')
                .then(r => r.text())
                .then(txt => {
                    updateUIState('UNLOCKED');
                    // Automatically expect closed state in 5 seconds to feel responsive
                    setTimeout(pollStatus, 5100);
                })
                .catch(err => {
                    alert("Failed to unlock door: " + err);
                    pollStatus();
                });
        }

        function checkServices() {
            const container = document.getElementById('servicesContainer');
            fetch('/check-services')
                .then(r => r.json())
                .then(data => {
                    // Update local MQTT status
                    updateDotStatus('local_mqtt', data.local_mqtt);

                    // Update backend services
                    if (data.backend_services) {
                        for (const [srv, details] of Object.entries(data.backend_services)) {
                            updateDotStatus(srv, details.online);
                        }
                    }
                })
                .catch(err => {
                    alert("Error verifying services: Check if backend is running.");
                    console.error(err);
                });
        }

        // Helper function for class update
        function updateDotStatus(id, online) {
            const dot = document.getElementById('dot-' + id);
            if (dot) {
                dot.className = online ? 'status-dot online' : 'status-dot offline';
            }
        }

        function loadUsers() {
            const container = document.getElementById('usersContainer');
            container.innerHTML = '<div style="color: var(--text-secondary); text-align: center; font-size: 0.8rem; width: 100%; padding: 10px 0;">Fetching users...</div>';

            fetch('/users')
                .then(r => r.json())
                .then(users => {
                    if (!users || users.length === 0) {
                        container.innerHTML = '<div style="color: var(--text-secondary); text-align: center; font-size: 0.8rem; width: 100%; padding: 10px 0;">No registered users found.</div>';
                        return;
                    }

                    container.innerHTML = '';
                    users.forEach(user => {
                        const div = document.createElement('div');
                        div.className = 'user-item';
                        const rfid = user.rfid_uid || user.uid || "Unknown UID";
                        const name = user.name || "Unassigned Tag";

                        div.innerHTML = `
                            <span>UID: <strong>${rfid}</strong></span>
                            <span class="status-value">${name}</span>
                        `;
                        div.onclick = () => showUserDetails(rfid);
                        container.appendChild(div);
                    });
                })
                .catch(err => {
                    container.innerHTML = '<div style="color: var(--text-secondary); text-align: center; font-size: 0.8rem; width: 100%; padding: 10px 0;">Error contacting backend.</div>';
                    console.error(err);
                });
        }

        function showUserDetails(uid) {
            fetch(`/user-details?uid=${encodeURIComponent(uid)}`)
                .then(r => r.json())
                .then(user => {
                    document.getElementById('detail-uid').innerText = user.rfid_uid || user.uid || uid;
                    document.getElementById('detail-name').innerText = user.name || "Unassigned / Auto-registered";
                    document.getElementById('detail-email').innerText = user.email || "No email assigned";

                    let regDate = "-";
                    if (user.created_at) {
                        regDate = new Date(user.created_at).toLocaleString();
                    } else if (user.updated_at) {
                        regDate = new Date(user.updated_at).toLocaleString();
                    }
                    document.getElementById('detail-date').innerText = regDate;

                    document.getElementById('detailsModal').style.display = 'flex';
                })
                .catch(err => {
                    alert("Error loading card details: " + err);
                });
        }

        function closeModal() {
            document.getElementById('detailsModal').style.display = 'none';
        }

        window.onclick = function(event) {
            const modal = document.getElementById('detailsModal');
            if (event.target === modal) {
                modal.style.display = "none";
            }
        }

        // Run initial status poll and check services
        pollStatus();
        checkServices();

        // Poll status every 2 seconds to reflect hardware adjustments
        setInterval(pollStatus, 2000);
    </script>
</body>
</html>
)rawliteral";
