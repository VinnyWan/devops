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
        component: () => import('../views/k8s/k8s-clusters.vue')
      },
      {
        path: 'k8s/cluster/:name',
        component: () => import('../views/k8s/ClusterDetail.vue')
      },
      {
        path: 'k8s/node',
        component: () => import('../views/k8s/NodeList.vue')
      },
      {
        path: 'k8s/node/:clusterName/:nodeName',
        component: () => import('../views/k8s/NodeDetail.vue')
      },
      {
        path: 'k8s/namespace',
        component: () => import('../views/k8s/NamespaceList.vue')
      },
      {
        path: 'k8s/workload',
        component: () => import('../views/k8s/WorkloadList.vue')
      },
      {
        path: 'k8s/workload/:kind/:clusterName/:namespace/:name',
        component: () => import('../views/k8s/WorkloadDetail.vue')
      },
      {
        path: 'k8s/pod/:name',
        component: () => import('../views/k8s/PodDetail.vue')
      },
      {
        path: 'k8s/network',
        component: () => import('../views/k8s/NetworkList.vue')
      },
      {
        path: 'k8s/config',
        component: () => import('../views/k8s/ConfigList.vue')
      },
      {
        path: 'k8s/storage',
        component: () => import('../views/k8s/StorageList.vue')
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
      },
      {
        path: 'audit/operation',
        component: () => import('../views/Audit/OperationLog.vue')
      },
      {
        path: 'audit/login',
        component: () => import('../views/Audit/LoginLog.vue')
      },
      {
        path: 'cmdb/hosts',
        component: () => import('../views/Cmdb/HostList.vue')
      },
      {
        path: 'cmdb/groups',
        component: () => import('../views/Cmdb/GroupList.vue')
      },
      {
        path: 'cmdb/credentials',
        component: () => import('../views/Cmdb/CredentialList.vue')
      },
      {
        path: 'cmdb/terminal/sessions',
        component: () => import('../views/Cmdb/TerminalSessionList.vue')
      },
      {
        path: 'cmdb/terminal/replay/:id',
        component: () => import('../views/Cmdb/TerminalReplay.vue')
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
