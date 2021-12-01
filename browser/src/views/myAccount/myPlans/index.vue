
<template>
    <div class="fs3_back">
      <div class="fs3_head">
        <div class="fs3_head_text">
          <div class="titleBg">My Plans</div>
          <h1>My Plans</h1>
        </div>
        <img src="@/assets/images/page_bg.png" class="bg" alt="">
      </div>
      <div class="fs3_cont">
        <el-card class="box-card" v-for="(item, index) in plan_list" :key="index">
          <div class="title">{{ item.backupPlanName }}</div>
          <div class="button">
            <div class="statusStyle"
                  v-if="item.status == 'Created'"
                  style="color: #0a318e">
                {{ item.status }}
            </div>
            <div class="statusStyle"
                  v-else-if="item.status == 'Running'"
                  style="color: #ffb822">
                {{ item.status }}
            </div>
            <div class="statusStyle"
                  v-else-if="item.status == 'Completed'"
                  style="color: #1dc9b7">
                {{ item.status }}
            </div>
            <div class="statusStyle"
                  v-else-if="item.status == 'Stopped'"
                  style="color: #f56c6c">
                {{ item.status }}
            </div>
            <div class="statusStyle" v-else style="color: rgb(255, 184, 34)">
                {{ item.status }}
            </div>
            <el-button @click="planSubmit(index, item)" :class="{'active': dialogIndex == index}">View details</el-button>
          </div>
        </el-card>
        <div v-if="plan_list.length<=0" style="text-align: center;">No Data</div>
      </div>

      <el-dialog
        :title="ruleForm.backupPlanName" custom-class="formStyle"
        :visible.sync="dialogVisible"
        :width="dialogWidth"
        :before-close="handleClose">
        <el-card class="box-card">
          <div class="statusStyle">
            <div class="list"><span>Add backup Plan ID:</span> {{ruleForm.backupPlanId}}</div>
            <div class="list"><span>Backup frequency:</span> {{ruleForm.backupInterval == '1'?'Backup Daily':'Backup Weekly'}}</div>
            <!-- <div class="list"><span>Backup region:</span> {{ruleForm.region}}</div> -->
            <div class="list"><span>Price:</span> {{ruleForm.price}} FIL</div>
            <div class="list"><span>Duration:</span> {{ruleForm.duration/24/60/2}} days</div>
            <div class="list"><span>Verified deal:</span> {{ !ruleForm.verifiedDeal ? 'No' : 'Yes'}}</div>
            <div class="list"><span>Fast retrieval:</span> {{ !ruleForm.fastRetrieval ? 'No' : 'Yes'}}</div>
            <div class="list"><span>Create Date:</span> {{ruleForm.createdOn}}</div>
            <div class="list"><span>Last Update:</span> {{ruleForm.updatedOn}}</div>
            <div class="list"><span>Last Backup Date:</span> {{ruleForm.lastBackupOn}}</div>
          </div>
        </el-card>
        <div slot="footer" class="dialog-footer">
          <el-button 
            :type="ruleForm.status&&ruleForm.status.toLowerCase() == 'running'?'danger':'info'"
            :disabled="ruleForm.status&&ruleForm.status.toLowerCase() == 'running'?false:true"
            @click="planStatus(ruleForm)"
          >STOP</el-button>
          <el-button 
            :type="ruleForm.status&&ruleForm.status.toLowerCase() == 'running'?'info':'success'"
            :disabled="ruleForm.status&&ruleForm.status.toLowerCase() == 'running'?true:false"
            @click="planStatus(ruleForm)"
          >START</el-button>
          <el-button type="success" @click="handleClose">OK</el-button>
        </div>
      </el-dialog>
    </div>
</template>

