# 🛡️ Intelligent IoT Smart Lock

A complex distributed system for an intelligent security lock. This project integrates IoT sensors (ESP32), AI behavioral analysis (Python), a concurrent microservice backend (Go), and an administrative web interface (Vue 3).

## 🚀 Quick Start

1. **Prerequisites**: Ensure you have [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed.
2. **Setup**: Run `make proto` to generate communication code.
3. **Launch**: Run `make up` to start the backend, AI, and Broker.

## 🛠️ Tooling Requirements

Para compilar os contratos de comunicação, modificar ficheiros `.proto` ou correr o servidor de desenvolvimento do ecossistema web, a equipa deve garantir a instalação dos seguintes requisitos:

### 📑 Protocol Buffers (gRPC)

* **Protobuf Compiler**: `protoc` (Download através dos releases do GitHub)
* **Go Plugins**:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

```


* **Python Plugins**: `pip install grpcio-tools`

### 🟢 Frontend Ecossistema (Vue 3)

* **Runtime**: [Node.js (v18+)](https://nodejs.org/) & npm
* **Instalação de Dependências**: Dentro da diretoria `/frontend`, executar `npm install` para mapear o Vite, Vue Router, Tailwind CSS, ApexCharts e as bibliotecas do Vuelidate (`@vuelidate/core` e `@vuelidate/validators`).

## 📂 Repository Structure

* `/api`: The "Source of Truth" (Protobuf definitions).
* `/backend`: Go microservices & Distributed logic. *(Ref: Programação Concorrente e Distribuída)*
* `/ai-service`: Python AI behavioral classification. *(Ref: Introdução à Inteligência Artificial)*
* `/firmware`: ESP32 C++ code & Local Web UI. *(Ref: Internet das Coisas)*
* `/deployments`: Docker Compose & Environment configs.
* `/frontend`: Vue 3 (TailAdmin) SPA Dashboard Single Page Application. *(Ref: Desenvolvimento Web / Projeto)*

---

## 🧭 Frontend Integration Guide & Data Flow

Este mapa orienta a interligação da interface Vue 3 aos endpoints disponibilizados pela API do Backend em Go (`http://localhost:8080`).

```
[TailAdmin Frontend]
   ├── Dashboard principal ──────► GET /api/health (Estado dos serviços)
   │                       ──────► GET /api/metrics/health (Histórico para o gráfico)
   ├── Device Control ───────────► GET /api/telemetry/latest (Estado físico do ESP32)
   │                  ───────────► POST /api/door/unlock (Botão de abertura remota)
   └── RFID Users Manager ───────► GET /api/users (Listagem e filtragem)
                           ───────► PUT /api/users/{uid} (Formulário de ativação com Vuelidate)

```

### 📊 1. Dashboard Principal (`DashboardView.vue`)

* **Estado dos Serviços (Cards de Conetividade)**: `GET /api/health` em Polling dinâmico (`setInterval` a cada 5s). Se `.online === true`, aplicar classes Tailwind verdes (`bg-meta-3`). Se for `false`, aplicar classes vermelhas (`bg-meta-1`).
* **Gráfico de Uptime Histórico (ApexCharts)**: `GET /api/metrics/health?range=24h&interval=5m`. Mapear o array de timestamps (`ts`) para o eixo X e o `status` (1 para online, 0 para offline) para o array de dados da série do gráfico.

### 🎛️ 2. Controlo do Dispositivo (`DeviceControlView.vue`)

* **Telemetria em Tempo Real do Hardware**: `GET /api/telemetry/latest?device_id=lock-1` (Polling a cada 2s). Mapeia diretamente os campos `distance_cm` (proximidade), `rssi` (sinal Wi-Fi) e `fails` (tentativas falhadas no leitor).
* **Botão "Unlock Door" (Abertura Remota)**: Evento `@click` dispara um `POST /api/door/unlock`. Ativar estado de `loading` no botão e exibir um toast de sucesso ao receber `{ "success": true }`.

### 👥 3. Gestão de Utilizadores (`UsersView.vue`)

* **Tabela de Cartões RFID**: `GET /api/users` traz todos os cartões registados no Postgres. O filtro `GET /api/users?incomplete=true` deve ser usado para isolar cartões auto-registados pelo hardware que aguardam atribuição de Nome/Email.
* **Janela Modal de Ações**: Submete um payload `PUT /api/users/{uid}` para associar credenciais ou alterar permissões (`is_accepted`, `is_blocked`).

---

## ⚡ 4. Arquitetura do Frontend (Vue 3 & Vite)

A interface administrativa foi desenhada tirando partido das capacidades modernas do Vue 3, estruturada sob os seguintes pilares:

* **Composition API (`<script setup>`)**: Utilização de uma sintaxe mais limpa, expressiva e performática, encapsulando a reatividade através de `ref`, `computed` e ganchos de ciclo de vida (`onMounted`, `onUnmounted`).
* **Estilização Utilitária (Tailwind CSS & TailAdmin)**: Interface totalmente responsiva com suporte nativo a Modo Escuro (`dark:bg-boxdark`), permitindo alterações estéticas fluidas e isoladas diretamente na estrutura de tags do componente.
* **Encapsulamento SPA (Single Page Application)**: O roteamento interno é inteiramente gerido no cliente pelo `vue-router`. Em ambiente de produção (Docker), o servidor HTTP está configurado com regras de fallback SPA, redirecionando qualquer variação de rota para o `index.html` estático gerado pelo empacotador Vite.

---

## 🛡️ 5. Estratégia de Validação no Frontend (`Vuelidate`)

Para garantir a integridade dos dados e uma excelente experiência de utilizador (UX), a submissão do formulário no Modal de Utilizadores é estritamente gerida pelo **Vuelidate**, impedindo o envio de dados corrompidos ou vazios para a API.

### ⚙️ Exemplo de Implementação no `<script setup>`

```javascript
import { ref } from 'vue'
import { useVuelidate } from '@vuelidate/core'
import { required, email } from '@vuelidate/validators'

const formData = ref({
  name: '',
  email: '',
  is_accepted: false,
  is_blocked: false
})

const rules = {
  name: { required },
  email: { required, email }
}

const v$ = useVuelidate(rules, formData)

const submitForm = async (uid) => {
  v$.value.$touch()
  if (v$.value.$error) return // Bloqueia o envio se houver erros locais

  try {
    const response = await fetch(`http://localhost:8080/api/users/${uid}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(formData.value)
    })
    if (response.ok) { /* Fechar modal e recarregar tabela */ }
  } catch (error) {
    console.error('Erro ao atualizar utilizador:', error)
  }
}

