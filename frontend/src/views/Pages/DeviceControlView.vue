<template>
  <AdminLayout>
    <div class="mx-auto max-w-7xl">
      <PageBreadcrumb :pageTitle="currentPageTitle" />
      
      <Device class="mt-4" />
    </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import AdminLayout from '@/components/layout/AdminLayout.vue'
import Device from '@/components/device/Device.vue'
import PageBreadcrumb from '@/components/common/PageBreadcrumb.vue'

const currentPageTitle = 'Device Control'
</script>