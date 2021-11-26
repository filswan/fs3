<template>
    <div class="company">
        <transition name="appFade" mode="out-in">
            <router-view></router-view>
        </transition>
    </div>
</template>
<script>
export default {
    data () {
        return {

        }
    }
}
</script>
<style lang="scss" scoped>
.company{
    box-sizing: border-box;
    padding: 0;
    background: rgb(255, 255, 255);
    font-size: 0.2rem;
}
    
.appFade-enter-active,
.appFade-leave-active {
  transition: opacity 0.3s;
}
.appFade-enter,
.appFade-leave-to {
  opacity: 0;
}
@media screen and (max-width:999px){
    
}
</style>