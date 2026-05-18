<template>
  <div class="space-y-6">
    <div class="flex flex-wrap items-center gap-8">
      <div>
        <label
          for="checkboxActiveUser"
          class="flex items-center text-sm font-medium text-gray-700 cursor-pointer select-none dark:text-gray-400"
        >
          <div class="relative">
            <input 
              type="checkbox" 
              id="checkboxActiveUser" 
              :checked="modelValue" 
              @change="$emit('update:modelValue', $event.target.checked)"
              class="sr-only" 
            />
            <div
              :class="
                modelValue
                  ? 'border-brand-500 bg-brand-500'
                  : 'bg-transparent border-gray-300 dark:border-gray-700'
              "
              class="mr-3 flex h-5 w-5 items-center justify-center rounded-md border-[1.25px] hover:border-brand-500 dark:hover:border-brand-500"
            >
              <span :class="modelValue ? '' : 'opacity-0'">
                <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
                  <path d="M11.6666 3.5L5.24992 9.91667L2.33325 7" stroke="white" stroke-width="1.94437" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </span>
            </div>
          </div>
          Permitir acesso imediato à plataforma (Conta Ativa)
        </label>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps(['modelValue'])
defineEmits(['update:modelValue'])
</script>