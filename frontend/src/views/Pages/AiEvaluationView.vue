<template>
  <AdminLayout>
    <PageBreadcrumb pageTitle="Avaliação do Modelo de IA" />

    <div class="grid grid-cols-1 gap-6 xl:grid-cols-3">
      <!-- Left Column: Config & Confusion Matrix -->
      <div class="xl:col-span-1 space-y-6">
        <!-- Configuration Card -->
        <ComponentCard title="Configuração da Avaliação" desc="Escolha a origem dos dados e execute o processo de avaliação das métricas de performance do modelo de IA.">
          
          <!-- Mode Toggle Switcher -->
          <div class="flex rounded-lg bg-gray-100 p-1 dark:bg-white/[0.03] mb-4">
            <button type="button" @click="evaluationMode = 'file'" 
                    class="flex-1 rounded-md py-1.5 text-xs font-medium transition-all"
                    :class="evaluationMode === 'file' ? 'bg-white text-gray-800 shadow-xs dark:bg-gray-800 dark:text-white/90' : 'text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-white/90'">
              Ficheiro Local
            </button>
            <button type="button" @click="evaluationMode = 'path'" 
                    class="flex-1 rounded-md py-1.5 text-xs font-medium transition-all"
                    :class="evaluationMode === 'path' ? 'bg-white text-gray-800 shadow-xs dark:bg-gray-800 dark:text-white/90' : 'text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-white/90'">
              Caminho no Servidor
            </button>
          </div>

          <form @submit.prevent="handleEvaluate" class="space-y-4">
            <!-- Mode 1: File Upload Selector -->
            <div v-if="evaluationMode === 'file'" class="space-y-2">
              <label class="block text-xs font-medium text-gray-500 dark:text-gray-400">
                Carregar Ficheiro CSV do Dataset
              </label>
              
              <!-- Drag and Drop Dropzone -->
              <div class="relative flex flex-col items-center justify-center rounded-xl border border-dashed border-gray-300 bg-gray-50/50 p-6 text-center hover:bg-gray-50 dark:border-gray-700 dark:bg-white/[0.01] dark:hover:bg-white/[0.02] transition-colors">
                <input type="file" accept=".csv" @change="handleFileChange" class="absolute inset-0 w-full h-full opacity-0 cursor-pointer" />
                
                <div v-if="!datasetFile" class="space-y-2">
                  <div class="flex justify-center">
                    <svg class="h-8 w-8 text-gray-400 dark:text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"></path>
                    </svg>
                  </div>
                  <div class="text-xs font-medium text-gray-800 dark:text-white/90">
                    Arraste um ficheiro CSV ou <span class="text-brand-500 hover:underline">procure</span>
                  </div>
                  <div class="text-[10px] text-gray-400 dark:text-gray-500">
                    Apenas ficheiros .csv
                  </div>
                  <div class="text-[9px] text-gray-400 dark:text-gray-500 italic mt-1 font-semibold">
                    Nota: O cabeçalho deve conter: fails, distance_cm, is_denied, severity
                  </div>
                </div>

                <div v-else class="space-y-3 w-full flex flex-col items-center">
                  <div class="flex items-center gap-3 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-lg max-w-full">
                    <!-- CSV Icon status -->
                    <svg class="h-6 w-6 text-success-500 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    </svg>
                    <div class="text-left overflow-hidden">
                      <div class="text-xs font-semibold text-gray-800 dark:text-white/90 truncate max-w-[150px]">
                        {{ datasetFile.name }}
                      </div>
                      <div class="text-[10px] text-gray-400 dark:text-gray-500">
                        {{ (datasetFile.size / 1024).toFixed(1) }} KB
                      </div>
                    </div>
                    <!-- Clear button -->
                    <button type="button" @click.stop="clearFile" class="text-gray-400 hover:text-error-500 transition-colors p-1">
                      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Mode 2: File Server Path Textbox -->
            <div v-else class="space-y-2">
              <label class="block text-xs font-medium text-gray-500 dark:text-gray-400">
                Caminho do Ficheiro no Servidor (.csv)
              </label>
              <input type="text" v-model="datasetPath" placeholder="ex: /data/lock_dataset.csv" 
                     class="w-full rounded-lg border border-gray-200 dark:border-gray-800 bg-transparent px-4 py-2.5 text-sm text-gray-800 dark:text-white/90 outline-none transition focus:border-brand-500 dark:focus:border-brand-500 focus:ring-1 focus:ring-brand-500" />
            </div>

            <!-- Action submit button -->
            <button type="submit" :disabled="loading"
                    class="w-full inline-flex items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-2.5 text-sm font-medium text-white hover:bg-brand-600 transition shadow-sm disabled:bg-brand-300">
              <svg v-if="loading" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {{ loading ? 'A avaliar...' : 'Avaliar Modelo' }}
            </button>
          </form>
        </ComponentCard>

        <!-- Confusion Matrix Card -->
        <ComponentCard v-if="showResults && results && results.confusion_matrix && results.confusion_matrix.length > 0"
                       title="Matriz de Confusão"
                       desc="Distribuição das classificações reais versus as predições do modelo.">
          <div class="mt-2 space-y-4">
            <!-- Dynamic Heatmap Matrix -->
            <div class="grid gap-1 text-center" :style="{ 'grid-template-columns': `repeat(${results.confusion_matrix[0].length + 1}, minmax(0, 1fr))` }">
              <!-- Empty Top-Left -->
              <div></div>
              
              <!-- Predicted Headers -->
              <div v-for="(_, colIdx) in results.confusion_matrix[0]" :key="colIdx" 
                   class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider flex items-center justify-center">
                P. C{{ colIdx }}
              </div>
              
              <!-- Rows (Actual) -->
              <template v-for="(row, rowIdx) in results.confusion_matrix" :key="rowIdx">
                <!-- Actual Header -->
                <div class="text-[10px] font-bold text-gray-500 dark:text-gray-400 uppercase tracking-wider flex items-center justify-center">
                  R. C{{ rowIdx }}
                </div>
                
                <!-- Matrix Value Cells -->
                <div v-for="(val, colIdx) in row" :key="colIdx" 
                     class="rounded-lg p-2 text-xs font-semibold transition-all border"
                     :class="[
                       rowIdx === colIdx 
                         ? 'bg-success-50 text-success-700 border-success-200 dark:bg-success-500/10 dark:text-success-400 dark:border-success-500/20' 
                         : val > 0 
                           ? 'bg-error-50 text-error-700 border-error-200 dark:bg-error-500/10 dark:text-error-400 dark:border-error-500/20'
                           : 'bg-gray-50 text-gray-400 border-gray-100 dark:bg-white/[0.02] dark:text-gray-600 dark:border-gray-800'
                     ]">
                  {{ val }}
                </div>
              </template>
            </div>
            
            <!-- Explanatory Footnotes -->
            <div class="text-[10px] text-gray-400 dark:text-gray-500 mt-2 space-y-1 border-t border-gray-100 dark:border-gray-800 pt-3">
              <p>• <strong>R. C[N]</strong>: Classe Real N | <strong>P. C[N]</strong>: Classe Prevista N</p>
              <p>• <strong>Classes</strong>: 0: OK/Normal, 1: Irregular, 2: Suspeito, 3: Crítico</p>
            </div>
          </div>
        </ComponentCard>
      </div>

      <!-- Right Column: Metrics Cards & History -->
      <div class="xl:col-span-2 space-y-6">
        <!-- Metrics Display -->
        <div v-if="showResults && results" class="space-y-6">
          <!-- Binary Metrics -->
          <ComponentCard title="Métricas de Deteção Binária" desc="Performance na distinção entre acessos normais (OK) vs anomalias de segurança.">
            <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
              <!-- Accuracy -->
              <div class="p-4 bg-brand-50 dark:bg-brand-500/10 border border-brand-100 dark:border-brand-500/20 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-brand-600 dark:text-brand-400 uppercase tracking-wider">Accuracy</div>
                <div class="mt-1 text-2xl font-bold text-brand-700 dark:text-brand-300">
                  {{ results.binary_metrics.accuracy.toFixed(1) }}%
                </div>
              </div>
              
              <!-- Precision -->
              <div class="p-4 bg-success-50 dark:bg-success-500/10 border border-success-100 dark:border-success-500/20 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-success-600 dark:text-success-400 uppercase tracking-wider">Precision</div>
                <div class="mt-1 text-2xl font-bold text-success-700 dark:text-success-300">
                  {{ results.binary_metrics.precision.toFixed(1) }}%
                </div>
              </div>

              <!-- Recall -->
              <div class="p-4 bg-warning-50 dark:bg-warning-500/10 border border-warning-100 dark:border-warning-500/20 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-warning-600 dark:text-warning-400 uppercase tracking-wider">Recall</div>
                <div class="mt-1 text-2xl font-bold text-warning-700 dark:text-warning-300">
                  {{ results.binary_metrics.recall.toFixed(1) }}%
                </div>
              </div>

              <!-- F1 Score -->
              <div class="p-4 bg-blue-50 dark:bg-blue-500/10 border border-blue-100 dark:border-blue-500/20 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-blue-600 dark:text-blue-400 uppercase tracking-wider">F1 Score</div>
                <div class="mt-1 text-2xl font-bold text-blue-700 dark:text-blue-300">
                  {{ results.binary_metrics.f1.toFixed(1) }}%
                </div>
              </div>
            </div>
          </ComponentCard>

          <!-- Macro Metrics -->
          <ComponentCard v-if="results.metrics && results.metrics.accuracy !== undefined"
                         title="Métricas de Severidade (Multi-classe Macro)"
                         desc="Performance na identificação correta de cada um dos quatro níveis de risco do sistema.">
            <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
              <!-- Accuracy -->
              <div class="p-4 bg-gray-50 dark:bg-white/[0.02] border border-gray-100 dark:border-gray-800 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Accuracy</div>
                <div class="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
                  {{ results.metrics.accuracy.toFixed(1) }}%
                </div>
              </div>

              <!-- Precision Macro -->
              <div class="p-4 bg-gray-50 dark:bg-white/[0.02] border border-gray-100 dark:border-gray-800 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Precision (Macro)</div>
                <div class="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
                  {{ results.metrics.precision_macro.toFixed(1) }}%
                </div>
              </div>

              <!-- Recall Macro -->
              <div class="p-4 bg-gray-50 dark:bg-white/[0.02] border border-gray-100 dark:border-gray-800 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">Recall (Macro)</div>
                <div class="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
                  {{ results.metrics.recall_macro.toFixed(1) }}%
                </div>
              </div>

              <!-- F1 Score Macro -->
              <div class="p-4 bg-gray-50 dark:bg-white/[0.02] border border-gray-100 dark:border-gray-800 rounded-2xl text-center">
                <div class="text-[11px] font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">F1 (Macro)</div>
                <div class="mt-1 text-2xl font-bold text-gray-800 dark:text-white/90">
                  {{ results.metrics.f1_macro.toFixed(1) }}%
                </div>
              </div>
            </div>
          </ComponentCard>
        </div>

        <!-- History Card -->
        <ComponentCard title="Histórico de Avaliações" desc="Registo das avaliações efetuadas durante esta sessão de navegação.">
          <EvaluationHistoryTable :history="history" />
        </ComponentCard>
      </div>
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAiEvaluation, type EvaluationResponse } from '@/composables/useAiEvaluation';
import EvaluationHistoryTable, { type HistoryItem } from '@/components/common/EvaluationHistoryTable.vue';
import AdminLayout from '@/components/layout/AdminLayout.vue';
import PageBreadcrumb from '@/components/common/PageBreadcrumb.vue';
import ComponentCard from '@/components/common/ComponentCard.vue';
import { API_BASE_URL } from '@/config';

