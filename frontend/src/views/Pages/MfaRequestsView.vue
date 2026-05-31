<template>
  <AdminLayout>
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between mb-4">
      <PageBreadcrumb pageTitle="Aprovações de Acesso (MFA)" />
      
      <!-- WebSocket Status Indicator -->
      <div class="flex items-center gap-2 px-3 py-1.5 rounded-full text-xs font-semibold bg-gray-100 dark:bg-white/[0.03] border border-gray-200 dark:border-gray-800">
        <span class="relative flex h-2 w-2">
          <span :class="wsConnected ? 'bg-success-500' : 'bg-warning-500 animate-ping'" class="absolute inline-flex h-full w-full rounded-full opacity-75"></span>
          <span :class="wsConnected ? 'bg-success-500' : 'bg-warning-500'" class="relative inline-flex rounded-full h-2 w-2"></span>
        </span>
        <span class="text-gray-600 dark:text-gray-400">
          {{ wsConnected ? 'Real-time: Ligado' : 'Real-time: A ligar...' }}
        </span>
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="error" class="mb-6 p-4 rounded-xl border border-error-200 bg-error-50 text-error-700 dark:border-error-500/20 dark:bg-error-500/10 dark:text-error-400 flex items-center gap-3">
      <span>⚠️</span>
      <span class="text-sm font-medium">{{ error }}</span>
    </div>

    <div class="space-y-8">
      <!-- Section 1: Pending Approvals Grid -->
      <ComponentCard title="Pedidos de Autenticação Pendentes" desc="Scans de cartões RFID classificados pela IA como suspeitos ou críticos que requerem validação manual.">
        
        <!-- Loading State -->
        <div v-if="loading && pendingRequests.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-400">
          <svg class="animate-spin h-8 w-8 mb-3 text-brand-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span class="text-sm">A carregar pedidos de autenticação...</span>
        </div>

        <!-- Blank Empty State -->
        <div v-else-if="pendingRequests.length === 0" class="flex flex-col items-center justify-center py-16 text-center">
          <div class="h-16 w-16 rounded-full bg-success-50 dark:bg-success-500/10 border border-success-100 dark:border-success-500/20 flex items-center justify-center text-success-500 dark:text-success-400 text-3xl mb-4">
            🛡️
          </div>
          <h3 class="text-base font-semibold text-gray-800 dark:text-white/90 mb-1">Sem Pedidos Pendentes</h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 max-w-sm">
            Nenhuma tentativa de acesso recente ativou o alerta da IA. O sistema está seguro e a funcionar normalmente.
          </p>
        </div>

        <!-- Cards Grid -->
        <div v-else class="grid grid-cols-1 gap-6 sm:grid-cols-2 xl:grid-cols-3">
          <div v-for="req in pendingRequests" :key="req.id" 
               class="flex flex-col rounded-2xl border border-gray-250 bg-white p-5 shadow-sm hover:shadow-md transition dark:border-gray-850 dark:bg-gray-dark-2">
            
            <!-- Card Header: User and Status -->
            <div class="flex items-start justify-between mb-4">
              <div>
                <h4 class="text-sm font-bold text-gray-800 dark:text-white/90 truncate max-w-[180px]">
                  {{ req.user_name || 'Utilizador Desconhecido' }}
                </h4>
                <p class="text-[11px] text-gray-400 dark:text-gray-500 font-mono mt-0.5">{{ req.rfid_uid }}</p>
              </div>
              <span :class="getSeverityBadgeClass(req.classification)" class="px-2 py-0.5 text-[10px] font-bold rounded uppercase">
                {{ getSeverityName(req.classification) }}
              </span>
            </div>

            <!-- Details List -->
            <div class="grid grid-cols-3 gap-2 py-3 px-3 bg-gray-50 dark:bg-white/[0.02] border border-gray-150 dark:border-gray-800 rounded-xl text-center mb-4">
              <div>
                <div class="text-[9px] font-medium text-gray-400 dark:text-gray-500 uppercase">Falhas</div>
                <div class="text-sm font-bold text-gray-800 dark:text-white/90 mt-0.5">{{ req.fails }}</div>
              </div>
              <div>
                <div class="text-[9px] font-medium text-gray-400 dark:text-gray-500 uppercase">Distância</div>
                <div class="text-sm font-bold text-gray-800 dark:text-white/90 mt-0.5">{{ req.distance_cm.toFixed(0) }} cm</div>
              </div>
              <div>
                <div class="text-[9px] font-medium text-gray-400 dark:text-gray-500 uppercase">Luminosidade</div>
                <div class="text-sm font-bold text-gray-800 dark:text-white/90 mt-0.5">{{ req.light_level }} lx</div>
              </div>
            </div>

            <!-- AI Recommendation & Confidence -->
            <div class="space-y-3 mb-5 grow">
              <div>
                <div class="flex justify-between items-center text-[10px] text-gray-400 dark:text-gray-500 mb-1">
                  <span>Confiança da IA</span>
                  <span class="font-bold text-gray-700 dark:text-white/80">{{ (req.confidence * 100).toFixed(0) }}%</span>
                </div>
                <div class="w-full bg-gray-100 dark:bg-gray-800 rounded-full h-1.5 overflow-hidden">
                  <div :style="{ width: `${req.confidence * 100}%` }" 
                       :class="req.classification === 3 ? 'bg-error-500' : 'bg-warning-500'" 
                       class="h-full rounded-full transition-all duration-500"></div>
                </div>
              </div>
              
              <div :class="req.classification === 3 ? 'bg-error-50 border-error-100 text-error-700 dark:bg-error-500/10 dark:border-error-500/20 dark:text-error-400' : 'bg-warning-50 border-warning-100 text-warning-700 dark:bg-warning-500/10 dark:border-warning-500/20 dark:text-warning-400'" 
                   class="p-2.5 rounded-lg border text-[11px] leading-relaxed font-medium">
                <strong>Recomendação:</strong> {{ req.recommendation }}
              </div>
            </div>

            <!-- Action Buttons -->
            <div class="grid grid-cols-2 gap-3 pt-3 border-t border-gray-100 dark:border-gray-800 mt-auto">
              <button @click="handleReject(req.id)" :disabled="actionLoading[req.id]"
                      class="flex items-center justify-center gap-1.5 py-2 px-3 rounded-lg border border-error-200 text-error-600 hover:bg-error-50 dark:border-error-500/20 dark:text-error-400 dark:hover:bg-error-500/10 text-xs font-semibold transition disabled:opacity-50">
                <span>❌</span>
                Bloquear
              </button>
              <button @click="handleApprove(req.id)" :disabled="actionLoading[req.id]"
                      class="flex items-center justify-center gap-1.5 py-2 px-3 rounded-lg bg-success-500 text-white hover:bg-success-600 text-xs font-semibold transition shadow-sm disabled:opacity-50">
                <span>🔑</span>
                Autorizar
              </button>
            </div>

          </div>
        </div>
      </ComponentCard>

      <!-- Section 2: Resolution History Table -->
      <ComponentCard title="Histórico de Resoluções MFA" desc="Registo histórico dos pedidos avaliados pela IA e a respetiva decisão administrativa.">
        
        <div v-if="resolvedRequests.length === 0" class="py-8 text-center text-xs text-gray-500 dark:text-gray-400">
          Nenhum registo resolvido encontrado no histórico.
        </div>
        
        <div v-else class="overflow-x-auto custom-scrollbar">
          <table class="min-w-full divide-y divide-gray-100 dark:divide-gray-800 text-left">
            <thead>
              <tr class="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-wider font-bold">
                <th class="py-3 px-4">Data/Hora</th>
                <th class="py-3 px-4">Utilizador</th>
                <th class="py-3 px-4">RFID UID</th>
                <th class="py-3 px-4">Nível Risco</th>
                <th class="py-3 px-4">Confiança</th>
                <th class="py-3 px-4">Status Decisão</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-150 dark:divide-gray-850 text-xs text-gray-600 dark:text-gray-300 font-medium">
              <tr v-for="req in resolvedRequests" :key="req.id" class="hover:bg-gray-50/50 dark:hover:bg-white/[0.01]">
                <td class="py-3 px-4 whitespace-nowrap text-[11px] text-gray-400 dark:text-gray-500">
                  {{ formatDate(req.created_at) }}
                </td>
                <td class="py-3 px-4 font-bold text-gray-800 dark:text-white/90">
                  {{ req.user_name || 'Desconhecido' }}
                </td>
                <td class="py-3 px-4 font-mono text-[11px] text-gray-400 dark:text-gray-500">
                  {{ req.rfid_uid }}
                </td>
                <td class="py-3 px-4 whitespace-nowrap">
                  <span :class="getSeverityBadgeClass(req.classification)" class="px-1.5 py-0.5 text-[9px] font-bold rounded uppercase">
                    {{ getSeverityName(req.classification) }}
                  </span>
                </td>
                <td class="py-3 px-4 whitespace-nowrap text-gray-800 dark:text-white/90 font-semibold">
                  {{ (req.confidence * 100).toFixed(0) }}%
                </td>
                <td class="py-3 px-4 whitespace-nowrap">
                  <span :class="req.status === 'approved' ? 'text-success-600 bg-success-50 dark:text-success-400 dark:bg-success-500/10' : 'text-error-600 bg-error-50 dark:text-error-400 dark:bg-error-500/10'" 
                        class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] uppercase font-bold border border-current/10">
                    <span class="w-1 h-1 rounded-full bg-current"></span>
                    {{ req.status === 'approved' ? 'Autorizado' : 'Bloqueado' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </ComponentCard>
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue';
import { useMfa } from '@/composables/useMfa';
import AdminLayout from '@/components/layout/AdminLayout.vue';
import PageBreadcrumb from '@/components/common/PageBreadcrumb.vue';
import ComponentCard from '@/components/common/ComponentCard.vue';

const { requests, loading, error, wsConnected, fetchRequests, approveRequest, rejectRequest, connectWebSocket } = useMfa();

const actionLoading = ref<Record<number, boolean>>({});

// Categorize requests into pending and resolved
const pendingRequests = computed(() => {
  return requests.value.filter(r => r.status === 'pending');
});

const resolvedRequests = computed(() => {
  return requests.value.filter(r => r.status !== 'pending');
});

const handleApprove = async (id: number) => {
  actionLoading.value[id] = true;
  const success = await approveRequest(id);
  actionLoading.value[id] = false;
  if (!success) {
    alert('Erro ao aprovar o acesso. Tente novamente.');
  }
};

const handleReject = async (id: number) => {
  actionLoading.value[id] = true;
  const success = await rejectRequest(id);
  actionLoading.value[id] = false;
  if (!success) {
    alert('Erro ao recusar/bloquear o acesso. Tente novamente.');
  }
};

const getSeverityName = (val: number): string => {
  switch (val) {
    case 0: return 'Normal';
    case 1: return 'Irregular';
    case 2: return 'Suspeito';
    case 3: return 'Crítico';
    default: return 'Desconhecido';
  }
};

const getSeverityBadgeClass = (val: number): string => {
  switch (val) {
    case 0: return 'bg-success-100 text-success-700 dark:bg-success-500/10 dark:text-success-400 border border-success-200 dark:border-success-500/20';
    case 1: return 'bg-blue-100 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400 border border-blue-200 dark:border-blue-500/20';
    case 2: return 'bg-warning-100 text-warning-700 dark:bg-warning-500/10 dark:text-warning-400 border border-warning-200 dark:border-warning-500/20';
    case 3: return 'bg-error-100 text-error-700 dark:bg-error-500/10 dark:text-error-400 border border-error-200 dark:border-error-500/20 animate-pulse';
    default: return 'bg-gray-100 text-gray-700 dark:bg-white/5 dark:text-gray-400 border border-gray-200 dark:border-white/10';
  }
};

const formatDate = (dateStr: string): string => {
  try {
    const d = new Date(dateStr);
    return d.toLocaleString('pt-PT', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' });
  } catch {
    return dateStr;
  }
};

onMounted(() => {
  fetchRequests();
  connectWebSocket();
});
</script>
