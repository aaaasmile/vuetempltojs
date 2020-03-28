<!-- my-component.vue -->
<template>
  <v-app>
    <v-content>
      <div class="home">
        <h2>Testing Vuetify</h2>
        <v-btn>Click Me</v-btn>
        <v-btn color="pink">Rosa</v-btn>
        <v-btn class="pink white--text">
          <v-icon left small>email</v-icon>
        </v-btn>
        <v-btn fab dark small depressed color="purple">
          <v-icon dark>favorite</v-icon>
        </v-btn>
      </div>
      <p>Buildnr: {{Buildnr}}</p>
    </v-content>
  </v-app>
</template>