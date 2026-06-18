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
        component: () => import('../views/k8s/ClusterDetail.vue'),
        // Keep the cluster list entry active on detail pages.
        meta: { activeMenu: '/k8s/cluster' }
      },
      {
        path: 'k8s/node',
        component: () => import('../views/k8s/NodeList.vue')
      },
      {
        path: 'k8s/node/:clusterName/:nodeName',
        component: () => import('../views/k8s/NodeDetail.vue'),
        meta: { activeMenu: '/k8s/node' }
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
        component: () => import('../views/k8s/WorkloadDetail.vue'),
        meta: { activeMenu: '/k8s/workload' }
      },
      {
        path: 'k8s/pod/:name',
        component: () => import('../views/k8s/PodDetail.vue'),
        meta: { activeMenu: '/k8s/workload' }
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
        component: () => import('../views/Cmdb/TerminalReplay.vue'),
        meta: { activeMenu: '/cmdb/terminal/sessions' }
      },
      {
        path: 'cmdb/permissions',
        component: () => import('../views/Cmdb/PermissionList.vue')
      },
      {
        path: 'cmdb/cloud-accounts',
        component: () => import('../views/Cmdb/CloudAccountList.vue')
      },
      {
        path: 'cmdb/files',
        name: 'CmdbFiles',
        component: () => import('../views/Cmdb/FileBrowser.vue'),
        meta: { title: '文件管理' }
      },
      {
        path: 'cmdb/batch-command',
        name: 'CmdbBatchCommand',
        component: () => import('../views/Cmdb/BatchCommand.vue'),
        meta: { title: '批量命令' }
      },
      {
        path: 'workflow/orders',
        component: () => import('../views/Workflow/OrderList.vue')
      },
      {
        path: 'workflow/orders/:id',
        component: () => import('../views/Workflow/OrderDetail.vue'),
        meta: { activeMenu: '/workflow/orders' }
      },
      {
        path: 'tools',
        component: () => import('../views/Workflow/ToolMarket.vue')
      },
      {
        path: 'tools/templates',
        component: () => import('../views/Workflow/ToolTemplates.vue')
      },
      {
        path: 'sql-audit',
        component: () => import('../views/Workflow/SqlAudit.vue')
      },
      {
        path: 'monitor/prometheus',
        component: () => import('../views/Monitor/PrometheusConfig.vue')
      },
      {
        path: 'monitor/metrics',
        component: () => import('../views/Monitor/HostMetrics.vue')
      },
      {
        path: 'monitor/agent',
        component: () => import('../views/Monitor/AgentStatus.vue')
      },
      {
        path: 'cicd/jenkins',
        component: () => import('../views/Cicd/JenkinsConfig.vue')
      },
      {
        path: 'cicd/jobs',
        component: () => import('../views/Cicd/JobList.vue')
      },
      {
        path: 'harbor/registry',
        component: () => import('../views/Harbor/RegistryList.vue')
      },
      {
        path: 'log/search',
        component: () => import('../views/Log/LogSearch.vue')
      },
      {
        path: 'knowledge/articles',
        component: () => import('../views/Knowledge/ArticleList.vue')
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
