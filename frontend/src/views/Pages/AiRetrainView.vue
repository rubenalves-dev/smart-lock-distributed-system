<template>
  <AdminLayout>
    <PageBreadcrumb pageTitle="Retreino do Modelo de IA" />

    <div class="grid grid-cols-1 gap-6 xl:grid-cols-3">
      <!-- Left Column: Training Configuration -->
      <div class="xl:col-span-1 space-y-6">
        <ComponentCard 
          title="Configuração do Treino" 
          desc="Defina as opções de época e caminho do dataset para iniciar o processo de treino e atualização dos pesos da rede neuronal."
        >
          <!-- Mode Toggle Switcher -->
          <div class="flex rounded-lg bg-gray-100 p-1 dark:bg-white/[0.03] mb-4">
            <button type="button" @click="retrainMode = 'db'" 
                    class="flex-1 rounded-md py-1.5 text-[10px] font-semibold transition-all"
                    :class="retrainMode === 'db' ? 'bg-white text-gray-800 shadow-xs dark:bg-gray-800 dark:text-white/90' : 'text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-white/90'">
              Base Dados
            </button>
            <button type="button" @click="retrainMode = 'file'" 
                    class="flex-1 rounded-md py-1.5 text-[10px] font-semibold transition-all"
                    :class="retrainMode === 'file' ? 'bg-white text-gray-800 shadow-xs dark:bg-gray-800 dark:text-white/90' : 'text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-white/90'">
              Ficheiro Local
            </button>
            <button type="button" @click="retrainMode = 'path'" 
                    class="flex-1 rounded-md py-1.5 text-[10px] font-semibold transition-all"
                    :class="retrainMode === 'path' ? 'bg-white text-gray-800 shadow-xs dark:bg-gray-800 dark:text-white/90' : 'text-gray-500 dark:text-gray-400 hover:text-gray-800 dark:hover:text-white/90'">
              Caminho Servidor
            </button>
          </div>

          <form @submit.prevent="handleRetrain" class="space-y-5">
            <!-- Mode 1: DB Telemetry Info -->
            <div v-if="retrainMode === 'db'" class="p-3.5 bg-brand-50/50 dark:bg-brand-500/5 rounded-xl border border-brand-100 dark:border-brand-500/20 text-xs text-brand-700 dark:text-brand-300 leading-relaxed font-medium">
              O retreino será executado utilizando os registos de telemetria atualmente armazenados na base de dados PostgreSQL.
            </div>

            <!-- Mode 2: File Upload Selector -->
            <div v-else-if="retrainMode === 'file'" class="space-y-2">
              <label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                Carregar Ficheiro CSV do Dataset
              </label>
              
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
                    <button type="button" @click.stop="clearFile" class="text-gray-400 hover:text-error-500 transition-colors p-1">
                      <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- Mode 3: Server File Path Textbox -->
            <div v-else class="space-y-2">
              <label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                Caminho do Dataset (Servidor)
              </label>
              <input 
                type="text" 
                v-model="datasetPath" 
                placeholder="ex: data/sensor_events.csv" 
                class="w-full rounded-lg border border-gray-200 dark:border-gray-800 bg-transparent px-4 py-2.5 text-sm text-gray-800 dark:text-white/90 outline-none transition focus:border-brand-500 dark:focus:border-brand-500 focus:ring-1 focus:ring-brand-500" 
              />
            </div>

            <!-- Epochs Input (Slider + Number) -->
            <div class="space-y-2">
              <div class="flex justify-between items-center">
                <label class="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  Número de Épocas
                </label>
                <span class="text-sm font-bold text-brand-500">{{ epochs }}</span>
              </div>
              <div class="flex items-center gap-4">
                <input 
                  type="range" 
                  v-model.number="epochs" 
                  min="5" 
                  max="100" 
                  step="5" 
                  class="w-full h-1.5 bg-gray-200 dark:bg-gray-800 rounded-lg appearance-none cursor-pointer accent-brand-500" 
                />
                <input 
                  type="number" 
                  v-model.number="epochs" 
                  min="5" 
                  max="100" 
                  class="w-16 rounded-lg border border-gray-200 dark:border-gray-800 bg-transparent px-2 py-1.5 text-center text-sm font-semibold text-gray-800 dark:text-white/90 outline-none focus:border-brand-500" 
                />
              </div>
            </div>

            <!-- Submit Button -->
            <button 
              type="submit" 
              :disabled="loading"
              class="w-full inline-flex items-center justify-center gap-2 rounded-lg bg-brand-500 px-4 py-3 text-sm font-medium text-white hover:bg-brand-600 transition shadow-sm disabled:bg-brand-300 dark:disabled:bg-brand-500/50 cursor-pointer disabled:cursor-not-allowed"
            >
              <svg v-if="loading" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {{ loading ? 'A treinar modelo...' : 'Iniciar Retreino' }}
            </button>
          </form>
        </ComponentCard>
      </div>

      <!-- Right Column: Training Diagnostics & Performance -->
      <div class="xl:col-span-2 space-y-6">
        <!-- Default State: No results yet -->
        <div 
          v-if="!showResults" 
          class="flex flex-col items-center justify-center p-12 border border-dashed border-gray-300 dark:border-gray-800 bg-white dark:bg-gray-900 rounded-xl text-center min-h-[300px]"
        >
          <div class="p-3 bg-brand-50 dark:bg-brand-500/10 text-brand-500 rounded-full mb-4">
            <svg class="h-10 w-10 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"></path>
            </svg>
          </div>
          <h3 class="text-sm font-semibold text-gray-800 dark:text-white/90">Aguardando Execução</h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 max-w-[280px]">
            Configure os parâmetros de treino na coluna lateral e clique em "Iniciar Retreino" para ver a performance da IA.
          </p>
        </div>

        <!-- Result Container -->
        <div v-else class="space-y-6">
          <!-- Status Alert message -->
          <div 
            v-if="results"
            :class="[
              'p-4 rounded-xl border text-xs font-semibold flex items-center gap-3',
              results.success 
                ? 'bg-success-50 text-success-700 border-success-200 dark:bg-success-500/10 dark:text-success-400 dark:border-success-500/20' 
                : 'bg-error-50 text-error-700 border-error-200 dark:bg-error-500/10 dark:text-error-400 dark:border-error-500/20'
            ]"
          >
            <span class="shrink-0">
              <svg v-if="results.success" class="h-5 w-5 text-success-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <svg v-else class="h-5 w-5 text-error-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
            </span>
            <div>
              {{ results.message }}
            </div>
          </div>

          <!-- Diagnostics Display -->
          <div v-if="results && results.success && results.diagnostics" class="space-y-6">
            
            <!-- Side-by-Side Accuracy and Loss Meters -->
            <ComponentCard 
              title="Métricas do Retreino" 
              desc="Diagnóstico detalhado extraído na última época de treino no conjunto de treino e validação (split de 20%)."
            >
              <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                <!-- Accuracy Metrics -->
                <div class="p-5 border border-gray-100 dark:border-gray-800 rounded-2xl bg-gray-50/50 dark:bg-white/[0.01]">
                  <h4 class="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest mb-4">Acurácia do Modelo</h4>
                  
                  <div class="space-y-4">
                    <!-- Training Accuracy -->
                    <div>
                      <div class="flex justify-between text-xs font-semibold mb-1.5">
                        <span class="text-gray-500 dark:text-gray-400">Accuracy (Treino)</span>
                        <span class="text-gray-800 dark:text-white/90 font-bold">{{ (results.diagnostics.train_accuracy * 100).toFixed(1) }}%</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-800 rounded-full h-2">
                        <div class="bg-brand-500 h-2 rounded-full transition-all duration-500" :style="{ width: `${results.diagnostics.train_accuracy * 100}%` }"></div>
                      </div>
                    </div>

                    <!-- Validation Accuracy -->
                    <div>
                      <div class="flex justify-between text-xs font-semibold mb-1.5">
                        <span class="text-gray-500 dark:text-gray-400">Accuracy (Validação)</span>
                        <span class="text-gray-800 dark:text-white/90 font-bold">{{ (results.diagnostics.validation_accuracy * 100).toFixed(1) }}%</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-800 rounded-full h-2">
                        <div class="bg-success-500 h-2 rounded-full transition-all duration-500" :style="{ width: `${results.diagnostics.validation_accuracy * 100}%` }"></div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Loss Metrics -->
                <div class="p-5 border border-gray-100 dark:border-gray-800 rounded-2xl bg-gray-50/50 dark:bg-white/[0.01]">
                  <h4 class="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-widest mb-4">Loss (Função de Custo)</h4>
                  
                  <div class="space-y-4">
                    <!-- Training Loss -->
                    <div>
                      <div class="flex justify-between text-xs font-semibold mb-1.5">
                        <span class="text-gray-500 dark:text-gray-400">Loss (Treino)</span>
                        <span class="text-gray-800 dark:text-white/90 font-bold">{{ results.diagnostics.train_loss.toFixed(4) }}</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-800 rounded-full h-2">
                        <div class="bg-amber-500 h-2 rounded-full transition-all duration-500" :style="{ width: `${Math.min((results.diagnostics.train_loss / 2.0) * 100, 100)}%` }"></div>
                      </div>
                    </div>

                    <!-- Validation Loss -->
                    <div>
                      <div class="flex justify-between text-xs font-semibold mb-1.5">
                        <span class="text-gray-500 dark:text-gray-400">Loss (Validação)</span>
                        <span class="text-gray-800 dark:text-white/90 font-bold">{{ results.diagnostics.validation_loss.toFixed(4) }}</span>
                      </div>
                      <div class="w-full bg-gray-200 dark:bg-gray-800 rounded-full h-2">
                        <div class="bg-red-500 h-2 rounded-full transition-all duration-500" :style="{ width: `${Math.min((results.diagnostics.validation_loss / 2.0) * 100, 100)}%` }"></div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </ComponentCard>

            <!-- Fitting Diagnostics Warnings -->
            <ComponentCard 
              title="Diagnóstico de Ajuste (Fitting)" 
              desc="Verificação automática baseada no diferencial de acurácia entre os dados de validação e treino."
            >
              <!-- Overfitting Detected -->
              <div 
                v-if="results.diagnostics.overfitting_detected" 
                class="p-5 border border-error-100 dark:border-error-500/20 bg-error-50/50 dark:bg-error-500/5 rounded-2xl flex items-start gap-4"
              >
                <span class="flex items-center justify-center p-2 rounded-xl bg-error-500 text-white shrink-0">
                  <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
                  </svg>
                </span>
                <div class="space-y-1">
                  <h5 class="text-sm font-semibold text-error-800 dark:text-error-400">Sobreajuste Detetado (Overfitting)</h5>
                  <p class="text-xs text-error-600 dark:text-error-400/80 leading-relaxed">
                    A acurácia no conjunto de treino é muito superior à acurácia no conjunto de validação (diferença maior que 10%). O modelo memorizou as particularidades do treino e falhará ao classificar novos padrões IoT.
                  </p>
                  <div class="pt-2 text-[10px] text-error-500 dark:text-error-400/60 font-semibold uppercase tracking-wider">
                    Recomendação: Diminuir o número de épocas ou expandir o volume de dados do dataset.
                  </div>
                </div>
              </div>

              <!-- Underfitting Detected -->
              <div 
                v-else-if="results.diagnostics.underfitting_detected" 
                class="p-5 border border-warning-100 dark:border-warning-500/20 bg-warning-50/50 dark:bg-warning-500/5 rounded-2xl flex items-start gap-4"
              >
                <span class="flex items-center justify-center p-2 rounded-xl bg-warning-500 text-white shrink-0">
                  <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
                  </svg>
                </span>
                <div class="space-y-1">
                  <h5 class="text-sm font-semibold text-warning-800 dark:text-warning-400">Subajuste Detetado (Underfitting)</h5>
                  <p class="text-xs text-warning-600 dark:text-warning-400/80 leading-relaxed">
                    A acurácia global do modelo está abaixo do limiar aceitável de 70% em ambos os conjuntos de teste. O modelo falhou em capturar as regras de severidade e segurança implícitas.
                  </p>
                  <div class="pt-2 text-[10px] text-warning-500 dark:text-warning-400/60 font-semibold uppercase tracking-wider">
                    Recomendação: Aumentar o número de épocas de treino ou rever a qualidade das labels no dataset.
                  </div>
                </div>
              </div>

              <!-- Balanced / Good model -->
              <div 
                v-else 
                class="p-5 border border-success-100 dark:border-success-500/20 bg-success-50/50 dark:bg-success-500/5 rounded-2xl flex items-start gap-4"
              >
                <span class="flex items-center justify-center p-2 rounded-xl bg-success-500 text-white shrink-0">
                  <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                  </svg>
                </span>
                <div class="space-y-1">
                  <h5 class="text-sm font-semibold text-success-800 dark:text-success-400">Modelo Ajustado com Sucesso</h5>
                  <p class="text-xs text-success-600 dark:text-success-400/80 leading-relaxed">
                    O modelo apresenta um ótimo balanço de treino e validação, indicando uma generalização de alta confiança sem desvios significativos de subajuste ou sobreajuste. Os pesos estão atualizados.
                  </p>
                  <div class="pt-2 text-[10px] text-success-500 dark:text-success-400/60 font-semibold uppercase tracking-wider">
                    Estado: Pronto para classificar incidentes de segurança em tempo real.
                  </div>
                </div>
              </div>
            </ComponentCard>
          </div>
        </div>
      </div>
    </div>

    <!-- Row: Session Retraining History Table -->
    <div class="mt-6">
      <ComponentCard 
        title="Histórico de Retreino da Sessão" 
        desc="Registo dos treinos iniciados durante a atual sessão de navegação no painel de controlo."
      >
        <div class="overflow-x-auto">
          <table class="min-w-full text-left border-collapse">
            <thead>
              <tr class="border-b border-gray-100 dark:border-gray-800 text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">
                <th class="py-3 px-4">Hora</th>
                <th class="py-3 px-4">Dataset</th>
                <th class="py-3 px-4 text-center">Épocas</th>
                <th class="py-3 px-4 text-center">Acc. Treino</th>
                <th class="py-3 px-4 text-center">Acc. Validação</th>
                <th class="py-3 px-4 text-center">Loss Treino</th>
                <th class="py-3 px-4 text-center">Loss Validação</th>
                <th class="py-3 px-4 text-center">Ajuste (Fitting)</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-gray-800 text-xs">
              <tr 
                v-for="item in history" 
                :key="item.id" 
                class="hover:bg-gray-50 dark:hover:bg-white/[0.01] transition-colors text-gray-700 dark:text-gray-300"
              >
                <td class="py-3.5 px-4 font-medium">{{ item.time }}</td>
                <td class="py-3.5 px-4 truncate max-w-[150px] font-mono">{{ item.dataset }}</td>
                <td class="py-3.5 px-4 text-center">{{ item.epochs }}</td>
                <td class="py-3.5 px-4 text-center">{{ item.trainAcc }}%</td>
                <td class="py-3.5 px-4 text-center">{{ item.valAcc }}%</td>
                <td class="py-3.5 px-4 text-center font-mono">{{ item.trainLoss }}</td>
                <td class="py-3.5 px-4 text-center font-mono">{{ item.valLoss }}</td>
                <td class="py-3.5 px-4 text-center">
                  <span 
                    :class="[
                      'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold',
                      item.fitting === 'Overfitting' 
                        ? 'bg-error-50 text-error-700 dark:bg-error-500/10 dark:text-error-400' 
                        : item.fitting === 'Underfitting' 
                          ? 'bg-warning-50 text-warning-700 dark:bg-warning-500/10 dark:text-warning-400' 
                          : 'bg-success-50 text-success-700 dark:bg-success-500/10 dark:text-success-400'
                    ]"
                  >
                    {{ item.fitting }}
                  </span>
                </td>
              </tr>
              <tr v-if="history.length === 0">
                <td colspan="8" class="text-center py-8 text-gray-400 dark:text-gray-500">
                  Nenhum retreino efetuado nesta sessão.
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
import { ref, onMounted } from 'vue';
import { useAiRetrain, type RetrainResponse } from '@/composables/useAiRetrain';
import AdminLayout from '@/components/layout/AdminLayout.vue';
import PageBreadcrumb from '@/components/common/PageBreadcrumb.vue';
import ComponentCard from '@/components/common/ComponentCard.vue';
import { API_BASE_URL } from '@/config';

