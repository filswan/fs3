<template>
    <div id="app">
        <router-view v-if="isRouterAlive"></router-view>
    </div>
</template>
<script>
export default {
  name: 'APP',
  provide () {
      return {
          reload: this.reload
      }
  },
  data() {
      return{
          isRouterAlive: true
      }
  },
  methods: {
      reload () {
          this.isRouterAlive = false;
          this.$nextTick(function () {
              this.isRouterAlive = true;
          })
      }
  },
  mounted() {
    /* window.onbeforeunload = function(){
      sessionStorage.removeItem('addrWeb')
    }*/
  }
}
</script>
<style>
    @import "./assets/css/main.css";
    @import "./assets/css/iconfont.css";
</style>
