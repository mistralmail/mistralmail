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
              10
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
          1024
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
          Sent today
        </template>

        <v-card-text>
          <v-col
          class="text-h4"
          cols="6"
        >
          42
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
          Received today
        </template>

        <v-card-text>
          <v-col
          class="text-h4"
          cols="6"
        >
          92
        </v-col>
        </v-card-text>
      </v-card>
    </v-col>

      </v-row>

      

    </v-responsive>
  </v-container>
</template>

<script setup>
  //
</script>
