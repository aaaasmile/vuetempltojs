export const app = new Vue({
  el: '#app',
  vuetify: new Vuetify(),
  data: {
      Buildnr: "0.1"
  },
  mounted: function () {
      // `this` points to the vm instance
      this.Buildnr = window.buildnr
  },
  template: `
  <v-app>
   <v-content>
      <div class="home">
        <h2>Tesing</h2>
        <v-btn color="pink">Rosa</v-btn>
        <v-btn>Cliccami</v-btn>
        <v-btn class="pink white--text">
          <v-icon left small>email di Igor</v-icon>
        </v-btn>
        <v-btn fab dark small depressed color="purple">
          <v-icon dark>favorite</v-icon>
        </v-btn>
      </div>
      <p>Buildnr: {{Buildnr}}</p>
    </v-content> 
  </v-app>`
})

console.log('Main is here!', window.buildnr)