interface HistoryItem {
  id: number;
  time: string;
  dataset: string;
  epochs: number;
  trainAcc: string;
  valAcc: string;
  trainLoss: string;
  valLoss: string;
  fitting: 'Balanced' | 'Overfitting' | 'Underfitting';
}

const retrainMode = ref<'db' | 'file' | 'path'>('db');
const datasetFile = ref<File | null>(null);
const datasetFileContent = ref<string>('');
const datasetPath = ref('data/sensor_events.csv');
const epochs = ref(10);
const showResults = ref(false);
const results = ref<RetrainResponse | null>(null);
const history = ref<HistoryItem[]>([]);

const { retrain, loading } = useAiRetrain();

const fetchRetrainHistory = async () => {
  try {
    const res = await fetch(`${API_BASE_URL}/ai/retrains`);
    if (res.ok) {
      const data = await res.json();
      history.value = data.map((item: any) => {
        let fittingState: 'Balanced' | 'Overfitting' | 'Underfitting' = 'Balanced';
        if (item.overfitting_detected) {
          fittingState = 'Overfitting';
        } else if (item.underfitting_detected) {
          fittingState = 'Underfitting';
        }
        return {
          id: item.id,
          time: new Date(item.created_at).toLocaleTimeString('pt-PT', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
          dataset: item.dataset_path,
          epochs: item.epochs,
          trainAcc: (item.train_accuracy * 100).toFixed(1),
          valAcc: (item.validation_accuracy * 100).toFixed(1),
          trainLoss: item.train_loss.toFixed(4),
          valLoss: item.validation_loss.toFixed(4),
          fitting: fittingState
        };
      });
    }
  } catch (err) {
    console.error('Erro ao carregar histórico de retreino:', err);
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

const handleRetrain = async () => {
  let targetPath = '';
  if (retrainMode.value === 'db') {
    targetPath = 'database';
  } else if (retrainMode.value === 'file') {
    if (!datasetFileContent.value) {
      alert('Por favor, selecione um ficheiro CSV primeiro.');
      return;
    }
    targetPath = datasetFileContent.value;
  } else {
    if (!datasetPath.value.trim()) {
      alert('Por favor, indique o caminho do ficheiro do dataset.');
      return;
    }
    targetPath = datasetPath.value;
  }

  const response = await retrain(epochs.value, targetPath);

  if (response) {
    results.value = response;
    showResults.value = true;
    await fetchRetrainHistory();
  } else {
    alert('Erro de comunicação ao treinar o modelo. Verifique os logs e a conexão.');
  }
};

onMounted(() => {
  fetchRetrainHistory();
});
</script>
