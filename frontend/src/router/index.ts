import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition || { left: 0, top: 0 }
  },
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: {
        title: 'Dashboard',
        fullWidth: true,
      },
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('../views/Tables/UsersView.vue'),
      meta: {
        title: 'RFID Users List',
      },
    },
    {
      path: '/device/control',
      name: 'DeviceControl',
      component: () => import('../views/Pages/DeviceControlView.vue'),
      meta: {
        title: 'Remote Unlock Control',
      },
    },
    {
      path: '/ai/evaluation',
      name: 'AiEvaluation',
      component: () => import('../views/Pages/AiEvaluationView.vue'),
      meta: {
        title: 'AI Evaluation Metrics',
      },
    },
    {
      path: '/add-users',
      name: 'Add Users',
      component: () => import('../views/Forms/AddUsers.vue'),
      meta: {
        title: 'Adicionar Utilizadores',
      },
    },
    {
      path: '/line-chart',
      name: 'Line Chart',
      component: () => import('../views/Chart/LineChart/LineChart.vue'),
    },
    {
      path: '/bar-chart',
      name: 'Bar Chart',
      component: () => import('../views/Chart/BarChart/BarChart.vue'),
    },
    {
      path: '/alerts',
      name: 'Alerts',
      component: () => import('../views/UiElements/AlertPage.vue'),
      meta: {
        title: 'Alerts',
      },
    },
    {
      path: '/avatars',
      name: 'Avatars',
      component: () => import('../views/UiElements/Avatars.vue'),
      meta: {
        title: 'Avatars',
      },
    },
    {
      path: '/badge',
      name: 'Badge',
      component: () => import('../views/UiElements/Badges.vue'),
      meta: {
        title: 'Badge',
      },
    },
    {
      path: '/buttons',
      name: 'Buttons',
      component: () => import('../views/UiElements/Buttons.vue'),
      meta: {
        title: 'Buttons',
      },
    },
    {
      path: '/images',
      name: 'Images',
      component: () => import('../views/UiElements/Images.vue'),
      meta: {
        title: 'Images',
      },
    },
    {
      path: '/videos',
      name: 'Videos',
      component: () => import('../views/UiElements/Videos.vue'),
      meta: {
        title: 'Videos',
      },
    },
    // Rota corrigida para a tua BlankView.vue baseada na pasta correta
    {
      path: '/blank',
      name: 'Blank',
      component: () => import('../views/Errors/BlankPage.vue'),
      meta: {
        title: 'Página Base',
      },
    },
    // Rota corrigida e limpa para o Erro 501
    {
      path: '/501',
      name: 'not-implemented',
      component: () => import('../views/Errors/FiveZeroOne.vue'),
      meta: { 
        title: '501 Not Implemented' 
      }
    },
    // O Catch-all do 404 deve ficar SEMPRE no fundo da lista
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: () => import('../views/Errors/FourZeroFour.vue'),
      meta: {
        title: '404 Not Found',
      },
    },
  ],
})

export default router

router.beforeEach((to, _from, next) => {
  document.title = `Vue.js ${to.meta.title} | TailAdmin - Vue.js Tailwind CSS Dashboard Template`
  next()
})
