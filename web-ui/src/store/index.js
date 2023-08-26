// Vuex
import { createStore } from 'vuex'

// Create a new store instance.
const store = createStore({
    state: {
        user: null, // Initialize as null or with default user data
        token: localStorage.getItem('token')
    },
    mutations: {
        setUser(state, user) {
            state.user = user;
        },
        setToken(state, token) {
            state.token = token;
        },
    },
    getters: {
        isLoggedIn(state) {
          return !!state.token;
        },
    },
})

export default store