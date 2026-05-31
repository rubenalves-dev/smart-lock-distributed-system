<template>
  <div class="w-full px-4 py-6 md:px-8 space-y-10">
    
    <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-gray-800 dark:bg-white/[0.03]">
      <div class="max-w-full overflow-x-auto custom-scrollbar">
        <table class="w-full table-auto">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-700">
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Serviço / Componente</th>
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Tipo de Infraestrutura</th>
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Latência de Resposta</th>
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Estado de Conexão</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-if="loading">
              <td colspan="4" class="px-5 py-6 text-center text-gray-500 dark:text-gray-400">
                A verificar conectividade dos microsserviços...
              </td>
            </tr>
            <tr v-else v-for="(info, serviceName) in services" :key="serviceName">
              <td class="px-5 py-4 whitespace-nowrap">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 rounded-full bg-primary/10 flex items-center justify-center text-primary dark:text-blue-400">
                    <span class="text-xs font-bold font-mono">{{ serviceName.substring(0, 3).toUpperCase() }}</span>
                  </div>
                  <span class="font-medium text-theme-sm text-gray-800 dark:text-white/90 capitalize">
                    {{ info.service }}
                  </span>
                </div>
              </td>
              <td class="px-5 py-4 text-gray-500 text-theme-sm dark:text-gray-400">
                <span class="px-2 py-0.5 rounded bg-gray-100 dark:bg-gray-800 font-mono text-xs">
                  {{ getServiceType(info.service) }}
                </span>
              </td>
              <td class="px-5 py-4 text-gray-500 text-theme-sm dark:text-gray-400 font-mono">
                {{ info.online ? `${info.latency_ms} ms` : '—' }}
              </td>
              <td class="px-5 py-4">
                <span :class="['rounded-full px-3 py-1 text-theme-xs font-medium inline-block', 
                  info.online ? 'bg-success-50 text-success-700 dark:bg-success-500/15' : 'bg-error-50 text-error-700 dark:bg-error-500/15']">
                  {{ info.online ? 'Online' : 'Offline' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

   
    <div class="w-full flex flex-col items-start">
      <PageBreadcrumb :pageTitle="currentPageTitle2" class="!max-w-none w-full px-0" />
      <div class="mt-6 w-full">
        <ComponentCard class="!max-w-none w-full">
          <UptimeChart />
        </ComponentCard>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import ComponentCard from "@/components/common/ComponentCard.vue";
import PageBreadcrumb from "@/components/common/PageBreadcrumb.vue";
import UptimeChart from "@/components/charts/LineChart/UptimeChart.vue";

const currentPageTitle2 = ref("Uptime dos Serviços (%)");

const services = ref({})
const loading = ref(true)
let intervalId = null

const fetchHealth = async () => {
  try {
    const response = await fetch('https://smartlock-api.raiiaa.dev/api/health')
    if (response.ok) {
      services.value = await response.json()
    }
  } catch (error) {
    console.error('Erro ao interrogar api/health:', error)
  } finally {
    loading.value = false
  }
}

const getServiceType = (service) => {
  if (service === 'postgres') return 'Relational DB'
  if (service === 'influxdb') return 'Time-Series DB'
  if (service === 'rabbitmq') return 'gRPC/AMQP Broker'
  if (service === 'mosquitto') return 'MQTT Broker'
  return 'Microservice'
}

onMounted(() => {
  fetchHealth()
  // Faz polling do estado a cada 5 segundos para atualizar as badges dinamicamente
  intervalId = setInterval(fetchHealth, 5000)
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
})
</script>