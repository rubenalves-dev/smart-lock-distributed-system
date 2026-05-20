<template>
  <div class="space-y-6">
    <div>
      <label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-400">
        Nível de Acesso (Role)
      </label>
      <div class="relative z-20 bg-transparent">
        <select
          :value="modelValue"
          @change="$emit('update:modelValue', $event.target.value)"
          class="dark:bg-dark-900 h-11 w-full appearance-none rounded-lg border border-gray-300 bg-transparent bg-none px-4 py-2.5 pr-11 text-sm text-gray-800 shadow-theme-xs focus:border-brand-300 focus:outline-hidden focus:ring-3 focus:ring-brand-500/10 dark:border-gray-700 dark:bg-gray-900 dark:text-white/90 dark:focus:border-brand-800"
        >
          <option value="" disabled>Selecione um cargo</option>
          <option value="admin">Administrador</option>
          <option value="manager">Gestor / Supervisor</option>
          <option value="user">Utilizador Comum</option>
        </select>
        <span class="absolute z-30 text-gray-700 -translate-y-1/2 pointer-events-none right-4 top-1/2 dark:text-gray-400">
          <svg class="stroke-current" width="20" height="20" viewBox="0 0 20 20" fill="none">
            <path d="M4.79175 7.396L10.0001 12.6043L15.2084 7.396" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps(['modelValue'])
defineEmits(['update:modelValue'])
</script>