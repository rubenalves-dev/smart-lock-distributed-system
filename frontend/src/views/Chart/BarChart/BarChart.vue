<template>
  <AdminLayout>
    <PageBreadcrumb :pageTitle="currentPageTitle" />
  </AdminLayout>
</template>

<script setup>
import PageBreadcrumb from "@/components/common/PageBreadcrumb.vue";
import AdminLayout from "@/components/layout/AdminLayout.vue";
import { ref } from "vue";
const currentPageTitle = ref("Bar Chart");
</script>
