<template>
  <div class="w-full px-4 py-6 md:px-8 space-y-10">
    
    <!-- Device Selector Dropdown -->
    <div class="w-full flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-gray-150 pb-5 dark:border-gray-800">
      <div>
        <h3 class="text-lg font-bold text-gray-800 dark:text-white/90">Seleção de Dispositivo</h3>
        <p class="text-xs text-gray-500 dark:text-gray-400">Controle e visualize telemetria de um dispositivo IoT específico.</p>
      </div>
      <div class="flex items-center gap-2">
        <label class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Dispositivo:</label>
        <select v-model="selectedDevice" @change="onDeviceChange" class="rounded-lg border border-gray-300 dark:border-gray-700 bg-white dark:bg-gray-900 px-4 py-2 text-xs font-semibold text-gray-800 dark:text-white/90 outline-none transition focus:border-brand-500">
          <option v-for="dev in devices" :key="dev" :value="dev">{{ dev }}</option>
          <option v-if="devices.length === 0" value="lock-1">lock-1 (Sem telemetria DB)</option>
        </select>
      </div>
    </div>

    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3 w-full">
      
      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]">
        <p class="text-sm font-medium text-gray-500 text-theme-xs dark:text-gray-400">Estado do Trinco</p>
        <div class="mt-2 flex items-center">
          <span :class="[
            'rounded-full px-3 py-1 text-theme-xs font-semibold inline-block',
            isUnlocked ? 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-400' : 'bg-error-50 text-error-700 dark:bg-error-500/15 dark:text-error-400'
          ]">
            {{ isUnlocked ? 'Destrancada (Acesso Livre)' : 'Trancada / Segura' }}
          </span>
        </div>
      </div>

      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]">
        <p class="text-sm font-medium text-gray-500 text-theme-xs dark:text-gray-400">Conetividade Hardware</p>
        <div class="mt-2 flex items-center">
          <span :class="[
            'rounded-full px-3 py-1 text-theme-xs font-semibold inline-block',
            deviceData?.status === 'online' ? 'bg-success-50 text-success-700 dark:bg-success-500/15 dark:text-success-400' : 'bg-error-50 text-error-700 dark:bg-error-500/15 dark:text-error-400'
          ]">
            {{ deviceData?.status ? deviceData.status.toUpperCase() : 'A VERIFICAR...' }}
          </span>
        </div>
      </div>

      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]">
        <p class="text-sm font-medium text-gray-500 text-theme-xs dark:text-gray-400">Intensidade do Sinal (RSSI)</p>
        <div class="mt-2 flex items-center font-mono text-sm font-semibold text-gray-800 dark:text-white">
          <span v-if="deviceData?.rssi">{{ deviceData.rssi }} dBm</span>
          <span v-else class="text-gray-400 font-sans font-normal text-xs">Sem dados</span>
        </div>
      </div>

    </div>

    <div class="grid grid-cols-1 gap-6 xl:grid-cols-2 w-full">
      
      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-white/[0.03]">
        <h3 class="text-base font-semibold text-gray-800 dark:text-white/90 mb-4">Abertura Remota via API Gateway</h3>
        
        <div class="space-y-5">
          <p class="text-theme-sm text-gray-500 dark:text-gray-400">
            Último Evento de Rede: 
            <span class="font-mono bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded text-xs ml-1 text-primary-500">
              {{ deviceData?.event || 'Nenhum' }}
            </span>
          </p>

          <button
            @click="remoteUnlock"
            :disabled="isLoading || deviceData?.status !== 'online'"
            class="w-full sm:w-auto px-5 py-2.5 rounded-lg text-white font-medium text-theme-sm transition-all shadow-sm focus:outline-none disabled:opacity-40 disabled:cursor-not-allowed bg-brand-500 hover:bg-brand-600"
          >
            {{ isLoading ? 'A enviar comando MQTT...' : 'Enviar Comando de Abertura' }}
          </button>
        </div>
      </div>

      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-white/[0.03]">
        <h3 class="text-base font-semibold text-gray-800 dark:text-white/90 mb-4">Informação de Diagnóstico do Dispositivo</h3>
        
        <div class="space-y-3 text-theme-sm">
          <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
            <span class="text-gray-500 dark:text-gray-400">Device ID Associado:</span>
            <span class="font-mono text-gray-800 dark:text-white/90">{{ deviceData?.device_id || 'lock-1' }}</span>
          </div>
          <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
            <span class="text-gray-500 dark:text-gray-400">Sensor de Proximidade:</span>
            <span class="text-gray-800 dark:text-white/90 font-mono">{{ deviceData?.distance_cm ? `${deviceData.distance_cm} cm` : '—' }}</span>
          </div>
          <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
            <span class="text-gray-500 dark:text-gray-400">Uptime do Hardware:</span>
            <span class="text-gray-800 dark:text-white/90 font-mono">{{ deviceData?.uptime ? `${deviceData.uptime}s` : '—' }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">Tentativas Falhadas Seguidas:</span>
            <span :class="['font-bold font-mono', deviceData?.fails > 0 ? 'text-error-500 animate-pulse' : 'text-success-500']">
              {{ deviceData?.fails ?? 0 }}
            </span>
          </div>
        </div>
      </div>

    </div>

    <div class="w-full flex flex-col items-start">
      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-white/[0.03]">
        <h3 class="text-base font-semibold text-gray-800 dark:text-white/90 mb-4">ESP32 Native Web Interface (Local Network UI)</h3>
        
        <div class="w-full overflow-hidden rounded-lg border border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
          <iframe 
            src="http://smartlock.local/" 
            class="w-full h-96 border-0"
            title="ESP32 Native Interface"
            sandbox="allow-forms allow-scripts allow-same-origin"
          ></iframe>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { API_BASE_URL } from '@/config'

const deviceData = ref(null)
const isUnlocked = ref(false)
const isLoading = ref(false)
const devices = ref([])
const selectedDevice = ref('lock-1')
let telemetryInterval = null

// Buscar todos os dispositivos com telemetria no banco
const fetchDevicesList = async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/telemetry/devices`)
    if (response.ok) {
      const data = await response.json()
      devices.value = data
      if (data.length > 0) {
        selectedDevice.value = data[0]
      }
    }
  } catch (error) {
    console.error('Erro ao recolher lista de dispositivos:', error)
  }
}

// Consulta o estado mais recente de telemetria da fechadura do Postgres
const fetchDeviceTelemetry = async () => {
  if (!selectedDevice.value) return
  try {
    const response = await fetch(`${API_BASE_URL}/telemetry/latest?device_id=${encodeURIComponent(selectedDevice.value)}`)
    if (response.ok) {
      deviceData.value = await response.json()
    } else {
      deviceData.value = null
    }
  } catch (error) {
    console.error('Erro ao recolher telemetria do gateway:', error)
  }
}

// Dispara a ordem assíncrona POST de abertura remota para o broker MQTT
const remoteUnlock = async () => {
  isLoading.value = true
  try {
    const response = await fetch(`${API_BASE_URL}/door/unlock`, {
      method: 'POST',
      mode: 'cors',
      headers: { 'Content-Type': 'application/json' }
    })

    if (response.ok) {
      const data = await response.json()
      if (data.success) {
        isUnlocked.value = true
        // Mantém a indicação de aberta temporariamente por 5 segundos (UX de simulação física)
        setTimeout(() => {
          isUnlocked.value = false
        }, 5000)
      }
    } else {
      throw new Error('Erro na resposta do Servidor Go')
    }
  } catch (error) {
    console.error('Erro de rede ao destrancar:', error)
    alert('Não foi possível enviar o comando de abertura remota.')
  } finally {
    isLoading.value = false
  }
}

const onDeviceChange = () => {
  fetchDeviceTelemetry()
  if (telemetryInterval) {
    clearInterval(telemetryInterval)
    telemetryInterval = setInterval(fetchDeviceTelemetry, 3000)
  }
}

onMounted(async () => {
  await fetchDevicesList()
  await fetchDeviceTelemetry()
  // Atualiza as métricas e o sinal Wi-Fi a cada 3 segundos automaticamente
  telemetryInterval = setInterval(fetchDeviceTelemetry, 3000)
})

onUnmounted(() => {
  if (telemetryInterval) clearInterval(telemetryInterval)
})
</script>

<style scoped>
.w-full {
  width: 100% !important;
  max-width: 100% !important;
}

:deep(.rounded-xl),
:deep(.bg-white),
:deep(.dark\:bg-white\/\[0\.03\]),
:deep(.grid) {
  width: 100% !important;
  max-width: 100% !important;
}
</style>