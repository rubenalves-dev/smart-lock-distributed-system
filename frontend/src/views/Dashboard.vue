<template>
  <admin-layout>
    <div class="mb-6 w-full">
      <dashboard-metrics />
    </div>

    <div class="grid grid-cols-12 gap-4 md:gap-6">
      <div class="col-span-12 space-y-6 xl:col-span-7">
        <monthly-target />
      </div>
      <div class="col-span-12 xl:col-span-5">
        <monthly-sale />
      </div>
    </div>
  </admin-layout>
</template>

<script>
import AdminLayout from '../components/layout/AdminLayout.vue'
import DashboardMetrics from '../components/ecommerce/DashboardMetrics.vue'

export default {
  components: {
    AdminLayout,
    DashboardMetrics,
  },
  name: 'Dashboard',
}
</script>