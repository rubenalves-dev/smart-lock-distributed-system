<script setup lang="ts">
import { ref } from 'vue';

// Estado do dispositivo
const isUnlocked = ref(false);
const isLoading = ref(false);

// Função para alternar o estado da porta
const toggleDoor = async () => {
  isLoading.value = true;
  const API_URL = '/api/device/door';

  try {
    const response = await fetch(API_URL, {
      method: 'POST',
      mode: 'cors',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        action: isUnlocked.value ? 'lock' : 'unlock',
        timestamp: new Date().toISOString(),
      }),
    });

    if (!response.ok) {
      throw new Error('Falha na comunicação');
    }

    // Se o pedido for bem sucedido, inverte o estado
    isUnlocked.value = !isUnlocked.value;
  } catch (error) {
    console.error('Erro:', error);
    alert('Erro ao comunicar com o dispositivo.');
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="p-6 bg-white rounded-lg shadow">
    <h2 class="text-xl font-bold mb-4">Controlo Remoto do Dispositivo</h2>
    
    <div class="mb-4">
      <p>Estado do Sistema:</p>
      <span :class="isUnlocked ? 'text-green-600' : 'text-red-600'" class="font-bold">
        Porta: {{ isUnlocked ? 'ABERTA (Destrancada)' : 'FECHADA (Trancada)' }}
      </span>
    </div>

    <button
      @click="toggleDoor"
      :disabled="isLoading"
      class="px-4 py-2 rounded text-white font-bold"
      :class="isUnlocked ? 'bg-red-500' : 'bg-green-600'"
    >
      {{ isLoading ? 'A processar...' : (isUnlocked ? 'Trancar Porta' : 'Destrancar Porta') }}
    </button>
  </div>
</template>