<template>
  <div class="w-full">
    <div v-if="loading" class="text-center py-10 text-gray-500 font-medium">
      A interrogar registos históricos no InfluxDB...
    </div>
    <VueApexCharts
      v-else
      type="area"
      height="350"
      :options="chartOptions"
      :series="series"
    />
  </div>
</template>

<script setup>
import VueApexCharts from "vue3-apexcharts";
import { ref, onMounted } from "vue";
import { API_BASE_URL } from '@/config';

const series = ref([]);
const loading = ref(true);

const chartOptions = ref({
  colors: ['#465fff', '#33D1FF', '#10B981', '#FF9C55'],
  chart: {
    fontFamily: 'Outfit, sans-serif',
    type: 'area',
    toolbar: { show: false },
  },
  stroke: { curve: 'smooth', width: 2.5 },
  fill: {
    type: 'gradient',
    gradient: {
      shadeIntensity: 1,
      opacityFrom: 0.2,
      opacityTo: 0.02,
      stops: [0, 90, 100]
    }
  },
  xaxis: {
    categories: [],
    labels: { style: { colors: '#9ca3af' } }
  },
  yaxis: {
    show: false // Ocultamos o eixo Y porque os números internos de offset confundiriam o utilizador
  },
  legend: { 
    position: 'top', 
    horizontalAlign: 'right',
    labels: { colors: '#9ca3af' }
  },
  dataLabels: { enabled: false },
  grid: { 
    strokeDashArray: 5,
    xaxis: { lines: { show: true } }
  }
});

const fetchUptimeMetrics = async () => {
  try {
    const response = await fetch(`${API_BASE_URL}/metrics/health?range=24h&interval=5m`);
    if (response.ok) {
      const data = await response.json();
      
      if (data.series && data.series.length > 0) {
        // 1. Extrair e formatar as horas do eixo X baseando-nos nos pontos do primeiro serviço
        const timeLabels = data.series[0].points.map(pt => {
          const time = new Date(pt.ts);
          return time.toLocaleTimeString('pt-PT', { hour: '2-digit', minute: '2-digit' });
        });

        // 2. Forçar reatividade atualizando a referência do objeto de opções do gráfico
        chartOptions.value = {
          ...chartOptions.value,
          xaxis: {
            ...chartOptions.value.xaxis,
            categories: timeLabels
          }
        };

        // 3. Mapear as séries aplicando o deslocamento vertical (offset) dinamicamente
        series.value = data.series.map((srv, index) => {
          // Cada serviço ganha um "andar" diferente no gráfico (Separados de 1.5 em 1.5 unidades)
          const offset = index * 1.5; 
          
          return {
            name: srv.service.toUpperCase(),
            // Se o status for 1 (ONLINE), renderiza no teto do seu canal. Se for 0 (OFFLINE), cai para o fundo do canal.
            data: srv.points.map(pt => pt.status === 1 ? 1 + offset : 0 + offset)
          };
        });
      }
    }
  } catch (error) {
    console.error('Erro ao carregar dados do InfluxDB:', error);
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  fetchUptimeMetrics();
});
</script>