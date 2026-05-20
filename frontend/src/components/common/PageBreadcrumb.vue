<template>
  <div class="flex flex-wrap items-center justify-between gap-3 mb-6">
    <h2 class="text-xl font-semibold text-gray-800 dark:text-white/90" x-text="pageTitle">
      {{ pageTitle }}
    </h2>
    
  </div>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'

interface BreadcrumbProps {
  pageTitle: string
}

defineProps<BreadcrumbProps>()
</script>
