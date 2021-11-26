<template>
    <div class="fs3_back">
      <div class="fs3_head">
        <div class="fs3_head_text">
          <div class="titleBg">{{linkTitle}}</div>
          <h1>{{linkTitle}}</h1>
        </div>
        <img src="@/assets/images/page_bg.png" class="bg" alt="">
      </div>
      <div class="fs3_cont">
        <el-breadcrumb separator-class="el-icon-right">
          <el-breadcrumb-item :to="{ name: 'my_account_dashboard' }">Dashboard</el-breadcrumb-item>
          <el-breadcrumb-item>{{linkTitle}}</el-breadcrumb-item>
        </el-breadcrumb>
        <el-table
          :data="tableData" stripe empty-text="No data" v-if="$route.params.type == 'backup_job'"
          style="width: 100%">
          <el-table-column prop="backupTaskId" label="Backup ID">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0]? scope.row.backupPlanTasks[0].backupTaskId : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="Date" label="Date">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data? scope.row.backupPlanTasks[0].data[0].date : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="miner_id" label="W3SSID">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data? scope.row.backupPlanTasks[0].data[0].miner_id : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="cost" label="Price">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data? scope.row.backupPlanTasks[0].data[0].cost : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="deal_cid" label="Deal CID">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data? scope.row.backupPlanTasks[0].data[0].deal_cid : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="payload_cid" label="Data CID">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data? scope.row.backupPlanTasks[0].data[0].payload_cid : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="Duration">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data? scope.row.backupPlanTasks[0].data[0].duration : '' }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="Status">
            <template slot-scope="scope">
                <div class="statusStyle"
                      v-if="scope.row.backupPlanTasks[0].status == 'Created'"
                      style="color: #0a318e">
                    {{ scope.row.backupPlanTasks[0].status }}
                </div>
                <div class="statusStyle"
                      v-else-if="scope.row.backupPlanTasks[0].status == 'Running'"
                      style="color: #ffb822">
                    {{ scope.row.backupPlanTasks[0].status }}
                </div>
                <div class="statusStyle"
                      v-else-if="scope.row.backupPlanTasks[0].status == 'Completed'"
                      style="color: #1dc9b7">
                    {{ scope.row.backupPlanTasks[0].status }}
                </div>
                <div class="statusStyle" v-else style="color: rgb(255, 184, 34)">
                    {{ scope.row.backupPlanTasks[0].status }}
                </div>
            </template>
          </el-table-column>
          <el-table-column prop="" label="">
            <template slot-scope="scope">
              <el-button v-if="scope.row.backupPlanTasks[0].status != 'Completed'"
                type="info"
                @click="dialogDis=true">Rebuild Image</el-button>
              <el-button v-else
                type="primary"
                @click="detailFun(scope.row)">Rebuild Image</el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-table
          :data="tableData_2" stripe empty-text="No data" v-else
          style="width: 100%">
          <el-table-column prop="rebuildTaskID" label="Rebuild Job ID"></el-table-column>
          <el-table-column prop="status" label="Status">
            <template slot-scope="scope">
                <div class="statusStyle"
                      v-if="scope.row.status == 'Created'"
                      style="color: #0a318e">
                    {{ scope.row.status }}
                </div>
                <div class="statusStyle"
                      v-else-if="scope.row.status == 'Running'"
                      style="color: #ffb822">
                    {{ scope.row.status }}
                </div>
                <div class="statusStyle"
                      v-else-if="scope.row.status == 'Completed'"
                      style="color: #1dc9b7">
                    {{ scope.row.status }}
                </div>
                <div class="statusStyle" v-else style="color: rgb(255, 184, 34)">
                    {{ scope.row.status }}
                </div>
            </template>
          </el-table-column>
          <el-table-column prop="miner_id" label="W3SSID"></el-table-column>
          <el-table-column prop="payload_cid" label="Data CID"></el-table-column>
          <el-table-column prop="backupTaskId" label="Backup ID"></el-table-column>
          <el-table-column prop="createdOn" label="Date Created"></el-table-column>
          <el-table-column prop="updatedOn" label="Date Updated"></el-table-column>
        </el-table>
      </div>

      <el-dialog
        title="Rebuild Image" custom-class="formStyle" 
        :visible.sync="dialogVisible"
        :width="dialogWidth">
        <img src="@/assets/images/small_bell.png" class="icon" alt="">
        <span class="span">Are you sure you want to rebuild volume from “<b>{{backupPlan.backupPlanName}}</b> at {{backupPlan.date}} ”?</span>
        <span class="span">This action will overwrite your existing file system,</span>
        <span class="span"><b>Proceed?</b></span>
        <div slot="footer" class="dialog-footer">
          <el-button @click="dialogVisible=false">Cancel</el-button>
          <el-button @click="goLink">Backup Current System</el-button>
          <el-button @click="confirm">OK</el-button>
        </div>
      </el-dialog>

      <el-dialog
        title="Rebuild Image" custom-class="formStyle" 
        :visible.sync="dialogConfirm"
        :width="dialogWidth">
        <img src="@/assets/images/check_sign.png" class="icon" alt="">
        <span class="span">Your rebuild has created successfully</span>
        <br>
        <el-card class="box-card">
          <div class="statusStyle">
            <div class="list"><span>Rebuild Job ID: </span> {{backupPlan.backupPlanTasks[0].backupTaskId}}</div>
            <div class="list"><span>Date Created:</span> {{backupPlan.date}}</div>
            <div class="list"><span>Backup ID:</span> {{backupPlan.backupPlanId}} </div>
            <div class="list"><span>Data CID:</span> {{backupPlan.payload_cid}} </div>
          </div>
        </el-card>
        <div slot="footer" class="dialog-footer">
          <el-button class="active" @click="handleClose">OK</el-button>
        </div>
      </el-dialog>

      <el-dialog
        title="Rebuild Image" custom-class="formStyle" 
        :visible.sync="dialogDis"
        :width="dialogWidth">
        <span class="span">Rebuild image is accessible only when status is completed</span>
        <div slot="footer" class="dialog-footer">
          <el-button class="active" @click="dialogDis=false">OK</el-button>
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
          dialogIndex: 0,
          dialogVisible: false,
          dialogConfirm: false,
          dialogDis: false,
          width: document.body.clientWidth>600?'400px':'95%',
          ruleForm: {
            name: 'ceshi',
            price: '1',
            duration: '1',
            verified: '2',
            fastRetirval: '1',
            frequency: 'Backup Daily',
            region: 'Global',
          },
          linkTitle: 'All Backup Job Detalls',
          tableData: [],
          tableData_2: [],
          backupPlan: {
            backupPlanName: '-',
            date: '-',
            "backupPlanId": '',
            "backupPlanTasks": [
                {
                    "backupTaskId": '-'
                }
            ],
          }
        }
    },
    watch: {},
    methods: {
      goLink() {
        this.$router.push({name: 'my_account_backupPlans'})
      },
      confirm() {
        this.dialogVisible = false
        this.dialogConfirm = true
      },
      planSubmit(index) {
        console.log(index)
        this.dialogIndex = index
        this.dialogVisible = true
      },
      handleClose(done) {
        this.dialogConfirm = false
      },
      detailFun(row) {
        this.dialogVisible = true
        this.backupPlan = row
        this.backupPlan.date = row.backupPlanTasks[0].data[0].date
        this.backupPlan.payload_cid = row.backupPlanTasks[0].data[0].payload_cid
      },
      productName() {
        let _this = this
        let paramsType = _this.$route.params.type
        if(paramsType == 'backup_job') {
            _this.linkTitle = 'All Backup Job Detalls'
            _this.getData(1)
        }else {
            _this.linkTitle = 'All Rebuild Job Detalls'
            _this.getData()
        }
      },
      getData(type) {
        let _this = this
        let postUrl = ''

        if(type){
          postUrl = _this.data_api + `/minio/backup/retrieve/volume`

          axios.get(postUrl, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
          }}).then((response) => {
              let json = response.data
              if (json.status == 'success') {
                _this.tableData = json.data.volumeBackupPlans
                _this.tableData.map(item => {
                  if(!item.data) return false
                  item.data.map(child => {
                    child.date = 
                      child.date?
                          child.date.length<13 ?
                              moment(new Date(parseInt(child.date * 1000))).format("YYYY-MM-DD HH:mm:ss") :
                              moment(new Date(parseInt(child.date))).format("YYYY-MM-DD HH:mm:ss")
                          :
                          '-'
                    child.duration = 
                      child.duration?
                          child.duration.length<13 ?
                              moment(new Date(parseInt(child.duration * 1000))).format("YYYY-MM-DD HH:mm:ss") :
                              moment(new Date(parseInt(child.duration))).format("YYYY-MM-DD HH:mm:ss")
                          :
                          '-'
                  })
                })
              }else{
                  _this.$message.error(json.message);
                  return false
              }

          }).catch(function (error) {
              console.log(error);
          });
        }else{
          postUrl = _this.data_api + `/minio/rebuild/retrieve/volume`

          axios.get(postUrl, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
          }}).then((response) => {
              let json = response.data
              if (json.status == 'success') {
                _this.tableData_2 = json.data.volumeRebuildTasks
                _this.tableData_2.map(item => {
                    item.createdOn = moment(new Date(parseInt(item.createdOn / 1000))).format("YYYY-MM-DD HH:mm:ss")
                    item.updatedOn = moment(new Date(parseInt(item.updatedOn / 1000))).format("YYYY-MM-DD HH:mm:ss")
                })
              }else{
                  _this.$message.error(json.message);
                  return false
              }

          }).catch(function (error) {
              console.log(error);
          });
        }
      }
    },
    watch: {
        $route: function (to, from) {
            this.productName()
        }
    },
    mounted () {
      this.productName()
    },
};
</script>

