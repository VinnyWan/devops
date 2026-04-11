import { createRouter, createWebHistory } from 'vue-router'
import { useUserStore } from '../stores/user'

const routes = [
  {
    path: '/login',
    component: () => import('../views/Login/index.vue')
  },
  {
    path: '/',
    component: () => import('../components/Layout/MainLayout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        component: () => import('../views/Dashboard/index.vue')
      },
      {
        path: 'k8s/cluster',
        component: () => import('../views/K8s/k8s-clusters.vue')
      },
      {
        path: 'k8s/cluster/:name',
        component: () => import('../views/K8s/ClusterDetail.vue')
      },
      {
        path: 'k8s/node',
        component: () => import('../views/K8s/NodeList.vue')
      },
      {
        path: 'k8s/namespace',
        component: () => import('../views/K8s/NamespaceList.vue')
      },
      {
        path: 'k8s/workload',
        component: () => import('../views/K8s/WorkloadList.vue')
      },
      {
        path: 'k8s/pod/:name',
        component: () => import('../views/K8s/PodDetail.vue')
      },
      {
        path: 'k8s/network',
        component: () => import('../views/K8s/NetworkList.vue')
      },
      {
        path: 'k8s/config',
        component: () => import('../views/K8s/ConfigList.vue')
      },
      {
        path: 'system/user',
        component: () => import('../views/System/UserList.vue')
      },
      {
        path: 'system/role',
        component: () => import('../views/System/RoleList.vue')
      },
      {
        path: 'system/department',
        component: () => import('../views/System/DepartmentList.vue')
      },
      {
        path: 'system/permission',
        component: () => import('../views/System/PermissionList.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  const userStore = useUserStore()

  if (!userStore.userInfo && userStore.token) {
    userStore.loadUserInfo()
  }

  if (to.path !== '/login' && !userStore.token) {
    next('/login')
  } else {
    next()
  }
})

export default router
