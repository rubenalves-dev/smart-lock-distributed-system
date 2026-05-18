<template>
  <div class="min-h-screen xl:flex">
    <app-sidebar />
    <Backdrop />
    <div
      class="flex-1 transition-all duration-300 ease-in-out"
      :class="[isExpanded || isHovered ? 'lg:ml-[290px]' : 'lg:ml-[90px]']"
    >
      <app-header />
      <div 
        :class="[
          'p-4 md:p-6',
          route.meta.fullWidth ? 'w-full' : 'mx-auto max-w-(--breakpoint-2xl)'
        ]"
      >
        <slot></slot>
      </div>
    </div>
  </div>
</template>

<script setup>
import AppSidebar from './AppSidebar.vue'
import AppHeader from './AppHeader.vue'
import { useSidebar } from '@/composables/useSidebar'
import Backdrop from './Backdrop.vue'
import { useRoute } from 'vue-router' // <- Importamos o useRoute

const route = useRoute() // <- Ativamos o acesso à rota atual
const { isExpanded, isHovered } = useSidebar()
</script>