<style lang="scss" scoped>
.el-dialog__wrapper /deep/ {  
    display: flex;
    align-items: center;
    left: 3.2rem;
    background: url('../../../assets/images/page_bg01.png') no-repeat center 16vh;
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
        padding: 0.3rem 10%;
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
        .icon{
          display: block;
          width: 0.2rem;
          margin: 0 auto 0.15rem;
        }
        .span{
          display: block;
          margin: 0.1rem auto 0;
          font-size: 0.16rem;
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
          margin: 0 3%;
          padding: 0 0.2rem;
          font-size: 0.14rem;
          font-family: 'm-regular';
          line-height: 2.3;
          text-align: center;
          border-radius: 0.06rem;
          color: #333;
          background: transparent;
          border: 1px solid;
                @media screen and (max-width:600px){
                  font-size: 16px;
                }
          &:last-child, &:hover{
            color: #fff;
            background: #84d088;
            border: 1px solid;
          }
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
    padding: 0.5rem 9% 0.2rem 9%;
    background: #7ecef4;
    color: #fff;
    .bg{
      position: absolute;
      right: 18%;
      width: 9%;
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
    padding: 0 0 0.4rem;
    .el-breadcrumb /deep/{
      display: flex;
      justify-content: flex-end;
      padding: 0 9%;
      .el-breadcrumb__item{
        line-height: 0.37rem;
        font-size: 0.14rem;
        .el-breadcrumb__separator{
          color: #333;
        }
        .el-breadcrumb__inner{
          color: #2f85e5;
          font-weight: normal;
        }
        .is-link{
          color: #333;
          &:hover{
            text-decoration: underline;
            color: #2f85e5;
          }
        }
      }
    }
    .el-table /deep/{
      overflow: visible;
      td,th{
        .cell{
          text-align: center;
          font-size: 0.14rem;
          color: #333;
          .el-button{
            margin: 0 auto 0;
            padding: 0 0.07rem;
            font-size: 0.14rem;
            font-family: 'm-regular';
            line-height: 2.2;
            color: #fff;
            text-align: center;
            border-radius: 0.06rem;
          }
          .el-button--primary{
            background: #7ecef4;
            border: 1px solid #7ecef4;
          }
          .statusStyle {
            display: inline-block;
            border: 1px solid;
            padding: 0 0.05rem;
            border-radius: 0.05rem;
            line-height: 0.28rem;
            // color: inherit !important;
          }
        }
      }
      th{
          padding: 0.2rem 0;
        font-size: 0.18rem;
        font-weight: bold;
        background: #e0eef4
      }
      .el-table__row--striped{
        td{
          background: #eee;
        }
      }
    }
  }
}

@media screen and (max-width:999px){
  .fs3_back{
    .fs3_head{
      padding: 0.5rem 2% 0.2rem 2%;
      .bg{
        top: 0.2rem;
      }
    }
    .fs3_cont {
      .el-breadcrumb /deep/{
        padding: 0.15rem 9%;
        .el-breadcrumb__item{
          font-size: 13px;
        }
      }
    }
  }
}
@media screen and (max-width:600px){
  .fs3_back{
    .fs3_head{
      .bg{
        top: 0.6rem;
        right: 0.2rem;
      }
    }
  }
}
</style>
