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
                <div class="list"><span class="el-icon-loading"></span> In process jobs: {{item.inProcessJobs}}</div>
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
          axios.get(_this.data_api + `/minio/backup/retrieve/volume`, {headers: {
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


          axios.get(_this.data_api + `/minio/rebuild/retrieve/volume`, {headers: {
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
      right: 13%;
      width: 14%;
      top: 0.3rem;
      z-index: 5;
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
                span{
                  margin: 0 10% 0 0;
                  font-size: 0.22rem;
                  color: #f8b551;
                  @media screen and (max-width: 441px) {
                      margin: 0 4% 0 0;
                      font-size: 17px;
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
              font-family: 'm-regular';
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
