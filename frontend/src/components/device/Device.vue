<template>
  <div class="w-full px-4 py-6 md:px-8 space-y-10">
    
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-3 w-full">
      
      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]">
        <p class="text-sm font-medium text-gray-500 text-theme-xs dark:text-gray-400">Door Status</p>
        <div class="mt-2 flex items-center">
          <span :class="[
            'rounded-full px-3 py-1 text-theme-xs font-semibold inline-block',
            isUnlocked ? 'bg-success-50 text-success-700 dark:bg-success-500/15' : 'bg-error-50 text-error-700 dark:bg-error-500/15'
          ]">
            {{ isUnlocked ? 'Unlocked / Aberta' : 'Locked / Fechada' }}
          </span>
        </div>
      </div>

      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]">
        <p class="text-sm font-medium text-gray-500 text-theme-xs dark:text-gray-400">ESP32 Status</p>
        <div class="mt-2 flex items-center">
          <span class="rounded-full bg-success-50 px-3 py-1 text-theme-xs font-semibold text-success-700 dark:bg-success-500/15 inline-block">
            Online
          </span>
        </div>
      </div>

      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-5 dark:border-gray-800 dark:bg-white/[0.03]">
        <p class="text-sm font-medium text-gray-500 text-theme-xs dark:text-gray-400">MQTT Broker</p>
        <div class="mt-2 flex items-center">
          <span class="rounded-full bg-success-50 px-3 py-1 text-theme-xs font-semibold text-success-700 dark:bg-success-500/15 inline-block">
            Connected
          </span>
        </div>
      </div>

    </div>

    <div class="grid grid-cols-1 gap-6 xl:grid-cols-2 w-full">
      
      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-white/[0.03]">
        <h3 class="text-base font-semibold text-gray-800 dark:text-white/90 mb-4">Remote Door Control</h3>
        
        <div class="space-y-5">
          <p class="text-theme-sm text-gray-500 dark:text-gray-400">
            Current State: 
            <span class="font-bold ml-1" :class="isUnlocked ? 'text-success-700 dark:text-success-500' : 'text-error-700 dark:text-error-500'">
              {{ isUnlocked ? 'LOCKED' : 'UNLOCKED' }}
            </span>
          </p>

          <button
            @click="toggleDoor"
            :disabled="isLoading"
            class="w-full sm:w-auto px-5 py-2.5 rounded-lg text-white font-medium text-theme-sm transition-all shadow-sm focus:outline-none disabled:opacity-50"
            :class="isUnlocked ? 'bg-error-600 hover:bg-error-700' : 'bg-success-600 hover:bg-success-700'"
          >
            {{ isLoading ? 'A processar...' : isUnlocked ? 'Trancar Porta' : 'Destrancar Porta' }}
          </button>
        </div>
      </div>

      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-white/[0.03]">
        <h3 class="text-base font-semibold text-gray-800 dark:text-white/90 mb-4">Device Information</h3>
        
        <div class="space-y-3 text-theme-sm">
          <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
            <span class="text-gray-500 dark:text-gray-400">Device ID:</span>
            <span class="font-mono text-gray-800 dark:text-white/90">ESP32-SmartLock-01</span>
          </div>
          <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
            <span class="text-gray-500 dark:text-gray-400">Firmware Version:</span>
            <span class="text-gray-800 dark:text-white/90">v1.0.4 (Stable)</span>
          </div>
          <div class="flex justify-between border-b border-gray-100 pb-2 dark:border-gray-800">
            <span class="text-gray-500 dark:text-gray-400">Uptime:</span>
            <span class="text-gray-800 dark:text-white/90">4 days, 12 hours</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-500 dark:text-gray-400">IP Address:</span>
            <a href="http://smartlock.local/" target="_blank" class="text-primary-600 hover:underline dark:text-primary-400">smartlock.local</a>
          </div>
        </div>
      </div>

    </div>

    <div class="w-full flex flex-col items-start">
      <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white p-6 dark:border-gray-800 dark:bg-white/[0.03]">
        <h3 class="text-base font-semibold text-gray-800 dark:text-white/90 mb-4">ESP32 Native Web Interface</h3>
        
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
import { ref } from 'vue'
import { API_BASE_URL } from '@/config'

const isUnlocked = ref(false)
const isLoading = ref(false)

const toggleDoor = async () => {
  isLoading.value = true
  const API_URL = `${API_BASE_URL}/device/door`
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), 10000)

  try {
    const response = await fetch(API_URL, {
      method: 'POST',
      mode: 'cors',
      signal: controller.signal,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        action: isUnlocked.value ? 'lock' : 'unlock',
        timestamp: new Date().toISOString(),
      }),
    })

    if (!response.ok) {
      throw new Error('Falha na comunicação')
    }

    isUnlocked.value = !isUnlocked.value
  } catch (error) {
    console.error('Erro:', error)
    alert('Erro ao comunicar com o dispositivo.')
  } finally {
    clearTimeout(timeoutId)
    isLoading.value = false
  }
}
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