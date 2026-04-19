import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'element-plus/dist/index.css'
import './assets/styles/variables.css'
import './assets/styles/global.css'
import App from './App.vue'
import router from './router'
import { vPermission, vPermissionField } from './directives/permission'

const app = createApp(App)

app.config.errorHandler = (err, instance, info) => {
  console.error('全局错误:', err, info)
}

app.use(createPinia())
app.use(router)

app.directive('permission', vPermission)
app.directive('permission-field', vPermissionField)

app.mount('#app')
