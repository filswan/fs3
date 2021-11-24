<template>
    <div class="fs3_back">
      <div class="fs3_head">
        <div class="fs3_head_text">
          <div class="titleBg">All Backup Job Detalls</div>
          <h1>All Backup Job Detalls</h1>
        </div>
        <img src="@/assets/images/page_bg.png" class="bg" alt="">
      </div>
      <div class="fs3_cont">
        <el-breadcrumb separator-class="el-icon-arrow-right">
          <el-breadcrumb-item :to="{ name: 'my_account_dashboard' }">Dashboard</el-breadcrumb-item>
          <el-breadcrumb-item>{{linkTitle}}</el-breadcrumb-item>
        </el-breadcrumb>
        <el-table
          :data="tableData" stripe empty-text="No data" v-if="$route.params.type == 'rebuild_job'"
          style="width: 100%">
          <el-table-column prop="backupId" label="Backup ID"></el-table-column>
          <el-table-column prop="Date" label="Date"></el-table-column>
          <el-table-column prop="W3SSID" label="W3SSID"></el-table-column>
          <el-table-column prop="Price" label="Price"></el-table-column>
          <el-table-column prop="DealCID" label="Deal CID"></el-table-column>
          <el-table-column prop="DataCID" label="Data CID"></el-table-column>
          <el-table-column prop="Duration" label="Duration"></el-table-column>
          <el-table-column prop="Status" label="Status"></el-table-column>
          <el-table-column prop="Status" label="">
                    <template slot-scope="scope">
                      <el-button @click="dialogVisible=true">Rebuild Image</el-button>
                    </template>
          </el-table-column>
        </el-table>

        <el-table
          :data="tableData_2" stripe empty-text="No data" v-else
          style="width: 100%">
          <el-table-column prop="RebuildJobID" label="Rebuild Job ID"></el-table-column>
          <el-table-column prop="Status" label="Status"></el-table-column>
          <el-table-column prop="W3SSID" label="W3SSID"></el-table-column>
          <el-table-column prop="DataCID" label="Data CID"></el-table-column>
          <el-table-column prop="backupId" label="Backup ID"></el-table-column>
          <el-table-column prop="Date" label="Date Created"></el-table-column>
          <el-table-column prop="DateUpdated" label="Date Updated"></el-table-column>
        </el-table>
      </div>

      <el-dialog
        title="Rebuild Image" custom-class="formStyle"
        :visible.sync="dialogVisible"
        :width="dialogWidth">
        <span class="span">Are you sure you want to rebuild volume from“ BD111 at 2021-11-17 20:56:21 ”?</span>
        <span class="span span1">This action will overwrite your existing file system,</span>
        <span class="span span2">Proceed?</span>
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
        <span class="span">Your rebuild has created successfully</span>
        <br>
        <el-card class="box-card">
          <div class="statusStyle">
            <div class="list"><span>Rebuild Job ID: </span> {{ruleForm.frequency}}</div>
            <div class="list"><span>Date Created:</span> {{ruleForm.region}}</div>
            <div class="list"><span>Backup ID:</span> {{ruleForm.price}} </div>
            <div class="list"><span>Data CID:</span> {{ruleForm.duration}} </div>
          </div>
        </el-card>
        <div slot="footer" class="dialog-footer">
          <el-button class="active" @click="handleClose">OK</el-button>
        </div>
      </el-dialog>
    </div>
</template>

<script>
import axios from 'axios'
export default {
    data() {
        return {
          dialogWidth: document.body.clientWidth<=600?'95%':'50%',
          dialogIndex: 0,
          dialogVisible: false,
          dialogConfirm: false,
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
          tableData: [{
            backupId: 'BD111',
            date: '2016-05-02',
            W3SSID: 't03354',
            Price: '0.005FIL',
            DealCID: 'abcdfer12',
            DataCID: 'abcdfer12',
            Duration: 'abcdfer12',
            Status: 'Completed',
          },{
            backupId: 'BD112',
            date: '2016-05-02',
            W3SSID: 't03354',
            Price: '0.005FIL',
            DealCID: 'abcdfer12',
            DataCID: 'abcdfer12',
            Duration: 'abcdfer12',
            Status: 'Completed',
          },{
            backupId: 'BD113',
            date: '2016-05-02',
            W3SSID: 't03354',
            Price: '0.005FIL',
            DealCID: 'abcdfer12',
            DataCID: 'abcdfer12',
            Duration: 'abcdfer12',
            Status: 'Completed',
          }],
          tableData_2: [{
            RebuildJobID: 'BD111',
            backupId: 'BD111',
            Date: '2016-05-02',
            W3SSID: 't03354',
            DataCID: 'abcdfer12',
            DateUpdated: '2016-05-02',
            Status: 'Completed',
          }]
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
      productName() {
        let _this = this
        let paramsType = _this.$route.params.type
        if(paramsType == 'rebuild_job') {
            _this.linkTitle = 'All Backup Job Detalls'
        }else {
            _this.linkTitle = 'All Rebuild Job Detalls'
        }
      },
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
        .span1{
          font-size: 0.14rem;
        }
        .span2{
          font-weight: bold;
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
          line-height: 2.3;
          text-align: center;
          border-radius: 0.06rem;
          color: #333;
          background: transparent;
          border: 1px solid;
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
            line-height: 2.2;
            color: #fff;
            text-align: center;
            border-radius: 0.06rem;
            background: #7ecef4;
            border: 1px solid #7ecef4;
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

@media screen and (max-width:769px){
  .fs3_back{

  }
}
@media screen and (max-width:600px){

}
</style>
