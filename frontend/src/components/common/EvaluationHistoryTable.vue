<template>
  <div class="w-full overflow-hidden">
    <div class="max-w-full overflow-x-auto custom-scrollbar">
      <table class="w-full table-auto">
        <thead>
          <tr class="border-b border-gray-200 dark:border-gray-700">
            <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Data/Hora</th>
            <th class="px-5 py-4 text-left font-medium text-gray-500 text-theme-xs dark:text-gray-400">Dataset</th>
            <th class="px-5 py-4 text-center font-medium text-gray-500 text-theme-xs dark:text-gray-400">Accuracy</th>
            <th class="px-5 py-4 text-center font-medium text-gray-500 text-theme-xs dark:text-gray-400">Precision</th>
            <th class="px-5 py-4 text-center font-medium text-gray-500 text-theme-xs dark:text-gray-400">Recall</th>
            <th class="px-5 py-4 text-center font-medium text-gray-500 text-theme-xs dark:text-gray-400">F1 Score</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
          <tr v-for="item in history" :key="item.id" class="hover:bg-gray-50/50 dark:hover:bg-white/[0.01] transition-colors">
            <!-- Time/Date -->
            <td class="px-5 py-4 whitespace-nowrap text-theme-sm text-gray-800 dark:text-white/90 font-medium">
              {{ item.date }}
            </td>
            <!-- Dataset Path -->
            <td class="px-5 py-4 whitespace-nowrap text-theme-sm text-gray-500 dark:text-gray-400 font-mono">
              {{ item.datasetPath || '/test_data.csv' }}
            </td>
            <!-- Accuracy -->
            <td class="px-5 py-4 whitespace-nowrap text-center text-theme-sm font-semibold text-brand-500">
              {{ item.accuracy }}%
            </td>
            <!-- Precision -->
            <td class="px-5 py-4 whitespace-nowrap text-center text-theme-sm font-semibold text-success-600 dark:text-success-400">
              {{ item.precision }}%
            </td>
            <!-- Recall -->
            <td class="px-5 py-4 whitespace-nowrap text-center text-theme-sm font-semibold text-warning-600 dark:text-warning-400">
              {{ item.recall }}%
            </td>
            <!-- F1 Score -->
            <td class="px-5 py-4 whitespace-nowrap text-center text-theme-sm font-semibold text-blue-600 dark:text-blue-400">
              {{ item.f1 }}%
            </td>
          </tr>
          <tr v-if="history.length === 0">
            <td colspan="6" class="px-5 py-8 text-center text-gray-500 dark:text-gray-400 text-theme-sm">
              Nenhuma avaliação realizada ainda.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
export interface HistoryItem {
  id: number;
  date: string;
  datasetPath: string;
  accuracy: number;
  precision: number;
  recall: number;
  f1: number;
}

defineProps<{
  history: HistoryItem[]
}>();
</script>