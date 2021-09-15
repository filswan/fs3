<template>
      <el-dialog title="" top="50px" :visible.sync="shareDialog" :custom-class="{'ShareObjectMobile': shareFileShowMobile, 'ShareObject': 1 === 1}" :before-close="getDiglogChange">
          <div class="shareContent">
              <el-row class="share_left" v-if="shareObjectShow">
                <div class="qrcode" id="qrcode" ref="qrCodeUrl"></div>
                <el-col :span="24" style="margin-bottom: 0.45rem;">
                  <h4>Share Object</h4>
                </el-col>
                <el-col :span="24">
                  <h5>Shareable Link</h5>
                  <el-input v-model="share_input" placeholder="" id="url-link" disabled></el-input>
                </el-col>
                <el-col :span="24">
                  <h5>Expires in (Max 7 days)</h5>
                  <el-row class="steppet">
                    <el-col :span="8" v-if="num">
                      <div class="set-expire-title">Days</div>
                      <el-input-number v-model="num.num_Day" @change="handleChange" :min="1" :max="7" label="Days"></el-input-number>
                    </el-col>
                    <el-col :span="8" v-if="num">
                      <div class="set-expire-title">Hours</div>
                      <el-input-number v-model="num.num_Hours" @change="handleChange" :min="0" :max="23" label="Hours"></el-input-number>
                    </el-col>
                    <el-col :span="8" v-if="num">
                      <div class="set-expire-title">Minutes</div>
                      <el-input-number v-model="num.num_Minutes" @change="handleChange" :min="0" :max="59" label="Minutes"></el-input-number>
                    </el-col>
                  </el-row>
                </el-col>
                <el-col :span="24">
                  <div class="btncompose">
                    <el-button @click="copyLink(share_input)">Copy Link</el-button>
                    <el-button @click="getDiglogChange">Cancel</el-button>
                  </div>
                </el-col>
              </el-row>


              <el-row class="share_right" v-if="shareFileShow">
                <el-button class="shareFileCoinSend" @click="submitForm('ruleForm')">Send</el-button>
                <el-col :span="24">
                  <!--h4 v-if="shareObjectShow">Share to Filecoin</h4-->
                  <h4>Backup to Filecoin</h4>
                </el-col>
                <el-col :span="24">
                  <el-form :model="ruleForm" :rules="rules" ref="ruleForm" label-width="110px" class="demo-ruleForm">
                    <el-form-item label="Provider ID:" prop="minerId">
                      <!--el-input v-model="ruleForm.minerId"></el-input-->
                      <el-menu :default-active="'1'" menu-trigger="click" class="el-menu-demo" mode="horizontal" @open="handleOpen" @close="handleClose" :unique-opened="true">
                          <el-submenu index="1" popper-class="myMenu" :popper-append-to-body="false">
                              <template slot="title">
                                  {{ name }}
                              </template>
                              <el-submenu :index="'1-'+n" v-for="(item, n) in locationOptions" :key="n" :attr="'1-'+n">
                                  <template slot="title">
                                      <span>{{ item.value }}</span>
                                  </template>
                                  <el-menu-item :index="'1-'+n+'-1'" :attr="'1-'+n+'-1'">
                                      <!-- <el-table :cell-class-name="tableCellClassName" @cell-click="cellClick" ref="multipleTable" :data="tableData" v-loading="loading" tooltip-effect="dark" style="width: 100%" @selection-change="handleSelectionChange"> -->
                                      <el-table ref="singleTable" :cell-class-name="tableCellClassName" @cell-click="cellClick" :data="tableData" v-loadmore="loadMore" v-loading="loading" height="255" highlight-current-row @current-change="handleCurrentChange" style="width: 100%">
                                          <el-table-column type="index" width="40">
                                              <template  slot-scope="scope">
                                                  <el-radio v-model="radio" :label="'1-'+n+'-'+scope.$index"></el-radio>
                                              </template>
                                          </el-table-column>
                                          <el-table-column property="miner_id" label="W3SS ID"></el-table-column>
                                          <el-table-column property="status" label="Status"></el-table-column>
                                          <el-table-column property="score" label="Score"></el-table-column>
                                      </el-table>
                                  </el-menu-item>
                              </el-submenu>
                          </el-submenu>
                      </el-menu>
                    </el-form-item>
                    <el-form-item label="Price:" prop="price">
                      <el-input v-model="ruleForm.price" onkeyup="value=value.replace(/^\D*(\d*(?:\.\d{0,20})?).*$/g, '$1')"></el-input> FIL
                    </el-form-item>
                    <el-form-item label="Duration:" prop="duration">
                      <el-input v-model="ruleForm.duration" onkeyup="value=value.replace(/^(0+)|[^\d]+/g,'')"></el-input> Day
                    </el-form-item>
                    <el-form-item label="Verified-Deal:" prop="verified">
                      <el-radio v-model="ruleForm.verified" label="1">True</el-radio>
                      <el-radio v-model="ruleForm.verified" label="2">False</el-radio>
                    </el-form-item>
                    <el-form-item label="Fast-Retrival:" prop="fastRetirval">
                      <el-radio v-model="ruleForm.fastRetirval" label="1">True</el-radio>
                      <el-radio v-model="ruleForm.fastRetirval" label="2">False</el-radio>
                    </el-form-item>
                  </el-form>
                </el-col>
                <el-col :span="24">
                   <h4 style="margin: 0;">Deal CID <i class="el-icon-document-copy" v-if="ruleForm.dealCID" @click="copyLink(ruleForm.dealCID)"></i></h4>
                 </el-col>
                 <el-col :span="24">
                   <el-input
                     type="textarea"
                     :rows="4"
                     placeholder=""
                     v-model="ruleForm.dealCID"
                     disabled>
                   </el-input>
                 </el-col>
              </el-row>
          </div>
      </el-dialog>

