<template>
    <div class="fs3_back">
      <div class="fs3_head">
        <div class="fs3_head_text">
          <div class="titleBg">Dashboard</div>
          <h1>Dashboard</h1>
        </div>
        <img src="@/assets/images/page_bg.png" class="bg" alt="">
      </div>
      <div class="fs3_cont">
        <el-row>
          <el-col :span="12" v-for="(item,index) in card" :key="index">
            <el-card class="box-card">
              <div slot="header" class="clearfix">
                <span>{{item.title}}</span>
              </div>
              <div class="statusStyle">
                <div class="list">
                  <svg t="1638936551170" class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" p-id="3518" width="128" height="128"><path d="M661.333333 170.666667l253.866667 34.133333-209.066667 209.066667zM362.666667 853.333333L108.8 819.2l209.066667-209.066667zM170.666667 362.666667L204.8 108.8l209.066667 209.066667z" fill="#f8b551" p-id="3519"></path><path d="M198.4 452.266667l-89.6 17.066666c-2.133333 14.933333-2.133333 27.733333-2.133333 42.666667 0 98.133333 34.133333 192 98.133333 264.533333l64-55.466666C219.733333 663.466667 192 588.8 192 512c0-19.2 2.133333-40.533333 6.4-59.733333zM512 106.666667c-115.2 0-217.6 49.066667-292.266667 125.866666l59.733334 59.733334C339.2 230.4 420.266667 192 512 192c19.2 0 40.533333 2.133333 59.733333 6.4l14.933334-83.2C563.2 108.8 537.6 106.666667 512 106.666667zM825.6 571.733333l89.6-17.066666c2.133333-14.933333 2.133333-27.733333 2.133333-42.666667 0-93.866667-32-185.6-91.733333-258.133333l-66.133333 53.333333c46.933333 57.6 72.533333 130.133333 72.533333 202.666667 0 21.333333-2.133333 42.666667-6.4 61.866666zM744.533333 731.733333C684.8 793.6 603.733333 832 512 832c-19.2 0-40.533333-2.133333-59.733333-6.4l-14.933334 83.2c25.6 4.266667 51.2 6.4 74.666667 6.4 115.2 0 217.6-49.066667 292.266667-125.866667l-59.733334-57.6z" fill="#f8b551" p-id="3520"></path><path d="M853.333333 661.333333l-34.133333 253.866667-209.066667-209.066667z" fill="#f8b551" p-id="3521"></path></svg> 
                  In process jobs: {{item.inProcessJobs}}
                </div>
                <div class="list"><span class="el-icon-success"></span> Completed jobs: {{item.completedJobs}}</div>
                <div class="list"><span class="el-icon-error"></span> Failed jobs: {{item.failedJobs}}</div>
              </div>
              <el-button @click="link(index)">{{item.btn}}</el-button>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </div>
</template>

