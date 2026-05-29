import EvaluationHistoryTable from '@/components/common/EvaluationHistoryTable.vue';

<template>
  <div class="history-container" style="margin-top: 30px;">
    <h3>Histórico de Avaliações</h3>
    <div style="overflow-x: auto; border: 1px solid #ddd; border-radius: 8px;">
      <table style="width: 100%; border-collapse: collapse; text-align: left;">
        <thead>
          <tr style="background-color: #f4f4f4; border-bottom: 2px solid #ddd;">
            <th style="padding: 12px;">Data/Hora</th>
            <th style="padding: 12px;">Accuracy</th>
            <th style="padding: 12px;">Precision</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in history" :key="item.id" style="border-bottom: 1px solid #eee;">
            <td style="padding: 12px;">{{ item.date }}</td>
            <td style="padding: 12px; color: #3c50e0; font-weight: bold;">{{ item.accuracy }}%</td>
            <td style="padding: 12px; color: #28a745; font-weight: bold;">{{ item.precision }}%</td>
          </tr>
          <tr v-if="history.length === 0">
            <td colspan="3" style="padding: 20px; text-align: center; color: #999;">
              Nenhuma avaliação realizada ainda.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
// Definimos a interface para garantir que o TypeScript valide os dados
interface HistoryItem {
  id: number;
  date: string;
  accuracy: number;
  precision: number;
}

// Recebemos o array 'history' como uma propriedade (prop)
defineProps<{
  history: HistoryItem[]
}>();
</script>

<style scoped>
/* Opcional: podes adicionar aqui estilos globais do teu projeto se necessário */
</style>