</template>

<script>
import axios from 'axios'
import QRCode from 'qrcodejs2'
export default {
    data() {
        return {
            postUrl: this.data_api + `/minio/webrpc`,
            shareFileShowMobile: false,
            ruleForm: {
              minerId: '',
              price: '',
              duration: '',
              verified: '2',
              fastRetirval: '1',
              textarea: 'lotus client',
              dealCID: '',
              loadSign: true,
              page: 0,
              total: 1
            },
            rules: {
               minerId: [
                 { required: true, message: 'Please enter Miner ID', trigger: 'blur' }
               ],
               price: [
                 { required: true, message: 'Please enter Price', trigger: 'blur' }
               ],
               duration: [
                 { required: true, message: 'Please enter Duration', trigger: 'blur' }
               ],
            },
            locationOptions: [
                {
                    value: "Global",
                    title: '1-0'
                },
                {
                    value: "Asia",
                    title: '1-1'
                },
                {
                    value: "Africa",
                    title: '1-2'
                },
                {
                    value: "North America",
                    title: '1-3'
                },
                {
                    value: "South America",
                    title: '1-4'
                },
                {
                    value: "Europe",
                    title: '1-5'
                },
                {
                    value: "Oceania",
                    title: '1-6'
                },
            ],
            tableData: [],
            loading: false,
            bodyWidth: document.documentElement.clientWidth < 1024 ? true : false,
            name: 'Please select Provider ID',
            parentLi: '',
            parentName: '',
            radio: '1'
        }
    },
    props: ['shareDialog','shareObjectShow','shareFileShow', 'num', 'share_input', 'postAdress', 'sendApi'],
    watch: {
      'shareDialog': function(){
        let _this = this
        if(!_this.shareDialog){
          _this.ruleForm = {
              minerId: '',
              price: '',
              duration: '',
              verified: '2',
              fastRetirval: '1',
              dealCID: ''
          }
        }
      },
      share_input: function(){
         let _this = this
          this.$nextTick(function () {
            _this.creatQrCode();
          })
      },
    },
    methods: {
      creatQrCode() {
          let _this = this
          document.getElementById("qrcode").innerHTML = ''
          let qrcode = new QRCode(_this.$refs.qrCodeUrl, {
              text: _this.share_input, // Content to be converted to QR code
              width: 100,
              height: 100,
              colorDark: '#000000',
              colorLight: '#ffffff',
              correctLevel: QRCode.CorrectLevel.L
          })
      },
      shareFileShowFun() {
        this.shareFileShow = !this.shareFileShow
        this.shareFileShowMobile = !this.shareFileShowMobile
      },
      submitForm(formName) {
        this.$refs[formName].validate((valid) => {
          if (valid) {

            let _this = this
            let postUrl = ''

            if(_this.sendApi == 1){
              console.log('backup to filecoin', _this.postAdress);
              postUrl = _this.data_api + `/minio/deals/` + _this.postAdress
            }else{
              postUrl = _this.data_api + `/minio/deal/` + _this.postAdress
            }

            //let postUrl = `http://192.168.88.41:9000/minio/deal/` + _this.postAdress

            let minioDeal = {
                "VerifiedDeal": _this.ruleForm.verified == '2'? 'false' : 'true',
                "FastRetrieval": _this.ruleForm.fastRetirval == '2'? 'false' : 'true',
                "MinerId": _this.ruleForm.minerId,
                "Price": _this.ruleForm.price,
                "Duration": String(_this.ruleForm.duration*24*60*2)   //（The number of days entered by the user on the UI needs to be converted into epoch to the backend. For example, 10 days is 10*24*60*2）
            }

            axios.post(postUrl, minioDeal, {headers: {
                 'Authorization':"Bearer "+ _this.$store.getters.accessToken
            }}).then((response) => {
                let json = response.data
                if (json.status == 'success') {
                  _this.ruleForm.dealCID = json.data.dealCid
                  _this.$message({
                    message: 'Transaction has been successfully sent.',
                    type: 'success'
                  });
                }else{
                    _this.$message.error(json.message);
                    return false
                }

            }).catch(function (error) {
                console.log(error);
            });

          } else {
            console.log('error submit!!');
            return false;
          }
        });
      },
      getDiglogChange() {
        this.$emit('getshareDialog', false)
      },
      handleChange(value) {
          console.log(this.num)
          this.$emit('getShareGet', this.num)
      },
      copyLink(text){
        let _this = this
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
            var msg = successful ? 'Link copied to clipboard!' : 'copy failed!';
            console.log('Copying text command was ' + msg);
            if (successful) {
                _this.$message({
                    message: msg,
                    type: 'success'
                });
            }
        } catch (err) {
            console.log('Oops, unable to copy');
        } finally {
            document.body.removeChild(txtArea);
        }
      },
      //Provider ID select
      tableCellClassName({row, column, rowIndex, columnIndex}){
          //Use the callback method of the classname of the cell to assign a value to the row and column index
          row.index=rowIndex;
          column.index=columnIndex;
      },
      cellClick(row, column, cell, event){
          console.log(row.index);  //Select row
          let _this = this
          _this.radio = _this.parentLi + '-' + row.index
      },
      toggleSelection(rows) {
          if (rows) {
              rows.forEach(row => {
                  this.$refs.multipleTable.toggleRowSelection(row);
              });
          } else {
              this.$refs.multipleTable.clearSelection();
          }
      },
      handleSelectionChange(val) {
          this.multipleSelection = val;
          console.log('check', val)
      },
      setCurrent(row) {
          this.$refs.singleTable.setCurrentRow(row);
      },
      handleCurrentChange(val) {
          this.currentRow = val;
          console.log(val)
          if(val && val.miner_id){
              //this.name = this.parentName + ' / ' + val.miner_id
              this.ruleForm.minerId = val.miner_id
              this.name = val.miner_id
          }
      },
      handleOpen(key, keyPath) {
          let _this = this
          if(key.indexOf('-') >= 0){
              _this.tableData = []
              _this.loading = true
              _this.page = 0
              _this.total = 1
              _this.parentLi = key
              _this.locationOptions.map(item => {
                  if(item.title == key){
                      _this.parentName = item.value
                  }
              })

              let postURL = 'http://192.168.88.216:5002/miners?location='+_this.parentName+'&status=&sort_by=score&order=ascending&limit=20&offset=0'
              axios.get(postURL).then((response) => {
                  let json = response.data.data.miner
                  _this.tableData = json
                  _this.loading = false
                  _this.loadSign = true
                  if(response.data.data.total_items > 20){
                      _this.total = (response.data.data.total_items)/20
                  }
              }).catch(function (error) {
                  console.log(error);
                  _this.loading = false
              });
          }
      },
      loadMore () {
          let _this = this
          if (_this.loadSign) {
              _this.loadSign = false
              _this.page++
              if (_this.page >= _this.total) {
                  console.log('finish:', _this.page)
                  return
              }

              _this.loading = true
              let postURL = 'http://192.168.88.216:5002/miners?location='+_this.parentName+'&status=&sort_by=score&order=ascending&limit=20&offset='+_this.page*20
              axios.get(postURL).then((response) => {
                  let json = response.data.data.miner
                  json.map(item => {
                      _this.tableData.push(item)
                  })

                  _this.loading = false
                  _this.loadSign = true

              }).catch(function (error) {
                  console.log(error);
                  _this.loading = false
              });
          }
      },
      handleClose(key, keyPath) {
          //console.log('close');
      },
    },
    mounted() {},
};
</script>

