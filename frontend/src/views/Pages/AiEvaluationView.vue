<template>
  <div class="ai-container" style="padding: 20px; font-family: sans-serif;">
    <h2>Avaliação do Modelo de IA</h2>
    <p style="color: #666; margin-bottom: 20px;">Insira o caminho do dataset ou os dados sintéticos para processar as métricas.</p>

    <div style="margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 8px; background: #fafafa;">
      <h3 style="margin-top: 0;">Configuração do Dataset</h3>
      <div style="margin-bottom: 15px;">
        <label style="display: block; margin-bottom: 5px; font-weight: bold;">Caminho do ficheiro (.csv):</label>
        <input type="text" v-model="datasetPath" placeholder="ex: /data/lock_dataset.csv" 
               style="width: 100%; padding: 8px; border: 1px solid #ccc; border-radius: 4px;" />
      </div>
      <button @click="handleEvaluate" :disabled="loading"
              style="background-color: #3c50e0; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-weight: bold;">
        {{ loading ? 'A avaliar...' : 'Avaliar Modelo' }}
      </button>
    </div>

    <div v-if="showResults && results">
      <h3 style="margin-bottom: 15px;">Métricas de Desempenho</h3>
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-bottom: 30px;">
        <div style="padding: 15px; background: #ebf5ff; border-left: 5px solid #3c50e0; border-radius: 4px; text-align: center;">
          <div style="font-size: 12px; color: #555; text-transform: uppercase;">Accuracy</div>
          <div style="font-size: 24px; font-weight: bold; color: #3c50e0;">{{ results.accuracy }}%</div>
        </div>
        <div style="padding: 15px; background: #eafaf1; border-left: 5px solid #28a745; border-radius: 4px; text-align: center;">
          <div style="font-size: 12px; color: #555; text-transform: uppercase;">Precision</div>
          <div style="font-size: 24px; font-weight: bold; color: #28a745;">{{ results.precision }}%</div>
        </div>
      </div>
    </div>

    <EvaluationHistoryTable v-if="history.length > 0" :history="history" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useAiEvaluation } from '@/composables/useAiEvaluation';
// Corrigido para apontar para a pasta 'common' onde guardaste o ficheiro
import EvaluationHistoryTable from '@/components/common/EvaluationHistoryTable.vue';

interface EvaluationResults {
  accuracy: number;
  precision: number;
}

interface HistoryItem {
  id: number;
  date: string;
  accuracy: number;
  precision: number;
}

const datasetPath = ref('');
const history = ref<HistoryItem[]>([]);

// Adicionadas de volta as variáveis que o template precisa para os cards
const showResults = ref(false);
const results = ref<EvaluationResults | null>(null);

const { evaluate, loading } = useAiEvaluation();

const handleEvaluate = async () => {
  const data = await evaluate(datasetPath.value);
  
  if (data && data.binary_metrics) {
    // 1. Atualiza os cards com os valores atuais
    results.value = {
      accuracy: data.binary_metrics.accuracy,
      precision: data.binary_metrics.precision
    };
    showResults.value = true;

    // 2. Adiciona a linha ao histórico da tabela
    history.value.push({
      id: Date.now(),
      date: new Date().toLocaleTimeString(),
      accuracy: data.binary_metrics.accuracy,
      precision: data.binary_metrics.precision
    });
  }
};
</script>