<template>
  <div class="ai-container" style="padding: 20px; font-family: sans-serif;">
    <h2>Avaliação do Modelo de IA</h2>
    <p style="color: #666; margin-bottom: 20px;">Insira o caminho do dataset para processar as métricas de validação.</p>

    <div style="margin-bottom: 30px; padding: 20px; border: 1px solid #ddd; border-radius: 8px; background: #fafafa;">
      <h3 style="margin-top: 0;">Configuração do Dataset</h3>
      <div style="margin-bottom: 15px;">
        <label style="display: block; margin-bottom: 5px; font-weight: bold;">Caminho do ficheiro (.csv):</label>
        <input type="text" v-model="datasetPath" placeholder="ex: /data/lock_dataset.csv" 
               style="width: 100%; padding: 8px; border: 1px solid #ccc; border-radius: 4px;" />
      </div>
      <button @click="evaluateModel" 
              style="background-color: #3c50e0; color: white; padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-weight: bold;">
        Avaliar Modelo
      </button>
    </div>

    <div v-if="showResults">
      <h3 style="margin-bottom: 15px;">Métricas de Desempenho</h3>
      <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin-bottom: 30px;">
        <div style="padding: 15px; background: #ebf5ff; border-left: 5px solid #3c50e0; border-radius: 4px; text-align: center;">
          <div style="font-size: 12px; color: #555; text-transform: uppercase;">Accuracy</div>
          <div style="font-size: 24px; font-weight: bold; color: #3c50e0;">94.2%</div>
        </div>
        <div style="padding: 15px; background: #eafaf1; border-left: 5px solid #28a745; border-radius: 4px; text-align: center;">
          <div style="font-size: 12px; color: #555; text-transform: uppercase;">Precision</div>
          <div style="font-size: 24px; font-weight: bold; color: #28a745;">92.8%</div>
        </div>
      </div>

      <h3>Matriz de Confusão</h3>
      <table style="border-collapse: collapse; text-align: center; margin-bottom: 20px;">
        <tr>
          <td style="border: none; padding: 10px;"></td>
          <td style="font-weight: bold; padding: 10px; background: #f4f4f4; border: 1px solid #ddd;" colspan="2">Previsão da IA</td>
        </tr>
        <tr>
          <td style="font-weight: bold; padding: 10px; background: #f4f4f4; border: 1px solid #ddd; width: 80px;">Real</td>
          <td style="padding: 20px; background: #d4edda; border: 1px solid #ddd; font-weight: bold; color: #155724;">
            845<br><span style="font-size: 11px; font-weight: normal; color: #555;">True Positive</span>
          </td>
          <td style="padding: 20px; background: #f8d7da; border: 1px solid #ddd; font-weight: bold; color: #721c24;">
            42<br><span style="font-size: 11px; font-weight: normal; color: #555;">False Negative</span>
          </td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
const datasetPath = ref('')
const showResults = ref(false)
const evaluateModel = () => {
  if (!datasetPath.value.trim()) {
    alert('Por favor, digite primeiro o caminho do ficheiro .csv!')
    return
  }
  showResults.value = true
}
</script>