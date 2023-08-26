<template>
  <v-app>
    <default-bar />

    <v-navigation-drawer
      >

        <v-list density="compact" nav>

          <v-list-item prepend-icon="mdi-monitor-dashboard" title="Dashboard" value="dashboard" to="/dashboard"></v-list-item>

          <v-list-item prepend-icon="mdi-account-multiple" title="Users" value="users" to="/users"></v-list-item>

        </v-list>


        <template v-slot:append v-if="isLoggedIn">
          <div class="pa-2 text-center">
            <v-btn prepend-icon="mdi-logout" color="red" @click="logout()">
              Logout
            </v-btn>
          </div>
        </template>

      </v-navigation-drawer>

    <default-view />
  </v-app>
</template>

<script setup>
  import DefaultBar from './AppBar.vue'
  import DefaultView from './View.vue'

  import { useStore } from 'vuex'

  import router from '@/router'
  
  const store = useStore()
  const isLoggedIn = store.getters.isLoggedIn


  function logout() {
    store.commit('setToken', null);
    localStorage.removeItem('token');
    router.push("/")
  }

  
</script>
