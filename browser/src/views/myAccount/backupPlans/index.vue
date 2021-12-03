<template>
    <div class="fs3_back">
      <div class="fs3_head">
        <div class="fs3_head_text">
          <div class="titleBg">Backup Plans</div>
          <h1>Backup Plans</h1>
        </div>
        <img src="@/assets/images/page_bg.png" class="bg" alt="">
      </div>
      <div class="fs3_cont">
        <el-form :model="ruleForm" :rules="rules" ref="ruleForm" class="demo-ruleForm" v-loading="loading">
          <el-form-item label="Backup plan name:" prop="name">
            <el-input v-model="ruleForm.name"></el-input>
          </el-form-item>
          <el-form-item label="Choose your backup frequency:" prop="frequency">
            <el-select v-model="ruleForm.frequency" placeholder="">
              <el-option
                v-for="item in ruleForm.frequencyOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item>
          <!-- <el-form-item label="Choose your backup region:" prop="region">
            <el-select v-model="ruleForm.region" placeholder="">
              <el-option
                v-for="item in ruleForm.regionOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value">
              </el-option>
            </el-select>
          </el-form-item> -->
          <el-form-item label="Price:" prop="price">
            <el-input v-model="ruleForm.price" class="input" onkeyup="value=value.replace(/^\D*(\d*(?:\.\d{0,20})?).*$/g, '$1')"></el-input> FIL
          </el-form-item>
          <el-form-item label="Duration:" prop="duration">
            <el-input v-model="ruleForm.duration" class="input" onkeyup="value=value.replace(/^(0+)|[^\d]+/g,'')"></el-input> Days
          </el-form-item>
          <el-form-item label="Verified-Deal:" prop="verified">
            <el-radio v-model="ruleForm.verified" label="1">Yes</el-radio>
            <el-radio v-model="ruleForm.verified" label="2">No</el-radio>
          </el-form-item>
          <el-form-item label="Fast-Retrival:" prop="fastRetirval">
            <el-radio v-model="ruleForm.fastRetirval" label="1">Yes</el-radio>
            <el-radio v-model="ruleForm.fastRetirval" label="2">No</el-radio>
          </el-form-item>
          <el-form-item>
            <el-button @click="submitForm('ruleForm')">Create</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-dialog
        :title="ruleForm.name" custom-class="formStyle"
        :visible.sync="dialogVisible"
        :width="dialogWidth">
        <el-card class="box-card">
          <div class="statusStyle">
            <div class="list"><span>Backup frequency:</span> {{ruleForm.frequency == '1'?'Backup Daily':'Backup Weekly'}}</div>
            <!-- <div class="list"><span>Backup region:</span> {{ruleForm.region}}</div> -->
            <div class="list"><span>Price:</span> {{ruleForm.price}} FIL</div>
            <div class="list"><span>Duration:</span> {{ruleForm.duration}} days</div>
            <div class="list"><span>Verified deal:</span> {{ruleForm.verified == '2'? 'No' : 'Yes'}}</div>
            <div class="list"><span>Fast retrieval:</span> {{ruleForm.fastRetirval == '2'? 'No' : 'Yes'}}</div>
          </div>
        </el-card>
        <div slot="footer" class="dialog-footer">
          <el-button @click="confirm">OK</el-button>
        </div>
      </el-dialog>

      <el-dialog
        title="Backup Plans" custom-class="formStyle"
        :visible.sync="dialogConfirm"
        :width="dialogWidth">
        <span class="span">Your backup has created successfully</span>
        <div slot="footer" class="dialog-footer">
          <router-link :to="{name: 'my_account_myPlans'}">VIEW</router-link>
          <el-button @click="dialogConfirm=false">OK</el-button>
        </div>
      </el-dialog>
    </div>
</template>

