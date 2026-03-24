/// <reference types="vite/client" />
import 'vue-router'

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}

interface ImportMetaEnv {
  readonly VITE_APP_TITLE: string
  readonly VITE_API_BASE_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module 'vue-router' {
  interface RouteMeta {
    layout?: 'blank'
    requiresAuth?: boolean
    title?: string
    permissions?: string[]
  }
}
