import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/LoginView.vue'),
    meta: { layout: 'blank', requiresAuth: false },
  },
  {
    path: '/',
    redirect: '/dashboard',
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/dashboard/DashboardView.vue'),
    meta: { title: '仪表盘' },
  },
  {
    path: '/asset',
    name: 'Asset',
    component: () => import('@/views/asset/AssetList.vue'),
    meta: { title: '资产管理' },
  },
  {
    path: '/cluster',
    name: 'ClusterList',
    component: () => import('@/views/cluster/ClusterList.vue'),
    meta: { title: '集群管理', permissions: ['cluster:list'] },
  },
  {
    path: '/cluster/:id',
    name: 'ClusterDetail',
    component: () => import('@/views/cluster/ClusterDetail.vue'),
    meta: { title: '集群详情', permissions: ['cluster:list'] },
  },
  {
    path: '/workload',
    name: 'Workload',
    component: () => import('@/views/workload/WorkloadList.vue'),
    meta: { title: '工作负载', permissions: ['cluster:list'] },
  },
  {
    path: '/network',
    name: 'Network',
    component: () => import('@/views/network/NetworkList.vue'),
    meta: { title: '网络管理', permissions: ['cluster:list'] },
  },
  {
    path: '/config',
    name: 'Config',
    component: () => import('@/views/config/ConfigList.vue'),
    meta: { title: '配置管理', permissions: ['cluster:list'] },
  },
  {
    path: '/node',
    name: 'NodeList',
    component: () => import('@/views/node/NodeList.vue'),
    meta: { title: '节点管理', permissions: ['cluster:list'] },
  },
  {
    path: '/storage',
    name: 'Storage',
    component: () => import('@/views/storage/StorageList.vue'),
    meta: { title: '存储管理', permissions: ['cluster:list'] },
  },
  {
    path: '/ops/alert',
    name: 'AlertCenter',
    component: () => import('@/views/ops/AlertCenter.vue'),
    meta: { title: '告警中心', permissions: ['alert:list'] },
  },
  {
    path: '/ops/log',
    name: 'LogCenter',
    component: () => import('@/views/ops/LogCenter.vue'),
    meta: { title: '日志检索', permissions: ['log:list'] },
  },
  {
    path: '/ops/monitor',
    name: 'MonitorCenter',
    component: () => import('@/views/ops/MonitorCenter.vue'),
    meta: { title: '监控配置', permissions: ['monitor:list'] },
  },
  {
    path: '/ops/harbor',
    name: 'HarborCenter',
    component: () => import('@/views/ops/HarborCenter.vue'),
    meta: { title: 'Harbor管理', permissions: ['harbor:list'] },
  },
  {
    path: '/ops/cicd',
    name: 'CICDCenter',
    component: () => import('@/views/ops/CICDCenter.vue'),
    meta: { title: 'CI/CD流水线', permissions: ['cicd:list'] },
  },
  {
    path: '/ops/app',
    name: 'AppCenter',
    component: () => import('@/views/ops/AppCenter.vue'),
    meta: { title: '应用管理', permissions: ['app:list'] },
  },
  {
    path: '/ops/audit',
    name: 'AuditCenter',
    component: () => import('@/views/ops/AuditCenter.vue'),
    meta: { title: '审计日志', permissions: ['audit:list'] },
  },
  {
    path: '/system/users',
    name: 'UserList',
    component: () => import('@/views/system/UserList.vue'),
    meta: { title: '用户管理', permissions: ['user:list'] },
  },
  {
    path: '/system/departments',
    name: 'DepartmentList',
    component: () => import('@/views/system/DepartmentList.vue'),
    meta: { title: '部门管理', permissions: ['department:list'] },
  },
  {
    path: '/system/roles',
    name: 'RoleList',
    component: () => import('@/views/system/RoleList.vue'),
    meta: { title: '角色管理', permissions: ['role:list'] },
  },
  {
    path: '/system/permissions',
    name: 'PermissionList',
    component: () => import('@/views/system/PermissionList.vue'),
    meta: { title: '权限管理', permissions: ['permission:list'] },
  },
]

export default routes
