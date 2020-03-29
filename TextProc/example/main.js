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
    template: ``
})

console.log('Main is here!', window.buildnr)