<template>
  <admin-layout>
    <div class="mb-6 w-full">
      <ecommerce-metrics />
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
import EcommerceMetrics from '../components/ecommerce/EcommerceMetrics.vue'

export default {
  components: {
    AdminLayout,
    EcommerceMetrics,
  },
  name: 'Ecommerce',
}
</script>