<style lang="scss" scoped>
.el-dialog__wrapper /deep/{
  justify-content: center;
  display: flex;
  align-items: center;
  .ShareObject{
        position: relative;
        width: auto;
        .shareFileCoin, .shareFileCoinSend{
            position: absolute;
            top: .14rem;
            right: 0.15rem;
            padding: .05rem .1rem;
            font-size: 0.14rem;
            color: #4070ff;
            border: 0;
            background-color: #fff;
            line-height: 1.5;
            border-radius: 2px;
            text-align: center;
            transition: all;
            transition-duration: 0s;
            transition-duration: .3s;
            font-weight: normal;
            text-decoration: underline;
        }
        .shareFileCoinSend{
            color: #fff;
            background-color: #33d46f;
            text-decoration: none;
        }
        .el-dialog__header{
            display: none;
            .el-dialog__title{
                font-size: 0.15rem;
                color: #333;
            }
        }
        .el-dialog__body{
          padding: 0;
          .shareContent{
            display: flex;
            //align-items: center;
            justify-content: center;
            flex-wrap: wrap;
            width: 100%;
            .el-row{
              position: relative;
              width: 400px;
              .qrcode{
                  display: inline-block;
                  position: absolute;
                  right: 0.2rem;
                  top: 0.2rem;
                  img {
                      width: 100px;
                      height: 100px;
                      background-color: #fff;
                      padding: 0;
                      box-sizing: border-box;
                  }
              }
              .el-col{
                padding: 0 0.2rem;
                margin-bottom: 0.25rem;
                h4{
                  font-weight: normal;
                  display: block;
                  margin: 20px 0 10px;;
                  line-height: 2;
                  font-size: .15rem;
                  color: #333;
                }
                h5{
                  font-size: 13px;
                  font-weight: normal;
                  display: block;
                  margin-bottom: 10px;
                  line-height: 2;
                  color: #8e8e8e;
                }
                .el-input{
                  .el-input__inner{
                    padding: 0.1rem;
                    border: 1px solid #eee;
                    border-radius: 0.02rem;
                    font-size: 13px;
                    cursor: text;
                    transition: border-color;
                    transition-duration: .3s;
                    background-color: transparent;
                  }
                }
                .steppet{
                  display: flex;
                  justify-content: center;
                  width: 100%;
                  .el-col{
                    position: relative;
                    margin: 0;
                    .set-expire-title {
                      position: absolute;
                      top: 40px;
                      left: 0;
                      right: 0;
                      font-size: 10px;
                      text-transform: uppercase;
                      text-align: center;
                      line-height: 1.42857143;
                      color: #8e8e8e;
                    }
                    .el-input-number{
                      display: flex;
                      flex-wrap: wrap;
                      width: 100%;
                      height: 125px;
                      span, .el-input{
                        width: 100%;
                      }
                      .el-input-number__decrease{
                        background: transparent url(../assets/images/down.png) no-repeat center;
                        background-size: auto 100%;
                        height: 20px;
                        top: auto;
                        bottom: 0;
                        border: 0;
                        i{
                          display: none;
                        }
                      }
                      .el-input-number__increase{
                        bottom: auto;
                        background: transparent url(../assets/images/up.png) no-repeat center;
                        background-size: auto 100%;
                        height: 20px;
                        top: 0;
                        border: 0;
                        i{
                          display: none;
                        }
                      }
                      .el-input{
                        position: absolute;
                        top: 27px;
                        bottom: 27px;
                        border: 1px solid #eee;
                        pointer-events: none;
                        .el-input__inner{
                          position: absolute;
                          bottom: 0;
                          border: 0;
                          background-color: transparent;
                          box-shadow: none;
                          color: #333;
                          font-size: 0.2rem;
                          font-weight: 400;
                        }
                      }
                    }
                  }
                }
                .btncompose{
                  display: flex;
                  align-items: center;
                  justify-content: center;
                  .el-button{
                    padding: 0.05rem 0.1rem;
                    margin: 0 0.03rem;
                    font-size: 12px;
                    color: #fff;
                    border: 0;
                    background-color: #33d46f;
                    line-height: 1.5;
                    border-radius: 0.02rem;
                    text-align: center;
                    transition: all;
                    transition-duration: .3s;
                    &:last-child{
                      color: #545454;
                      background-color: #eee;
                    }
                  }
                }
              }
            }
            .share_right{
              padding: 0.05rem 0 .2rem;
              .el-col{
                display: flex;
                align-items: center;
                margin-bottom: 0.1rem;
                h4{
                  i{
                    margin: 0 0 0 5px;
                    cursor: pointer;
                  }
                }
                h5{
                  margin: 0 5px 0 0;
                  white-space: nowrap;
                }
                .el-input{
                  .el-input__inner{
                    padding: 0.1rem;
                    border: 1px solid #eee;
                    border-radius: 0.02rem;
                    font-size: 13px;
                    cursor: text;
                    transition: border-color;
                    transition-duration: .3s;
                    background-color: transparent;
                  }
                }
                .el-form{
                  width: 100%;
                  .el-form-item{
                    margin-bottom: 0.05rem;
                    .el-form-item__content{
                      .el-input{
                        width: calc(100% - 40px);
                      }
                    }
                  }
                  .el-form-item.is-error, .el-form-item.is-required{
                    margin-bottom: 0.15rem;
                  }

                  .el-menu{
                      float: left;
                      position: relative;
                      display: inline-block;
                      border: solid 1px #e6e6e6;
                      // width: 80%;
                      .el-submenu.is-active{
                        // width: 100%;
                      }
                      li.el-submenu{
                        .el-submenu__title{
                          &:hover{
                            color: #66b1ff;
                            background-color: #eef6ff;
                          }
                        }
                      }
                      .el-submenu__title{
                          display: flex;
                          justify-content: space-between;
                          align-items: center;
                          border: 0;
                          height: 35px;
                          line-height: 35px;
                          padding: 0 0.1rem 0 0.2rem;
                      }
                      .el-menu--horizontal{
                          .el-submenu{
                              .el-menu--horizontal{
                                  position: absolute !important;
                                  left: 200px !important;
                                  top: 0 !important;
                                  bottom: 0;
                                  .el-menu{
                                      // width: 100%;
                                      height: 100%;
                                      padding: 0;
                                      .el-menu-item{
                                          height: 100%;
                                          max-height: 300px;
                                          padding: 0;
                                          .el-table{
                                              width: 100%;
                                              max-width: 450px;
                                              height: 100%;
                                              padding: 0;
                                              .el-table__header-wrapper{
                                                .el-table__header{
                                                  width: 100% !important;
                                                }
                                              }
                                              th, td{
                                                  padding: 0.05rem 0;
                                                  background-color: #fff !important;
                                                  font-size: 0.13rem;
                                                  font-weight: 600;
                                                  line-height: 1.5;
                                                  border: 0;
                                                  .cell{
                                                      padding-right: 0;
                                                      line-height: 1.5;
                                                      .el-radio{
                                                          .el-radio__label{
                                                              display: none;
                                                          }
                                                      }
                                                      .el-checkbox{
                                                          .el-checkbox__original{
                                                              display: none;
                                                          }
                                                      }
                                                  }
                                              }
                                              th{
                                                  font-size: 0.12rem;
                                                  font-weight: normal;
                                                  .cell{
                                                      .el-checkbox{
                                                          display: none;
                                                      }
                                                  }
                                              }
                                              .el-table__body{
                                                tr{
                                                  &:hover{
                                                    td{
                                                      background-color: #eef6ff !important;
                                                    }
                                                  }
                                                }
                                              }
                                              &::before, &::after {
                                                  height: 0;
                                              }
                                          }
                                      }
                                  }
                              }
                          }
                          .is-opened{
                              .el-submenu__title{
                                  color: #66b1ff;
                                  background-color: #eef6ff;
                                  i{
                                      color: #66b1ff;
                                  }
                              }
                          }
                      }
                  }


                }
              }
              &:after{
                content: '';
                position: absolute;
                left: 0;
                top: 0;
                bottom: 0;
                width: 1px;
                height: 100%;
                background-color: #eee;
              }
            }
          }
        }
      }
    }

@media screen and (max-width:769px){

  .el-dialog__wrapper /deep/{
    .ShareObject{
      .el-dialog__body{
        .shareContent{
          .el-row{
            width: 100%;
            max-width: 500px;
          }
        }
      }
    }
  }
}
@media screen and (max-width:600px){

  .el-dialog__wrapper /deep/{
    .ShareObject{
      .el-dialog__body{
        padding: 0;
        .shareContent{
          flex-wrap: wrap;
          .el-row{
            width: 100%;
            max-width: 400px;
          }
        }
      }
    }
    .ShareObjectMobile {
      margin-top: 55vh !important;
    }
  }


}
@media screen and (max-width:441px){
  .el-dialog__wrapper /deep/{
    .ShareObject{
      .el-dialog__body{
        .shareContent{
          .share_right{
            .el-col{
              .el-form{
                .el-menu{
                  .el-menu--horizontal{
                      .el-submenu{
                         min-width: 150px;
                        .el-menu--horizontal{
                            left: 100px !important;
                        }
                      }
                  }
                }
              }
            }
          }
        }
      }
    }
  }


}
</style>
