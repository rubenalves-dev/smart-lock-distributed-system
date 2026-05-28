const char WIFI_HTML[] = R"rawliteral(
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Biometric Smart Lock - WiFi Setup</title>
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
            --color-primary: #3b82f6;
            --color-primary-hover: #2563eb;
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
            align-items: center;
            justify-content: center;
            padding: 20px;
        }

        .container {
            width: 100%;
            max-width: 400px;
        }

        .card {
            background: var(--glass-bg);
            backdrop-filter: blur(16px);
            -webkit-backdrop-filter: blur(16px);
            border: 1px solid var(--glass-border);
            border-radius: 24px;
            padding: 30px;
            box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
            text-align: center;
        }

        .header-section {
            margin-bottom: 24px;
        }

        .header-section h1 {
            font-size: 1.4rem;
            font-weight: 700;
            letter-spacing: -0.5px;
            background: linear-gradient(to right, #60a5fa, #a78bfa);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 6px;
        }

        .header-section p {
            font-size: 0.85rem;
            color: var(--text-secondary);
            line-height: 1.4;
        }

        .form-group {
            margin-bottom: 20px;
            text-align: left;
        }

        .form-group label {
            display: block;
            font-size: 0.8rem;
            font-weight: 600;
            color: var(--text-secondary);
            margin-bottom: 8px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .form-control {
            width: 100%;
            padding: 14px 16px;
            border-radius: 12px;
            background: rgba(15, 23, 42, 0.4);
            border: 1px solid var(--glass-border);
            color: var(--text-primary);
            font-size: 0.95rem;
            transition: all 0.2s ease;
        }

        .form-control:focus {
            outline: none;
            border-color: var(--color-primary);
            box-shadow: 0 0 10px rgba(59, 130, 246, 0.2);
            background: rgba(15, 23, 42, 0.6);
        }

        .btn {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 100%;
            padding: 14px;
            border-radius: 12px;
            font-size: 0.95rem;
            font-weight: 600;
            border: none;
            cursor: pointer;
            transition: all 0.2s ease;
            color: var(--text-primary);
            margin-top: 10px;
        }

        .btn-submit {
            background: linear-gradient(135deg, #3b82f6, #6366f1);
            box-shadow: 0 8px 20px var(--glow-primary);
        }

        .btn-submit:hover {
            transform: translateY(-2px);
            box-shadow: 0 12px 24px rgba(99, 102, 241, 0.4);
        }

        .btn-submit:active {
            transform: translateY(1px);
        }

        .btn-back {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid var(--glass-border);
            color: var(--text-secondary);
            margin-top: 12px;
            text-decoration: none;
        }

        .btn-back:hover {
            background: rgba(255, 255, 255, 0.1);
            color: var(--text-primary);
        }

        /* Success Message Panel */
        .status-panel {
            display: none;
            flex-direction: column;
            align-items: center;
            gap: 16px;
            animation: fadeIn 0.4s ease forwards;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(10px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .success-icon {
            font-size: 3rem;
            color: #10b981;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0% { transform: scale(1); }
            50% { transform: scale(1.08); }
            100% { transform: scale(1); }
        }

        .status-text h2 {
            font-size: 1.2rem;
            font-weight: 700;
            margin-bottom: 8px;
        }

        .status-text p {
            font-size: 0.85rem;
            color: var(--text-secondary);
            line-height: 1.5;
        }
    </style>
</head>
<body>

    <div class="container">
        <div class="card">

            <!-- Standard Form Configuration -->
            <form id="wifiForm" onsubmit="saveConfig(event)">
                <div class="header-section">
                    <h1>WiFi Configuration</h1>
                    <p>Enter the credentials for your local network. The smart lock will save them in NVRAM (EEPROM) and reboot.</p>
                </div>

                <div class="form-group">
                    <label for="ssid">Network SSID (Name)</label>
                    <input type="text" id="ssid" class="form-control" placeholder="Enter network name" required autocomplete="off">
                </div>

                <div class="form-group">
                    <label for="password">Network Password</label>
                    <input type="password" id="password" class="form-control" placeholder="Enter network password" autocomplete="off">
                </div>

                <button type="submit" class="btn btn-submit" id="submitBtn">SAVE CREDENTIALS</button>
                <a href="/" class="btn btn-back">CANCEL & RETURN</a>
            </form>

            <!-- Success/Saving Status Panel -->
            <div class="status-panel" id="statusPanel">
                <div class="success-icon">💾</div>
                <div class="status-text">
                    <h2>Credentials Saved!</h2>
                    <p>The lock controller is storing your credentials and restarting to connect to the new network.</p>
                    <p style="margin-top: 10px; font-size: 0.75rem; color: #a78bfa;">You can now close this tab or return in a few seconds.</p>
                </div>
                <a href="/" class="btn btn-back" style="width: auto; padding: 10px 24px;">Return to Home</a>
            </div>

        </div>
    </div>

    <script>
        function saveConfig(event) {
            event.preventDefault();

            const ssid = document.getElementById('ssid').value;
            const password = document.getElementById('password').value;
            const submitBtn = document.getElementById('submitBtn');

            submitBtn.disabled = true;
            submitBtn.innerText = 'SAVING CONFIG...';

            const params = new URLSearchParams();
            params.append('ssid', ssid);
            params.append('password', password);

            fetch('/wifi-save', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                },
                body: params
            })
            .then(r => r.text())
            .then(txt => {
                document.getElementById('wifiForm').style.display = 'none';
                document.getElementById('statusPanel').style.display = 'flex';
            })
            .catch(err => {
                alert("Failed to save credentials: " + err);
                submitBtn.disabled = false;
                submitBtn.innerText = 'SAVE CREDENTIALS';
            });
        }
    </script>
</body>
</html>
)rawliteral";
