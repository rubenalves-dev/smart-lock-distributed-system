const char INDEX_HTML[] = R"(
<!DOCTYPE html><html>
<head>
    <title>Biometric Lock</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        html { font-family: sans-serif; text-align: center; background-color: #f0f2f5; }
        body { display: flex; flex-direction: column; align-items: center; padding-top: 20px; }
        .card { background: white; padding: 2rem; border-radius: 20px; box-shadow: 0 10px 25px rgba(0,0,0,0.1); width: 320px; }
        .status { font-size: 1.5rem; font-weight: bold; margin: 15px 0; }
        .LOCKED { color: #e74c3c; }
        .UNLOCKED { color: #2ecc71; }
        .btn { display: block; background: #3498db; color: white; text-decoration: none; 
               padding: 12px; border-radius: 10px; font-size: 1rem; margin: 10px 0; transition: 0.2s; border: none; width: 100%; cursor: pointer; }
        .btn:active { transform: scale(0.98); background: #2980b9; }
        .btn-status { background: #95a5a6; }
        .btn-user { background: #2ecc71; }
    </style>
</head>
<body>
    <div class="card">
        <h2>Smart Lock</h2>
        <div class="status LOCK_CLASS">LOCK_STATE</div>
        
        <a href="/toggle" class="btn">REMOTELY UNLOCK</a>
        
        <hr>
        
        <a href="/wifi-setup" class="btn btn-status">WIFI LOGIN</a>
        <a href="/add-user" class="btn btn-user">ADD USERS</a>
        <a href="/health" class="btn btn-status">VERIFICAR SERVIÇOS</a>
    </div>
</body>
</html>
)";