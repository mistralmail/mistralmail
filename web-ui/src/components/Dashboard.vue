<template>
  <v-container class="fill-height">
    <v-responsive class="fill-height d-flex">
      

      <v-row>

        <v-col cols="12" sm="6" md="3">

          <v-card
            class="mx-auto "
            prepend-icon="mdi-account-multiple"
          >
            <template v-slot:title>
              Users
            </template>

            <v-card-text>
              <v-col
              class="text-h4"
              cols="6"
            >
            {{  metrics.Users }}
            </v-col>
            </v-card-text>
          </v-card>

        </v-col>

        

        <v-col cols="12" sm="6" md="3">


      <v-card
        class="mx-auto"
        prepend-icon="mdi-email"
      >
        <template v-slot:title>
          Total messages
        </template>

        <v-card-text>
          <v-col
          class="text-h4"
          cols="6"
        >
        {{  metrics.Messages }}
        </v-col>
        </v-card-text>
      </v-card>

    </v-col>

    <v-col cols="12" sm="6" md="3">

      <v-card
        class="mx-auto"
        prepend-icon="mdi-email-arrow-right"
      >
        <template v-slot:title>
          SMTP delivered
        </template>

        <v-card-text>
          <v-col
          class="text-h4"
          cols="6"
        >
        {{  metrics.SMTPDelivered }}
        </v-col>
        </v-card-text>
      </v-card>

      </v-col>

      <v-col cols="12" sm="6" md="3">

      <v-card
        class="mx-auto"
        prepend-icon="mdi-email-plus"
      >
        <template v-slot:title>
          SMTP received
        </template>

        <v-card-text>
          <v-col
          class="text-h4"
          cols="6"
        >
        {{  metrics.SMTPReceived }}
        </v-col>
        </v-card-text>
      </v-card>
    </v-col>

      </v-row>

      

    </v-responsive>
  </v-container>
</template>

<script setup>
  import { ref, onMounted } from 'vue';
  import { useStore } from 'vuex';
  import axios from 'axios';
  
  // Access the Vuex store
  const store = useStore();
  
  // Get the token from the Vuex store
  const token = store.state.token;
  
  const metrics = ref([]);

  
  onMounted(async () => {
    load_metrics();
  });

  function load_metrics() {
    axios.get('/api/metrics', {
    headers: {
        Authorization: `Bearer ${token}`,
    },
    })
    .then(function(response){
        metrics.value = response.data;
    })
    .catch (function(error){
        console.error('Error fetching metrics:', error);
    })
  }
</script>
