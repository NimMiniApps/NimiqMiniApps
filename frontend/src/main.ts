import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'
import { setPageMeta, resetPageMeta } from './utils/meta'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: () => import('./views/HomeView.vue'), meta: { title: 'Home' } },
    { path: '/apps', component: () => import('./views/AppsView.vue'), meta: { title: 'All Apps' } },
    { path: '/apps/:slug/update', component: () => import('./views/RequestUpdateView.vue'), meta: { title: 'Suggest an Update' } },
    { path: '/apps/:slug', component: () => import('./views/AppDetailView.vue') },
    { path: '/developers/:slug', redirect: (to) => ({ path: '/apps', query: { developer: to.params.slug as string } }) },
    { path: '/developers', redirect: '/apps' },
    { path: '/categories', redirect: '/apps' },
    { path: '/build', component: () => import('./views/BuildView.vue'), meta: { title: 'Build a Mini App' } },
    { path: '/submit', component: () => import('./views/SubmitView.vue'), meta: { title: 'Submit Your App' } },
    { path: '/status/:slug', component: () => import('./views/StatusView.vue'), meta: { title: 'Submission Status' } },
    { path: '/admin', component: () => import('./views/AdminView.vue'), meta: { title: 'Admin' } },
  ],
  scrollBehavior(to) {
    if (to.hash) return { el: to.hash, behavior: 'smooth' }
    return { top: 0 }
  },
})

router.afterEach((to) => {
  if (/^\/apps\/[^/]+$/.test(to.path) && !to.path.endsWith('/update')) return
  if (to.path === '/apps' && to.query.developer) return
  const title = typeof to.meta.title === 'string' ? to.meta.title : undefined
  if (title) {
    setPageMeta({
      title,
      url: window.location.href,
    })
  } else {
    resetPageMeta()
  }
})

createApp(App).use(router).mount('#app')
