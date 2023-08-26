<template>
    <v-container class="fill-height">
      <v-responsive class="fill-height">

        <v-btn prepend-icon="mdi-account-plus" color="green" @click="createUserDialog = true">
            Create User
        </v-btn>
        
        <v-table>
            <thead>
            <tr>
                <th class="text-left">
                Email
                </th>
                <th class="text-left">
                Created
                </th>
                <th></th>
            </tr>
            </thead>
            <tbody>
            <tr
                v-for="user in users"
                :key="user.name"
            >
                <td>{{ user.Email }}</td>
                <td>{{ format_date(user.CreatedAt) }}</td>
                <td>
                    <v-dialog
                        v-model="deleteUserDialog"
                        width="auto"
                    >
                        <template v-slot:activator="{ props }">
                            <v-btn
                            prepend-icon="mdi-delete" color="red" size="small"
                            v-bind="props"
                            >
                            Delete
                            </v-btn>
                        </template>
                        <v-card>
                            <v-card-title class="text-h5">
                            Delete user?
                            </v-card-title>
                            <v-card-text></v-card-text>
                            <v-card-actions>
                            <v-spacer></v-spacer>
                            <v-btn
                                color="green"
                                @click="deleteUserDialog = false"
                            >
                                No
                            </v-btn>
                            <v-btn
                                color="red"
                                @click="deleteUser(user.ID)"
                            >
                                Yes
                            </v-btn>
                            </v-card-actions>
                        </v-card>
                    </v-dialog>
                </td>
            </tr>
            </tbody>
        </v-table>

        <v-dialog
            v-model="createUserDialog"
            width=""
            >
            <v-card>
                <v-card-title>Create new user</v-card-title>
                <v-card-text>


                    <v-alert v-if="createUserDialogError"  :text="createUserDialogError" type="error"></v-alert>

                    <v-container style="max-width: 400px;">

                        

                        <v-form @submit.prevent="submitForm">
                            <v-text-field
                                v-model="email"
                                label="E-mail address"
                                type="email"
                                prepend-icon="mdi-email"
                                required
                            ></v-text-field>
                            <v-text-field
                                v-model="password"
                                label="Password"
                                type="password"
                                prepend-icon="mdi-lock"
                                required
                            ></v-text-field>
                            <v-text-field
                                v-model="confirmPassword"
                                label="Confirm password"
                                type="password"
                                prepend-icon="mdi-lock"
                                required
                            ></v-text-field>

                            <v-btn prepend-icon="mdi-account-plus" color="green" type="submit">
                                Create
                            </v-btn>

                        </v-form>

                    </v-container>
                    
                </v-card-text>
                <v-card-actions>
                <v-btn color="primary" block @click="createUserDialog = false">Close Dialog</v-btn>
                </v-card-actions>
            </v-card>
            </v-dialog>

      </v-responsive>
    </v-container>
  </template>
  
  <script setup>
  import { ref, onMounted } from 'vue';
  import { useStore } from 'vuex';
  import axios from 'axios';
  import moment from 'moment'
  
  // Access the Vuex store
  const store = useStore();
  
  // Get the token from the Vuex store
  const token = store.state.token;
  
  const users = ref([]);

  const email = ref('');
  const password = ref('');
  const confirmPassword = ref('');

  const createUserDialogError = ref('')
  const createUserDialog = ref(false)

  const deleteUserDialog = ref(false)
  
  onMounted(async () => {
    load_users();
  });

  function load_users() {
    axios.get('/api/users', {
    headers: {
        Authorization: `Bearer ${token}`,
    },
    })
    .then(function(response){
        users.value = response.data;
    })
    .catch (function(error){
        console.error('Error fetching users:', error);
    })
  }

  function submitForm() {

    // Create a user object with the form data
    const user = {
        email: email.value,
        password: password.value,
        confirmPassword: confirmPassword.value,
    };

    // create user on server
    axios.post('/api/users', user, {headers: {
            Authorization: `Bearer ${token}`,
        }})
        .then(response => {
            load_users()
            resetForm()
            createUserDialog.value = false;
        })
        .catch(error => {
            console.log(error)
            createUserDialogError.value = error.response.data.error
        });

  }

  function resetForm() {
    email.value = ''
    password.value = ''
    confirmPassword.value = ''
    createUserDialogError.value = ''
  }


  function deleteUser(id) {

    axios.delete('/api/users/' + id, {headers: {
        Authorization: `Bearer ${token}`,
    }})
    .then(response => {
        load_users()
    })
    .catch(error => {
        alert("Couldn't delete user: " + error.response.data.error)
    });
    
    deleteUserDialog.value = false
  }

  function format_date(value) {
    if (value) {
        return moment(String(value)).format("LL")
    }
  }
  
  </script>