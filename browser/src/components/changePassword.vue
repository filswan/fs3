<template>
      <el-dialog title="" :visible.sync="changePass" :custom-class="{'ShareObjectMobile': shareFileShowMobile, 'ShareObject': 1 === 1}" :width='width' top="50px" :before-close="getClose">
          <div class="shareContent">
              <h4>Change Password</h4>

              <el-form :model="ruleForm" status-icon ref="ruleForm" class="demo-ruleForm">
                  <el-form-item prop="Accesskey">
                      <div class="title">Current Access Key</div>
                      <el-input type="text" v-model="ruleForm.Accesskey" autocomplete="off" :disabled="true"></el-input>
                  </el-form-item>
                  <el-form-item prop="Secretkey">
                      <div class="title">Current Secret Key</div>
                      <el-input show-password v-model="ruleForm.Secretkey" placeholder="Current Secret Key"></el-input>
                  </el-form-item>
                  <el-form-item prop="NewSecretkey">
                      <div class="title">New Secret Key</div>
                      <el-input show-password v-model="ruleForm.NewSecretkey" placeholder="New Secret Key"></el-input>
                  </el-form-item>
                  <el-form-item>
                      <div class="action">
                        <el-button type="primary" @click="generate">Generate</el-button>
                        <el-button type="success" :disabled="updateBack?true:false" @click="submitForm('ruleForm')">Update</el-button>
                        <el-button type="info" @click="getClose">Cancel</el-button>
                      </div>
                  </el-form-item>
              </el-form>

          </div>
      </el-dialog>

</template>

<script>
import axios from 'axios'
import Moment from 'moment'
export default {
    data() {
        return {
            postUrl: this.data_api + `/minio/webrpc`,
            shareFileShowMobile: false,
            width: document.documentElement.clientWidth < 1024 ? '90%' : '400px',
            ruleForm: {
                Accesskey: '',
                Secretkey: '',
                NewSecretkey: ''
            },
            rules: {
                Accesskey: [
                  { required: true, message: '', trigger: 'blur' }
                ],
                Secretkey: [
                  { required: true, message: '', trigger: 'blur' }
                ],
                NewSecretkey: [
                   { required: true, message: '', trigger: 'blur' }
                ]
            },
            updateBack: true
        }
    },
    props: ['changePass'],
    watch: {
      'changePass': function(){
        let _this = this
        if(_this.changePass){
          let assound = JSON.parse(localStorage.getItem('MinioAccountNumber'))
          _this.ruleForm.Accesskey = assound.Accesskey
          _this.ruleForm.Secretkey = assound.Secretkey
        }
      },
      'ruleForm.NewSecretkey': function(){
         this.updateBack = !this.ruleForm.NewSecretkey && this.ruleForm.NewSecretkey.length < 8 ? true : false
      }
    },
    methods: {
      getClose() {
        this.$emit('getChangePass', false)
      },
      generate(){
        let arr = new Uint8Array(40)
        window.crypto.getRandomValues(arr)
        const binStr = Array.prototype.map
          .call(arr, v => {
            return String.fromCharCode(v)
          })
          .join("")
        const base64Str = btoa(binStr)
        this.ruleForm.NewSecretkey = base64Str.replace(/\//g, "+").substr(0, 40)
      },
      submitForm(formName) {
        this.$refs[formName].validate((valid) => {
          if (valid) {
            let _this = this
            let dataSetAuth = {
                id: 1,
                jsonrpc: "2.0",
                method: "web.SetAuth",
                params:{
                    currentAccessKey: _this.ruleForm.Accesskey,
                    currentSecretKey: _this.ruleForm.Secretkey,
                    newAccessKey: _this.ruleForm.Accesskey,
                    newSecretKey: _this.ruleForm.NewSecretkey
                }
            }
            axios.post(_this.postUrl, dataSetAuth, {
               headers: {
                    'Authorization':"Bearer "+ _this.$store.getters.accessToken
               }
            }).then((response) => {
                let json = response.data
                let error = json.error
                let result = json.result
                if (error) {
                    _this.$message.error(error.message);
                    return false
                }

                _this.$message({
                  message: 'Credentials updated successfully.',
                  type: 'success'
                });

            }).catch(function (error) {
                console.log(error);
            });
          } else {
            console.log('error submit!!');
            return false;
          }
        });
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
                color: #333;
            }
        }
        .el-dialog__body{
          padding: 0;
          .shareContent{
            padding: 0 0.2rem;
            h4{
              width: 100%;
              font-weight: normal;
              display: block;
              margin: 0.2rem 0 0.3rem;
              line-height: 2;
              font-size: .17rem;
              color: #333;
            }

            .el-form {
                width: 100%;
                margin: auto;
                .el-form-item{
                    width: 100%;
                    .title{
                        color: #8e8e8e;
                        line-height: 1;
                    }
                    .el-input{
                        .el-input__inner{
                            background-color: transparent;
                            box-shadow: none;
                            border: 0;
                            border-bottom: 1px solid #ccc;
                            border-radius: 0;
                            color: #8e8e8e;
                            padding: 0;
                            font-size: 13px;
                        }
                    }
                    .action{
                      display: flex;
                      justify-content: center;
                      align-items: center;
                      width: 100%;
                      .el-button{
                          padding: 0.07rem 0.13rem;
                          border-radius: 0;
                          font-size: 13px;
                          margin: 0 0.03rem;
                          opacity: 0.9;
                          i{
                              font-size: 0.22rem;
                              font-weight: bold;
                          }
                          &:hover{
                              opacity: 1;
                          }
                      }
                    }
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
            width: 300px;
          }
        }
      }
    }
  }
}
@media screen and (max-width:600px){

  .el-dialog__wrapper /deep/{
    .ShareObject{
      width: 90%;
      .el-dialog__body{
        padding: 0;
        .shareContent{
          flex-wrap: wrap;
          .el-row{
            width: 100%;
          }
        }
      }
    }
    .ShareObjectMobile {
      margin-top: 55vh !important;
    }
  }
}
</style>