const evaluationMode = ref<'file' | 'path'>('file');
const datasetFile = ref<File | null>(null);
const datasetFileContent = ref<string>('');
const datasetPath = ref('');

const history = ref<HistoryItem[]>([]);
const showResults = ref(false);
const results = ref<EvaluationResponse | null>(null);

const { evaluate, loading } = useAiEvaluation();

const fetchEvaluationHistory = async () => {
  try {
    const res = await fetch(`${API_BASE_URL}/ai/evaluations`);
    if (res.ok) {
      const data = await res.json();
      history.value = data.map((item: any) => ({
        id: item.id,
        date: new Date(item.created_at).toLocaleTimeString('pt-PT', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
        datasetPath: item.dataset_path,
        accuracy: Number((item.accuracy * 100).toFixed(1)),
        precision: 0,
        recall: 0,
        f1: 0
      }));
    }
  } catch (err) {
    console.error('Erro ao carregar histórico:', err);
  }
};

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement;
  const file = target.files?.[0];
  if (file) {
    datasetFile.value = file;
    const reader = new FileReader();
    reader.onload = (event) => {
      datasetFileContent.value = event.target?.result as string;
    };
    reader.readAsText(file);
  } else {
    clearFile();
  }
};

const clearFile = () => {
  datasetFile.value = null;
  datasetFileContent.value = '';
};

const handleEvaluate = async () => {
  let targetInput = '';

  if (evaluationMode.value === 'file') {
    if (!datasetFileContent.value) {
      alert('Por favor, selecione um ficheiro CSV primeiro.');
      return;
    }
    targetInput = datasetFileContent.value;
  } else {
    targetInput = datasetPath.value;
  }

  const data = await evaluate(targetInput);
  
  if (data) {
    results.value = data;
    showResults.value = true;
    await fetchEvaluationHistory();
  } else {
    alert('Erro ao processar a avaliação. Verifique a estrutura do ficheiro CSV ou o caminho indicado.');
  }
};

onMounted(() => {
  fetchEvaluationHistory();
});
</script>