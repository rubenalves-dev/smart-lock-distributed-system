<template>
  <div class="w-full px-4 py-6 md:px-8 space-y-10">
    
    <div class="w-full overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-gray-800 dark:bg-white/[0.03]">
      <div class="max-w-full overflow-x-auto custom-scrollbar">
        <table class="w-full table-auto">
          <thead>
            <tr class="border-b border-gray-200 dark:border-gray-700">
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">User</th>
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Entry</th>
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Exit</th>
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Method</th>
              <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Status</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="(user, index) in users" :key="index">
              <td class="px-5 py-4 whitespace-nowrap">
                <div class="flex items-center gap-3">
                  <div class="w-10 h-10 overflow-hidden rounded-full shrink-0">
                    <img :src="user.avatar" :alt="user.name" />
                  </div>
                  <span class="font-medium text-gray-800 text-theme-sm dark:text-white/90">{{ user.name }}</span>
                </div>
              </td>
              <td class="px-5 py-4 text-gray-500 text-theme-sm dark:text-gray-400">{{ user.entry }}</td>
              <td class="px-5 py-4 text-gray-500 text-theme-sm dark:text-gray-400">{{ user.exit }}</td>
              <td class="px-5 py-4 text-gray-500 text-theme-sm dark:text-gray-400">{{ user.method }}</td>
              <td class="px-5 py-4">
                <span :class="['rounded-full px-3 py-1 text-theme-xs font-medium inline-block', 
                  user.status === 'Accepted' ? 'bg-success-50 text-success-700 dark:bg-success-500/15' : 'bg-error-50 text-error-700 dark:bg-error-500/15']">
                  {{ user.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="w-full flex flex-col items-start">
      <PageBreadcrumb :pageTitle="currentPageTitle1" class="!max-w-none w-full px-0" />
      
      <div class="mt-6 w-full">
        <ComponentCard class="!max-w-none w-full"> 
          <div class="w-full overflow-hidden">
             <BarChartOne class="w-full" /> 
          </div>
        </ComponentCard>
      </div>
    </div>


    <div class="w-full flex flex-col items-start">
      <PageBreadcrumb :pageTitle="currentPageTitle2" class="!max-w-none w-full px-0" />
      
      <div class="mt-6 w-full">
        <ComponentCard class="!max-w-none w-full">
        <!-- Aqui chamamos o novo componente que vamos criar abaixo -->
        <UptimeChart />
      </ComponentCard>
      </div>
    </div>





    
  </div>
</template>

<script setup>
import { ref } from 'vue'
import VueApexCharts from 'vue3-apexcharts'
import BarChartOne from "@/components/charts/BarChart/BarChartOne.vue";
import ComponentCard from "@/components/common/ComponentCard.vue";
import PageBreadcrumb from "@/components/common/PageBreadcrumb.vue";
const currentPageTitle1 = ref("Números de Entradas nos últimos dias");
const currentPageTitle2 = ref("Uptime dos Serviços (%)");
import UptimeChart from "@/components/charts/LineChart/UptimeChart.vue";

const users = ref([
  {
    name: 'Rúben Alves',
    avatar: '/images/user/user-17.jpg',
    entry: '10:30',
    exit: '18:00',
    method: 'RFID Card',
    status: 'Accepted',
  },
  {
    name: 'Rodrigo Ventura',
    avatar: '/images/user/user-06.jpg',
    entry: '09:30',
    exit: '20:00',
    method: 'RFID Card',
    status: 'Denied',
  },
])

const series = ref([
  {
    name: 'Sales',
    data: [6, 10, 201, 10, 15, 6, 2],
  },
])


</script>
<style scoped>
/* 1. Alvo direto na div que envolve a tabela e nos cartões de componentes */
.w-full {
  width: 100% !important;
  max-width: 100% !important;
}

/* 2. Forçar a herança profunda em todos os wrappers de cartões do template */
:deep(.rounded-xl),
:deep(.bg-white),
:deep(.dark\:bg-white\/\[0\.03\]),
:deep(.grid) {
  width: 100% !important;
  max-width: 100% !important;
}

/* 3. Obrigar os gráficos internos do ApexCharts a recalcularem e preencherem o espaço */
:deep(.apexcharts-canvas), 
:deep(.apexcharts-canvas) * {
  width: 100% !important;
}
</style>