<script>
import axios from 'axios'
export default {
    data() {
        return {
          width: document.body.clientWidth>600?'400px':'95%',
          card: [
            {
              title: 'Your backup jobs',
              btn: 'All backup jobs details',
              inProcessJobs: 0,
              completedJobs: 0,
              failedJobs: 0
            },
            {
              title: 'Your rebuild jobs',
              btn: 'All rebuild jobs details',
              inProcessJobs: 0,
              completedJobs: 0,
              failedJobs: 0
            }
          ]
        }
    },
    watch: {},
    methods: {
      link(index) {
        let type = index ? 'rebuild_job' : 'backup_job'
        this.$router.push({name: 'my_account_dashboard_detail', params: { type: type}})
      },
      getData() {
          let _this = this
          let params = {
            "Offset":0,   //default as 0 
            "Limit":10   //default as 10
          }

          axios.post(_this.data_api + `/minio/backup/retrieve/volume`, params, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
          }}).then((response) => {
              let json = response.data
              if (json.status == 'success') {
                _this.card[0].inProcessJobs = json.data.inProcessVolumeBackupTasksCounts
                _this.card[0].completedJobs = json.data.completedVolumeBackupTasksCounts
                _this.card[0].failedJobs = json.data.failedVolumeBackupTasksCounts
              }else{
                  _this.$message.error(json.message);
                  return false
              }

          }).catch(function (error) {
              console.log(error);
          });

          axios.post(_this.data_api + `/minio/rebuild/retrieve/volume`, params, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
          }}).then((response) => {
              let json = response.data
              if (json.status == 'success') {
                _this.card[1].inProcessJobs = json.data.inProcessVolumeRebuildTasksCounts
                _this.card[1].completedJobs = json.data.completedVolumeRebuildTasksCounts
                _this.card[1].failedJobs = json.data.failedVolumeRebuildTasksCounts
              }else{
                  _this.$message.error(json.message);
                  return false
              }

          }).catch(function (error) {
              console.log(error);
          });
      }
    },
    mounted() {
      this.getData()
    },
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
    padding: 0.3rem 30% 0.05rem 9%;
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
    padding: 0.8rem 9% 0.4rem;
    .el-row /deep/{
      padding: 0 0.17rem;
      .el-col{
        width: calc(50% - 0.54rem);
        margin: 0 0.27rem;
        .box-card {
          width: 100%;
          box-shadow: 0 4px 10px 0px rgba(0, 0, 0, 0.15);
          border: 1px solid #b1b1b1;
          border-radius: 0.06rem;
          color: #333;
          .el-card__header{
            padding: 0.15rem 0;
            border-bottom: 1px solid #b1b1b1;
            .clearfix{
              font-size: 0.23rem;
              line-height: 1.1;
              text-align: center;
                @media screen and (max-width:600px){
                  font-size: 12px;
                }
            }
          }
          .el-card__body{
            padding: 0 15% 0.57rem;
            .statusStyle{
              padding: 0.1rem 0 0.7rem;
              .list{
                display: flex;
                align-items: center;
                position: relative;
                // padding: 0 0 0 26%;
                margin: 0.3rem 0 0;
                font-size: 0.14rem;
                line-height: 0.25rem;
                @media screen and (max-width: 441px) {
                    font-size: 14px;
                    line-height: 1.5;
                }
                span, svg{
                  width: 24px;
                  height: 24px;
                  margin: 0 10% 0 0;
                  font-size: 22px;
                  color: #f8b551;
                  @media screen and (max-width: 441px) {
                      width: 20px;
                      height: 20px;
                      margin: 0 4% 0 0;
                      font-size: 20px;
                  }
                }
                // &::before{
                //   position: absolute;
                //   content: '';
                //   left: 0.17rem;
                //   top: 0.06rem;
                //   width: 0.13rem;
                //   height: 0.13rem;
                //   border-radius: 100%;
                //   background: #f8b551;
                // }
                &:nth-child(2){
                  span{
                    color: #89c997;
                  }
                }
                &:nth-child(3){
                  span{
                    color: #ff0000;
                  }
                }
              }
            }
            .el-button{
              width: 100%;
              padding: 0.1rem 0;
              font-size: 0.18rem;
              font-family: inherit;
              line-height: 0.3rem;
              text-align: center;
              border-radius: 0.06rem;
              color: #333;
              background: transparent;
              border: 1px solid;
              &:hover{
                color: #fff;
                background: #7ecef4;
                border: 1px solid;
              }
                @media screen and (max-width:600px){
                  font-size: 12px;
                }
            }
          }
        }
        // &:nth-child(2){
        //   .box-card{
        //     .el-card__body{
        //       .el-button{
        //         color: #333;
        //         background: transparent;
        //         border: 1px solid;
        //       }
        //     }
        //   }
        // }
      }
    }
  }
}

@media screen and (max-width:999px){
  .fs3_back{
    .fs3_cont {
      .el-row /deep/ {
        .el-col{
            width: 95%;
            margin: 0 auto 0.5rem;
        }
      }
    }
  }
}
@media screen and (max-width:600px){

}
</style>
