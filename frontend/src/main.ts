import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('./views/HomeView.vue') },
    { path: '/apps', component: () => import('./views/AppsView.vue') },
    { path: '/apps/:slug', component: () => import('./views/AppDetailView.vue') },
    { path: '/categories', redirect: '/apps' },
    { path: '/developers/:slug', component: () => import('./views/DeveloperView.vue') },
    { path: '/submit', component: () => import('./views/SubmitView.vue') },
    { path: '/admin', component: () => import('./views/AdminView.vue') },
  ],
  scrollBehavior: () => ({ top: 0 }),
})

createApp(App).use(router).mount('#app')
