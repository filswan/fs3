<template>
    <div id="app">
        <router-view></router-view>
    </div>
</template>
<script>
export default {
  name: 'APP',
  data() {
      return{
      }
  },
  methods: {
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
