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
              desc: 'Choose your backup plan based on frequency, duration,and price of your backup resources.',
            },
            {
              title: 'Create successfully',
              desc: 'Users can check all created backup plans at My Plans. FS3 also offers rebuilt services of buckets at Dashboard.',
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
  @media screen and (max-width:600px){
    font-size: 14px;
  }
  .fs3_head{
    position: relative;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.3rem 30% 0.1rem 9%;
    background: #7ecef4;
    color: #fff;
    .bg{
      position: absolute;
      right: 11%;
      width: 1.6rem;
      top: 0.3rem;
      z-index: 5;
      @media screen and (max-width:999px){    
        right: 0.2rem;
        width: 1.8rem;
      }
      @media screen and (max-width:600px){    
        top: 1.2rem;
      }
    }
    .fs3_head_text{
      .titleBg{
        font-size: 0.6rem;
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
        @media screen and (max-width:600px){
          font-size: 14px;
        }
      }
      h3{
        margin: 0.05rem 0 0.05rem;
        line-height: 1.1;
        font-size: 0.22rem;
        @media screen and (max-width:600px){
          font-size: 12px;
        }
      }
      h5{
        line-height: 1.2;
        font-size: 0.14rem;
        @media screen and (max-width:600px){
          font-size: 12px;
        }
      }
    }
  }
  .fs3_cont{
    padding: 0.4rem 9%;
    .title{
      font-size: 0.23rem;
      color: #2f85e5;
      line-height: 1.74;
        @media screen and (max-width:600px){
          font-size: 15px;
        }
    }
    .content_list{
      display: flex;
      margin: 0.26rem 0 0;
      .left{
        width: 0.45rem;
        @media screen and (max-width:600px){
          width: 30px;
        }
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
          @media screen and (max-width:600px){
            width: 15px;
            height: 15px;
            font-size: 12px;
          }
        }
      }
      .right{
        width: calc(100% - 0.45rem);
        @media screen and (max-width:600px){
          width: calc(100% - 30px);
        }
        h4{
          font-size: 0.18rem;
          line-height: 1.13;
          @media screen and (max-width:600px){
            font-size: 14px;
          }
        }
        h6{
          margin: 0.1rem 0 0.07rem;
          line-height: 1.8;
          font-size: 0.14rem;
          @media screen and (max-width:600px){
            font-size: 12px;
          }
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
