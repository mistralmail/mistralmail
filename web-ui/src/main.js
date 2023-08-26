/**
 * main.js
 *
 * Bootstraps Vuetify and other plugins then mounts the App`
 */

// Components
import App from './App.vue'

// Composables
import { createApp } from 'vue'

// Plugins
import { registerPlugins } from '@/plugins'

// Axios
import axios from 'axios'
import VueAxios from 'vue-axios'

import store from '@/store'
  

const app = createApp(App)

app.use(VueAxios, axios)
app.use(store)

registerPlugins(app)

app.mount('#app')
