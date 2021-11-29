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
          :data="tableData" stripe empty-text="No data" v-if="$route.params.type == 'backup_job'">
          <el-table-column prop="backupTaskId" label="Backup ID" width="90">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].backupTaskId }}
            </template>
          </el-table-column>
          <el-table-column prop="updatedOn" label="Last Updata" min-width="110">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].updatedOn }}
            </template>
          </el-table-column>
          <el-table-column prop="createdOn" label="Date Created" min-width="110">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].createdOn }}
            </template>
          </el-table-column>
          <el-table-column prop="miner_id" label="W3SSID" min-width="110">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data.dealInfo[0].miner_id }}
            </template>
          </el-table-column>
          <el-table-column prop="cost" label="Price" min-width="130">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data.dealInfo[0].cost | NumFormatPrice}} FIL
            </template>
          </el-table-column>
          <el-table-column prop="deal_cid" label="Deal CID" min-width="180">
            <template slot-scope="scope">
                <div class="hot-cold-box">
                    <el-popover
                        placement="top" width="160"
                        trigger="hover"
                        v-model="scope.row.backupPlanTasks[0].data.dealInfo[0].visible">
                        <div class="upload_form_right">
                            <p>{{scope.row.backupPlanTasks[0].data.dealInfo[0].deal_cid}}</p>
                        </div>
                        <el-button slot="reference" @click="copyTextToClipboard(scope.row.backupPlanTasks[0].data.dealInfo[0].deal_cid)">
                            <img src="@/assets/images/copy.png" alt="">
                            {{scope.row.backupPlanTasks[0].data.dealInfo[0].deal_cid}}
                        </el-button>
                    </el-popover>
                </div>
            </template>
          </el-table-column>
          <el-table-column prop="payload_cid" label="Data CID" min-width="180">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data.dealInfo[0].payload_cid }}
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="Duration" min-width="110">
            <template slot-scope="scope">
              {{ scope.row.backupPlanTasks[0].data.duration }} 
              <br>
              ({{ scope.row.backupPlanTasks[0].data.duration_time }})
            </template>
          </el-table-column>
          <el-table-column prop="status" label="Status" width="110">
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
          <el-table-column prop="" label="" width="130">
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
          :data="tableData_2" stripe empty-text="No data" v-else>
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
          <el-table-column prop="deal_cid" label="Deal CID" min-width="110">
            <template slot-scope="scope">
                <div class="hot-cold-box">
                    <el-popover
                        placement="top" width="160"
                        trigger="hover"
                        v-model="scope.row.visible">
                        <div class="upload_form_right">
                            <p>{{scope.row.deal_cid}}</p>
                        </div>
                        <el-button slot="reference" @click="copyTextToClipboard(scope.row.deal_cid)">
                            <img src="@/assets/images/copy.png" alt="">
                            {{scope.row.deal_cid}}
                        </el-button>
                    </el-popover>
                </div>
            </template>
          </el-table-column>
          <el-table-column prop="payload_cid" label="Data CID" min-width="110"></el-table-column>
          <el-table-column prop="backupTaskId" label="Backup ID"></el-table-column>
          <el-table-column prop="createdOn" label="Date Created" width="110"></el-table-column>
          <el-table-column prop="updatedOn" label="Date Updated" width="110"></el-table-column>
        </el-table>
      </div>

      <el-dialog
        title="Rebuild Image" custom-class="formStyle" 
        :visible.sync="dialogVisible"
        :width="dialogWidth">
        <img src="@/assets/images/small_bell.png" class="icon" alt="">
        <span class="span">Are you sure you want to rebuild volume from <b>{{backupPlan.backupPlanName}}</b> ?</span>
        <span class="span">This action will overwrite your existing file system,</span>
        <span class="span"><b>Proceed?</b></span>
        <div slot="footer" class="dialog-footer">
          <el-button @click="dialogVisible=false">Cancel</el-button>
          <!-- <el-button @click="goLink">Backup Current System</el-button> -->
          <el-button @click="confirm">OK</el-button>
        </div>
      </el-dialog>

      <el-dialog
        title="Rebuild Image" custom-class="formStyle" 
        :visible.sync="dialogConfirm"
        :width="dialogWidth">
        <img src="@/assets/images/check_sign.png" class="icon" alt="">
        <span class="span">Your rebuild image job has created successfully</span>
        <br>
        <el-card class="box-card">
          <div class="statusStyle">
            <div class="list"><span>Rebuild Job ID: </span> {{backupPlan.rebuildTaskID}}</div>
            <div class="list"><span>Date Created:</span> {{backupPlan.createdOn}}</div>
            <div class="list"><span>W3SSID:</span> {{backupPlan.miner_id}}</div>
            <div class="list"><span>Backup ID:</span> {{backupPlan.backupTaskId}} </div>
            <div class="list"><span>Data CID:</span> {{backupPlan.payload_cid}} </div>
            <div class="list"><span>Deal CID:</span> {{backupPlan.deal_cid}} </div>
            <div class="list">
              <span>Stauts:</span>
              <small
                    v-if="backupPlan.status == 'Created'"
                    style="color: #0a318e">
                  {{ backupPlan.status }}
              </small>
              <small
                    v-else-if="backupPlan.status == 'Running'"
                    style="color: #ffb822">
                  {{ backupPlan.status }}
              </small>
              <small
                    v-else-if="backupPlan.status == 'Completed'"
                    style="color: #1dc9b7">
                  {{ backupPlan.status }}
              </small>
              <small v-else style="color: rgb(255, 184, 34)">
                  {{ backupPlan.status }}
              </small>
            </div>
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
        let _this=this
        _this.dialogVisible = true

        let postUrl = _this.data_api + `/minio/rebuild/add/job`
        let params = {
          "BackupTaskId": row.backupPlanTasks[0].backupTaskId,
          "BackupPlanId": row.backupPlanId
        }

        axios.post(postUrl, params, {headers: {
              'Authorization':"Bearer "+ _this.$store.getters.accessToken
        }}).then((response) => {
            let json = response.data
            if (json.status == 'success') {
              _this.backupPlan = json.data
              if(_this.backupPlan.createdOn) _this.backupPlan.createdOn = moment(new Date(parseInt(_this.backupPlan.createdOn / 1000))).format("YYYY-MM-DD HH:mm:ss")
            }else{
                _this.$message.error(json.message);
                return false
            }

        }).catch(function (error) {
            console.log(error);
        });
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
      copyTextToClipboard(text) {
          let _this = this
          let saveLang = "Success";
          var txtArea = document.createElement("textarea");
          txtArea.id = 'txt';
          txtArea.style.position = 'fixed';
          txtArea.style.top = '0';
          txtArea.style.left = '0';
          txtArea.style.opacity = '0';
          txtArea.value = text;
          document.body.appendChild(txtArea);
          txtArea.select();

          try {
              var successful = document.execCommand('copy');
              var msg = successful ? 'successful' : 'unsuccessful';
              console.log('Copying text command was ' + msg);
              if (successful) {
                  _this.$message({
                      message: saveLang,
                      type: 'success'
                  });
                  return true;
              }
          } catch (err) {
              console.log('Oops, unable to copy');
          } finally {
              document.body.removeChild(txtArea);
          }
          return false;
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
                  item.backupPlanTasks.map(child => {
                    child.data.dealInfo[0].visible = false
                    child.data.duration_time = 
                      child.data.duration?
                          moment(new Date(parseInt((parseInt(child.data.duration)*30 + parseInt(1598306471)) * 1000))).format("YYYY-MM-DD HH:mm:ss")
                          :
                          '-'
                    child.createdOn = 
                      child.createdOn?
                          moment(new Date(parseInt(child.createdOn / 1000))).format("YYYY-MM-DD HH:mm:ss")
                          :
                          '-'
                    child.updatedOn = 
                      child.updatedOn?
                          moment(new Date(parseInt(child.updatedOn / 1000))).format("YYYY-MM-DD HH:mm:ss")
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
                    item.visible = false
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
    filters: {
        NumFormatPrice (value) {
            if(value == 0) return 0;
            if(!value) return '-';
            // 18 - need / 1000000000000000000
            let valueNum = String(value)
            if(value.length > 18){
                let v1 = valueNum.substring(0, valueNum.length - 18)
                let v2 = valueNum.substring(valueNum.length - 18)
                let v3 = String(v2).replace(/(0+)\b/gi,"")
                if(v3){
                    return v1 + '.' + v3
                }else{
                    return v1
                }
                return parseFloat(v1.replace(/(\d)(?=(?:\d{3})+$)/g, "$1,") + '.' + v2)
            }else{
                let v3 = ''
                for(let i = 0; i < 18 - valueNum.length; i++){
                    v3 += '0'
                }
                return '0.' + String(v3 + valueNum).replace(/(0+)\b/gi,"")
            }
        }
    }
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
                small{
                  font-size: inherit;
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
    padding: 0.3rem 9% 0.05rem 9%;
    background: #7ecef4;
    color: #fff;
    .bg{
      position: absolute;
      right: 15%;
      width: 9%;
      top: 0.2rem;
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
          word-break: break-word;
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
          .el-rate__icon{
              font-size: 0.16rem;
              margin-right: 0;
          }
          .hot-cold-box{
              .el-button{
                  width: 100%;
                  border: 0;
                  padding: 0;
                  background-color: transparent;
                  font-size: 0.1372rem;
                  word-break: break-word;
                  color: #000;
                  text-align: center;
                  line-height: 0.25rem;
                  overflow: hidden;
                  text-overflow: ellipsis;
                  white-space: normal;
                  display: -webkit-box;
                  -webkit-line-clamp: 2;
                  -webkit-box-orient: vertical;
                  span{
                      line-height: 0.25rem;
                      overflow: hidden;
                      text-overflow: ellipsis;
                      white-space: normal;
                      display: -webkit-box;
                      -webkit-line-clamp: 2;
                      -webkit-box-orient: vertical;
                  }
                  img{
                      display: none;
                      float: left;
                      width: 0.17rem;
                      margin-top: 0.03rem;
                  }
              }
              .el-button:hover{
                  img{
                      display: inline-block;
                  }
              }
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