<script>
import axios from 'axios'
let that
export default {
    data() {
        var validateDuration = (rule, value, callback) => {
            if (!value) {
                return callback(new Error('Please enter the duration'));
            }
            setTimeout(() => {
                if (value < 180) {
                    callback(new Error('The duration must be equal to or greater than 180 days.'));
                } else {
                    callback();
                }
            }, 100);
        };
        return {
          width: document.body.clientWidth>600?'400px':'95%',
          dialogWidth: document.body.clientWidth<=600?'95%':'50%',
          dialogVisible: false,
          dialogConfirm: false,
          ruleForm: {
            name: '',
            price: '',
            duration: '',
            verified: '2',
            fastRetirval: '1',
            frequency: '1',
            frequencyOptions: [{
              value: '1',
              label: 'Backup Daily'
            },{
              value: '7',
              label: 'Backup Weekly'
            }],
            region: 'Global',
            regionOptions: [{
              value: 'Global',
              label: 'Global'
            },{
              value: 'Asia',
              label: 'Asia'
            },{
              value: 'Africa',
              label: 'Africa'
            },{
              value: 'North America',
              label: 'North America'
            },{
              value: 'Sorth America',
              label: 'Sorth America'
            },{
              value: 'Europe',
              label: 'Europe'
            },{
              value: 'Oceania',
              label: 'Oceania'
            }]
          },
          rules: {
              name: [
                { required: true, message: 'Please enter Backup plan Name', trigger: 'blur' }
              ],
              price: [
                { required: true, message: 'Please enter Price', trigger: 'blur' }
              ],
              duration: [
                  { validator: validateDuration, trigger: 'blur'}
              ],
          },
          loading: false
        }
    },
    watch: {},
    methods: {
      confirm() {
        this.dialogVisible = false
        this.dialogConfirm = true
      },
      submitForm(formName) {
        let _this = this
        _this.$refs[formName].validate((valid) => {
          if (valid) {
            _this.loading = true
            let minioDeal = {
              "BackupPlanName": _this.ruleForm.name,
              "BackupInterval": _this.ruleForm.frequency,      //unit in day
              "Price": _this.ruleForm.price,          //unit in FIL
              "Duration":  String(_this.ruleForm.duration*24*60*2),   //（The number of days entered by the user on the UI needs to be converted into epoch to the backend. For example, 10 days is 10*24*60*2）
              "VerifiedDeal": _this.ruleForm.verified == '2'? false : true,
              "FastRetrieval": _this.ruleForm.fastRetirval == '2'? false : true
            }

            let postUrl = _this.data_api + `/minio/backup/add/plan`

            axios.post(postUrl, minioDeal, {headers: {
                  'Authorization':"Bearer "+ _this.$store.getters.accessToken
            }}).then((response) => {
                _this.loading = false
                let json = response.data
                if (json.status == 'success') {
                  _this.dialogVisible = true
                }else{
                    _this.$message.error(json.message);
                    return false
                }

            }).catch(function (error) {
                _this.loading = false
                console.log(error);
            });

          } else {
            console.log('error submit!!');
            return false;
          }
        });
      },
    },
    mounted() {
      that = this
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
        .el-button, a{
          display: block;
          margin: 0 5%;
          padding: 0 0.2rem;
          font-size: 0.14rem;
          font-family: 'm-regular';
          line-height: 2.3;
          color: #fff;
          text-align: center;
          border-radius: 0.06rem;
          background: #84d088;
          border: 1px solid #84d088;
                @media screen and (max-width:600px){
                  font-size: 16px;
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
    padding: 0.3rem 30% 0.05rem 9%;
    background: #7ecef4;
    color: #fff;
    .bg{
      position: absolute;
      right: 18%;
      width: 13%;
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
    padding: 0.5rem 16%;
    .el-form /deep/{
      width: 100%;
      padding: 0.5rem 0 0.4rem;
      background: #eeeeee;
      .el-form-item{
        display: flex;
        justify-content: center;
        margin-bottom: 0.25rem;
        .el-form-item__label{
          display: block;
          width: 38%;
          padding: 0 2% 0 0;
          font-size: 0.14rem;
          color: #333;
        }
        .el-form-item__content{
          color: #333;
          width: 60%;
          font-size: 0.14rem;
          .el-radio{
            .el-radio__label{
              font-size: inherit;
            }
          }
          .el-select{
            width: 95%;
            max-width: 440px;
            .el-input{
              width: 100%;
              font-size: inherit;
            }
          }
          .el-input{
            width: 95%;
            max-width: 440px;
            .el-input__inner{
              font-size: inherit;
              // height: 0.35rem;
              // line-height: 0.35rem;
            }
          }
          .input{
            width: calc(100% - 40px);
            max-width: 250px;
          }
          .el-button{
            display: block;
            margin: 0.4rem auto 0;
            padding: 0.1rem 0.2rem;
            font-size: 0.18rem;
            font-family: 'm-regular';
            line-height: 1.2;
            color: #fff;
            text-align: center;
            border-radius: 0.06rem;
            background: #7ecef4;
            border: 1px solid #7ecef4;
          }
        }
        &:last-child{
          .el-form-item__content{
            width: 100%;
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
  .fs3_back{
    .fs3_cont{
        padding: 0.8rem 4%;
        .el-form /deep/ {
          .el-form-item{
            flex-wrap: wrap;
            padding: 0 5%;
            .el-form-item__label{
              width: 100%;
              text-align: left;
            }
            .el-form-item__content{
              .el-button{
                font-size: 16px;
              }
            }
          }
        }
    }
  }
}
</style>
