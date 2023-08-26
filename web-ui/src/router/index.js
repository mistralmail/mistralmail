// Composables
import { createRouter, createWebHistory } from 'vue-router'
import store from '@/store'

const routes = [
  {
    path: '/',
    component: () => import('@/layouts/default/Default.vue'),
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: '/users',
        name: 'Users',
        component: () => import(/* webpackChunkName: "home" */ '@/views/Users.vue'),
      },
      {
        path: '/dashboard',
        name: 'Dashboard',
        component: () => import(/* webpackChunkName: "home" */ '@/views/Dashboard.vue'),
      },
      {
        path: '/login',
        name: 'Login',
        meta: {
          noAuth: true
        },
        component: () => import(/* webpackChunkName: "home" */ '@/views/Login.vue'),
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(process.env.BASE_URL),
  routes,
})

router.beforeEach((to, from, next) => {
  if (to.matched.some(record => record.meta.noAuth)) {
    // this route requires auth, check if logged in
    // if not, redirect to login page.
    next() // does not require auth, make sure to always call next()!
    

  } else {
    
    if (!store.getters.isLoggedIn) {
      next({ name: 'Login' })
    } else {
      next() // go to wherever I'm going
    }
  }
})

export default router