<script>
import axios from 'axios'
import moment from "moment"
export default {
    data() {
        return {
          dialogWidth: document.body.clientWidth<=600?'95%':'50%',
          dialogIndex: NaN,
          dialogVisible: false,
          width: document.body.clientWidth>600?'400px':'95%',
          ruleForm: {},
          plan_list: []
        }
    },
    watch: {},
    methods: {
      planSubmit(index, row) {
        // console.log(index, row)
        this.dialogIndex = index
        this.dialogVisible = true
        this.ruleForm = row
      },
      handleClose(done) {
        this.dialogIndex = NaN
        this.dialogVisible = false
      },
      planStatus(row){
          let _this = this
          let params = {
            "BackupPlanId": row.backupPlanId,
            "Status": row.status
          }

          axios.post(`${_this.data_api}/minio/backup/update/plan`, params, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
          }}).then((response) => {
              let json = response.data
              if (json.status == 'success') {
                _this.ruleForm = json.data
                _this.ruleForm.createdOn = _this.ruleForm.createdOn?moment(new Date(parseInt(_this.ruleForm.createdOn / 1000))).format("YYYY-MM-DD HH:mm:ss"):'-'
                _this.ruleForm.updatedOn = _this.ruleForm.updatedOn?moment(new Date(parseInt(_this.ruleForm.updatedOn / 1000))).format("YYYY-MM-DD HH:mm:ss"):'-'
                _this.ruleForm.lastBackupOn = _this.ruleForm.lastBackupOn?moment(new Date(parseInt(_this.ruleForm.lastBackupOn / 1000))).format("YYYY-MM-DD HH:mm:ss"):'-'
                _this.getData()
              }else{
                  _this.$message.error(json.message);
                  return false
              }

          }).catch(function (error) {
              console.log(error);
          });

      },
      getData() {
          let _this = this
          let postUrl = _this.data_api + `/minio/backup/retrieve/plan`

          axios.get(postUrl, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
          }}).then((response) => {
              let json = response.data
              if (json.status == 'success') {
                _this.plan_list = json.data.volumeBackupJobPlans
                _this.plan_list.map(item => {
                    item.createdOn = item.createdOn?moment(new Date(parseInt(item.createdOn / 1000))).format("YYYY-MM-DD HH:mm:ss"):'-'
                    item.updatedOn = item.updatedOn?moment(new Date(parseInt(item.updatedOn / 1000))).format("YYYY-MM-DD HH:mm:ss"):'-'
                    item.lastBackupOn = item.lastBackupOn?moment(new Date(parseInt(item.lastBackupOn / 1000))).format("YYYY-MM-DD HH:mm:ss"):'-'
                })
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
.el-dialog__wrapper /deep/ {
    display: flex;
    align-items: center;
    left: 3.2rem;
    background: url('../../../assets/images/page_bg01.png') no-repeat center 13vh;
    background-size: 400px;
    @media screen and (max-width:600px){
      left: 0;
      background-size: 95%;
    }
    .formStyle{
      border-radius: 0.06rem;
      overflow: hidden;
      .el-dialog__header{
        padding: 0;
        line-height: 2.2;
        background: #eeeeee;
        text-align: center;
        font-size: 0.18rem;
        color: #333;
        box-shadow: 0 4px 10px 0px rgba(0, 0, 0, 0.1);
        .el-dialog__headerbtn{
          display: none;
          top: 0.2rem;
          font-size: 0.4rem;
          .el-dialog__close{
            color: #fff;
          }
        }
      }
      .el-dialog__body{
        .box-card {
          width: 95%;
          max-width: 460px;
          margin: auto;
          box-shadow: 0 4px 10px 0px rgba(0, 0, 0, 0.1);
          border-radius: 0.06rem;
          color: #333;
          .el-card__body{
            padding: 0;
            .statusStyle{
              padding: 0.1rem 10%;
              .list{
                position: relative;
                display: flex;
                margin: 0.05rem 0 0;
                font-size: 0.14rem;
                line-height: 2;
                @media screen and (max-width:600px){
                  font-size: 14px;
                }
                span{
                  display: block;
                  width: 55%;
                }
              }
            }
          }
        }
        .span{
          display: block;
          font-size: 0.18rem;
          text-align: center;
        }
      }
      .dialog-footer{
        display: flex;
        justify-content: center;
        align-items: center;
        padding: 0 0 0.1rem;
        .el-button{
          display: block;
          margin: 0 5%;
          padding: 0 0.2rem;
          font-size: 0.14rem;
          font-family: 'm-regular';
          line-height: 2.3;
          color: #fff;
          text-align: center;
          border-radius: 0.06rem;
          @media screen and (max-width:600px){
            font-size: 16px;
          }
        }
        .el-button--success{
          background: #84d088;
          border: 1px solid #84d088;
          &:hover{
            background: #8bc68e;
          }
        }
        .el-button--info{
          opacity: 0.5;
        }
      }
    }
}
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
      width: 9%;
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
    padding: 0.4rem 9%;
    .box-card /deep/{
      width: 100%;
      margin: 0 0 0.2rem;
      box-shadow: 0 4px 10px 0px rgba(0, 0, 0, 0.1);
      border-radius: 0;
      border: 1px solid #f3f3f3;
      color: #333;
      .el-card__body{
        display: flex;
        justify-content: space-between;
        padding: 0.2rem;
        .title{
          font-size: 0.18rem;
        }
        .button{
          display: flex;
          align-items: center;
          font-size: 0.14rem;
          .statusStyle {
            display: inline-block;
            border: 1px solid;
            padding: 0 0.15rem;
            margin: 0 0.1rem;
            font-size: inherit;
            border-radius: 0.05rem;
            line-height: 1.8;
            // color: inherit !important;
          }
          .el-button{
            padding: 0 0.25rem;
            font-size: inherit;
            font-family: 'm-regular';
            line-height: 1.8;
            text-align: center;
            border-radius: 0.04rem;
            color: #333;
            background: transparent;
            border: 1px solid;    
          }
          .active, .el-button:hover{
            color: #fff;
            background: #7ecef4;
            border: 1px solid;
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
.fs3_back {
  .fs3_cont{
    padding: 0.6rem 4%;
    .box-card /deep/{
       .el-card__body {
         .title, .button{
           font-size: 16px;
         }
         
       }
    }
  }
}
}
</style>
