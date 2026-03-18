import { createApp } from 'vue'
import { createPinia } from 'pinia'
import '@fontsource/noto-sans-sc/chinese-simplified-400.css'
import '@fontsource/noto-sans-sc/chinese-simplified-500.css'
import '@fontsource/noto-sans-sc/chinese-simplified-600.css'
import '@fontsource/noto-sans-sc/chinese-simplified-700.css'
import '@fontsource/noto-sans-sc/latin-400.css'
import '@fontsource/noto-sans-sc/latin-500.css'
import '@fontsource/noto-sans-sc/latin-600.css'
import '@fontsource/noto-sans-sc/latin-700.css'

import AppRoot from '@/app/AppRoot.vue'
import router from '@/router'
import '@/styles.css'

const app = createApp(AppRoot)

app.use(createPinia())
app.use(router)
app.mount('#app')
