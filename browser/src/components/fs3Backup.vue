<template>
    <div class="fs3_back">
      <div class="fs3_head">
        <div class="fs3_head_text">
          <div class="titleBg">FS3 Backup</div>
          <h1>FS3 Backup</h1>
          <h3>A simpler and faster storage calculation service for Filecoin network.</h3>
          <h5>FS3 Backup automatically retains backups for resources according to users backup plan.</h5>
        </div>
        <img src="@/assets/images/page_bg.png" class="bg" alt="">
      </div>
      <div class="fs3_cont">
        <div class="title">How to use FS3</div>
        <div class="content_list" v-for="(item, index) in conList" :key="index">
          <div class="left"><span>{{ index+1 }}</span></div>
          <div class="right">
            <h4>{{item.title}}</h4>
            <h6>{{item.desc}}</h6>
          </div>
        </div>
      </div>
    </div>
</template>

<script>
import axios from 'axios'
export default {
    data() {
        return {
          width: document.body.clientWidth>600?'400px':'95%',
          conList: [
            {
              title: 'Upload resources',
              desc: 'Start with your backup by creating buckets and uploading resources.',
            },
            {
              title: 'Choose your plan',
              desc: 'Choose your backup plan including frequency,location and price your backup resources.',
            },
            {
              title: 'Create successfully',
              desc: '',
            }
          ]
        }
    },
    watch: {},
    methods: {

    },
    mounted() {},
};
</script>

<style lang="scss" scoped>
.fs3_back{
  font-size: 0.18rem;
  .fs3_head{
    position: relative;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 39% 0.2rem 9%;
    background: #7ecef4;
    color: #fff;
    .bg{
      position: absolute;
      right: 13%;
      width: 17%;
      top: 0.5rem;
      z-index: 5;
    }
    .fs3_head_text{
      .titleBg{
        font-size: 0.76rem;
        font-family: 'm-light';
        color: #fff;
        opacity: 0.3;
        line-height: 0.45;
        text-indent: -0.04rem;
      }
      h1{
        line-height: 1.6;
        font-size: 0.23rem;
        font-weight: bold;
        // font-family: 'm-semibold';
      }
      h3{
        margin: 0.2rem 0 0.05rem;
        line-height: 1.2;
        font-size: 0.22rem;
      }
      h5{
        line-height: 1.2;
        font-size: 0.14rem;
      }
    }
  }
  .fs3_cont{
    padding: 0.4rem 9%;
    .title{
      font-size: 0.23rem;
      color: #2f85e5;
      line-height: 1.74;
    }
    .content_list{
      display: flex;
      margin: 0.26rem 0 0;
      .left{
        width: 0.45rem;
        span{
          display: flex;
          align-items: center;
          justify-content: center;
          width: 0.22rem;
          height: 0.22rem;
          background: #4472c4;
          font-size: 0.18rem;
          color: #fff;
          border-radius: 100%;
        }
      }
      .right{
        h4{
          font-size: 0.18rem;
          line-height: 1.13;
        }
        h6{
          margin: 0.1rem 0 0.07rem;
          line-height: 1.8;
          font-size: 0.14rem;
        }
      }
    }
  }
}

@media screen and (max-width:769px){
  .fs3_back{

  }
}
@media screen and (max-width:600px){

}
</style>