```

### 🎨 Feedback Visual com Tailwind

```html
<div class="mb-4.5">
  <label class="block text-black dark:text-white font-medium">Nome</label>
  <input 
    v-model="formData.name"
    type="text"
    :class="v$.name.$error ? 'border-meta-1 focus:border-meta-1' : 'border-stroke focus:border-primary'"
    class="w-full rounded border bg-transparent py-3 px-5 outline-none transition"
  />
  <span v-if="v$.name.$error" class="text-xs text-meta-1 mt-1 block">O nome é obrigatório.</span>
</div>

```

---

## 💡 Boas Práticas Técnicas (Notas de Desenvolvimento)

1. **Gestão de Estados Nulos (`v-if`)**: Inicializar objetos reativos de telemetria como `null` e envolver o HTML num `<div v-if="telemetry">` para evitar erros de leitura antes da primeira resposta HTTP.
2. **Destruição de Timers (`onUnmounted`)**: Sempre que for iniciado um `setInterval` para fazer polling nas dashboards, limpar obrigatoriamente o identificador do timer com `clearInterval(id)` dentro do ciclo de vida `onUnmounted` para prevenir fugas de memória.
3. **Justificação de Arquitetura (Para a Defesa)**: Se o júri questionar o porquê de duplicar a validação (Vuelidate no frontend vs Validação de modelos no Go), a resposta técnica correta é: *A validação no frontend (Vuelidate) otimiza a UX com feedback imediato e poupa tráfego de rede desnecessário. A validação no backend (Go) é mandatória por motivos de segurança, garantindo que o sistema permanece íntegro mesmo que a API seja atacada diretamente por ferramentas como o Postman ou cURL.*