<template>
<v-container class="fill-height fluid align-center">
    
    <div class="d-flex align-center justify-center" style="height: 100vh">

        
        

        <v-sheet width="400" class="mx-auto">

            <v-alert v-if="loginError"  :text="loginError" type="error" class="mb-8"></v-alert>
            
            <v-form fast-fail @submit.prevent="login">
                <v-text-field v-model="email" label="E-mail" type="email"></v-text-field>

                <v-text-field v-model="password" label="password" type="password"></v-text-field>
                
                <!-- <a href="#" class="text-body-2 font-weight-regular">Forgot Password?</a>  -->

                <v-btn type="submit" color="primary" block class="mt-2">Sign in</v-btn>

            </v-form>
            <!--<div class="mt-2">
                <p class="text-body-2">Don't have an account? <a href="#">Sign Up</a></p>
            </div>-->
        </v-sheet>
    </div>

</v-container>
</template>

<script>
const LOGIN_URL = '/auth/login';

export default {
    data() {
        return {
            email: '',
            password: '',
            loginError: '',
        };
    },
    methods: {
        login() {
            // Create a FormData object
            const formData = new FormData();

            // Append form fields (username and password) to the FormData object
            formData.append('email', this.email);
            formData.append('password', this.password);

            // Send the FormData object in the POST request
            this.axios.post(LOGIN_URL, formData)
            .then((response) => {
                // Assuming the response includes user and token data
                const { user, token } = response.data;
                this.$store.commit('setUser', user);
                this.$store.commit('setToken', token);
                // You can also save the token in localStorage for persistence
                localStorage.setItem('token', token);

                this.$router.push("/")
            })
            .catch((error) => {
                // Handle errors here
                console.error('Error:', error);

                if (error.response && error.response.data.error) {
                    this.loginError = error.response.data.error;
                } else {
                    this.loginError = "Oops, something went wrong..."
                }
            });
        },
    },
}
</script>
  