
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
        <el-card class="box-card" v-for="n in 3" :key="n">
          <div class="title">My backup plan name {{n}}</div>
          <el-button @click="planSubmit(n)" :class="{'active': dialogIndex == n}">View details</el-button>
        </el-card>
      </div>

      <el-dialog
        :title="ruleForm.name" custom-class="formStyle"
        :visible.sync="dialogVisible"
        :width="dialogWidth"
        :before-close="handleClose">
        <el-card class="box-card">
          <div class="statusStyle">
            <div class="list"><span>Backup frequency:</span> {{ruleForm.frequency}}</div>
            <div class="list"><span>Backup region:</span> {{ruleForm.region}}</div>
            <div class="list"><span>Price:</span> {{ruleForm.price}} FIL</div>
            <div class="list"><span>Duration:</span> {{ruleForm.duration}} days</div>
            <div class="list"><span>Verified deal:</span> {{ruleForm.verified == '2'? 'No' : 'Yes'}}</div>
            <div class="list"><span>Fast retrieval:</span> {{ruleForm.fastRetirval == '2'? 'No' : 'Yes'}}</div>
          </div>
        </el-card>
        <div slot="footer" class="dialog-footer">
          <el-button @click="handleClose">OK</el-button>
        </div>
      </el-dialog>
    </div>
</template>

<script>
export default {
    data() {
        return {
          dialogWidth: document.body.clientWidth<=600?'95%':'50%',
          dialogIndex: 0,
          dialogVisible: false,
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
        }
    },
    watch: {},
    methods: {
      planSubmit(index) {
        console.log(index)
        this.dialogIndex = index
        this.dialogVisible = true
      },
      handleClose(done) {
        this.dialogIndex = 0
        this.dialogVisible = false
      }
    },
    mounted() {},
};
</script>

<style lang="scss" scoped>
.el-dialog__wrapper /deep/ {
    display: flex;
    align-items: center;
    left: 3.2rem;
    .formStyle{
      border-radius: 0.06rem;
      overflow: hidden;
      .el-dialog__header{
        padding: 0;
        line-height: 0.4rem;
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
          line-height: 2.3;
          color: #fff;
          text-align: center;
          border-radius: 0.06rem;
          background: #84d088;
          border: 1px solid #84d088;
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
    padding: 0.5rem 39% 0.2rem 9%;
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
        .el-button{
          padding: 0 0.25rem;
          font-size: 0.14rem;
          line-height: 0.25rem;
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

@media screen and (max-width:769px){
  .fs3_back{

  }
}
@media screen and (max-width:600px){

}
</style>
