<template>
  <AdminLayout>
    <PageBreadcrumb :page-title="currentPageTitle" />
    <div class="space-y-5 sm:space-y-6">
      <ComponentCard title="Success Alert">
        <Alert
          variant="success"
          title="Success Message"
          message="Be cautious when performing this action."
          :show-link="true"
          link-href="/"
          link-text="Learn more"
        />
      </ComponentCard>
      
      </div>
  </AdminLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import PageBreadcrumb from '@/components/common/PageBreadcrumb.vue'
import AdminLayout from '@/components/layout/AdminLayout.vue'
import ComponentCard from '@/components/common/ComponentCard.vue'
import Alert from '@/components/ui/Alert.vue'

const currentPageTitle = ref<string>('Alerts')
</script>