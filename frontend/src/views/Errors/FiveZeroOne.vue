<template>
  <div
    class="relative flex flex-col items-center justify-center min-h-screen p-6 overflow-hidden z-1"
  >
    <common-grid-shape />

    <div class="mx-auto w-full max-w-[242px] text-center sm:max-w-[472px]">
      <h1
        class="mb-8 font-bold text-gray-800 text-title-md dark:text-white/90 xl:text-title-2xl tracking-widest"
      >
        Not Implemented
      </h1>

      <div class="my-6 flex justify-center text-[8rem] font-black leading-none text-gray-800 dark:text-white/90 sm:text-[11rem]">
        <div class="flex items-center tracking-tighter">
            <span>5</span>
            <svg
            class="mx-4 h-20 w-20 fill-current text-gray-800 dark:text-white/90 sm:h-28 sm:w-28"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
            >
            <path d="M22 12c0 5.523-4.477 10-10 10S2 17.523 2 12 6.477 2 12 2s10 4.477 10 10zm-2 0a8 8 0 1 0-16 0 8 8 0 0 0 16 0zm-8-3a1.5 1.5 0 1 1 0-3 1.5 1.5 0 0 1 0 3zm0 7.5a3.5 3.5 0 0 1-3.163-2h6.326a3.5 3.5 0 0 1-3.163 2z"/>
            </svg>
            <span>1</span>
        </div>
      </div>

      <p
        class="mt-4 mb-6 text-base text-gray-700 dark:text-gray-400 sm:text-lg"
      >
        Esta opção requer integração com módulos adicionais de hardware ou endpoints de API que estão atualmente no roadmap do Smart Lock.
      </p>

      <router-link
        to="/"
        class="inline-flex items-center justify-center rounded-lg border border-gray-300 bg-white px-5 py-3.5 text-sm font-medium text-gray-700 shadow-theme-xs hover:bg-gray-50 hover:text-gray-800 dark:border-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:hover:bg-white/[0.03] dark:hover:text-gray-200"
      >
        Voltar à Página Inicial
      </router-link>
    </div>

    <p
      class="absolute text-sm text-center text-gray-500 -translate-x-1/2 bottom-6 left-1/2 dark:text-gray-400"
    >
      &copy; {{ currentYear }} - Smart Lock Distributed System
    </p>
  </div>
</template>

<script setup>
import { ref } from "vue";
import CommonGridShape from "../../components/common/CommonGridShape.vue";

const currentYear = ref(new Date().getFullYear());
</script>