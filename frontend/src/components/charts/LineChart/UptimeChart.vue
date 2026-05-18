<template>
  <div class="w-full">
    <VueApexCharts
      type="area"
      height="350"
      :options="chartOptions"
      :series="series"
    />
  </div>
</template>

<script setup>
import VueApexCharts from "vue3-apexcharts";
import { ref } from "vue";

const series = ref([
  { name: 'IA', data: [100, 100, 99, 100, 100, 100, 100] },
  { name: 'Base de dados', data: [100, 98, 100, 100, 99, 100, 100] },
  { name: 'MQTT', data: [95, 92, 98, 100, 100, 97, 100] },
  { name: 'ESP32', data: [85, 80, 88, 90, 92, 85, 89] },
  { name: 'Backend', data: [100, 100, 100, 100, 100, 100, 100] }
]);

const chartOptions = ref({
  colors: ['#465fff', '#33D1FF', '#10B981', '#FF9C55', '#FF4560'],
  chart: {
    fontFamily: 'Outfit, sans-serif',
    type: 'area',
    toolbar: { show: false },
  },
  stroke: { curve: 'smooth', width: 2 },
  fill: {
    type: 'gradient',
    gradient: {
      shadeIntensity: 1,
      opacityFrom: 0.4,
      opacityTo: 0.1,
      stops: [0, 90, 100]
    }
  },
  xaxis: {
    categories: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '23:59'],
  },
  yaxis: {
    min: 0,
    max: 100,
    tickAmount: 5,
  },
  legend: { position: 'top', horizontalAlign: 'right' },
  dataLabels: { enabled: false },
  grid: { strokeDashArray: 5 }
});
</script>