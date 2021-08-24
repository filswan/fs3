<template>
  <div class="landing">

      <div class="table">
        <el-table :data="exChangeList" stripe style="width: 100%" class="demo-table-expand">
            <el-table-column prop="data.timeStamp" label="Date">
              <template slot-scope="scope">
                {{exChangeList[scope.$index].data.timeStamp}}
                <!-- {{ props.row.date }} -->
              </template>
            </el-table-column>
            <el-table-column prop="data.minerId" label="W3SS ID"></el-table-column>
            <el-table-column prop="data.price" label="Price">
              <template slot-scope="scope">
                {{exChangeList[scope.$index].data.price}} FIL
              </template>
            </el-table-column>
            <el-table-column prop="data.dealCid" label="Deal CID">
              <template slot-scope="scope">
                <div class="hot-cold-box">
                    <el-popover
                        placement="top"
                        trigger="hover"
                        v-model="exChangeList[scope.$index].data.visible">
                        <div class="upload_form_right">
                            <p>{{exChangeList[scope.$index].data.dealCid}}</p>
                        </div>
                        <el-button slot="reference" @click="copyLink(exChangeList[scope.$index].data.dealCid)">
                            <p><i class="el-icon-document-copy"></i>{{exChangeList[scope.$index].data.dealCid}}</p>
                        </el-button>
                    </el-popover>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="data.dataCid" label="Data CID">
              <template slot-scope="scope">
                <div class="hot-cold-box">
                    <el-popover
                        placement="top"
                        trigger="hover"
                        v-model="exChangeList[scope.$index].data.visibleDataCid">
                        <div class="upload_form_right">
                            <p>{{exChangeList[scope.$index].data.dataCid}}</p>
                        </div>
                        <el-button slot="reference" @click="copyLink(exChangeList[scope.$index].data.dataCid)">
                            <p><i class="el-icon-document-copy"></i>{{exChangeList[scope.$index].data.dataCid}}</p>
                        </el-button>
                    </el-popover>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="data.duration" label="Duration"></el-table-column>
        </el-table>
      </div>


  </div>
</template>

<script>
import axios from 'axios'
import Moment from 'moment'
let that
export default {
  name: 'all_deals',
  data() {
    return {
      postUrl: this.data_api + `/minio/webrpc`,
      logo: require("@/assets/images/title.svg"),
      bodyWidth: document.body.clientWidth>600?true:false,
      tableData: [],
      direction: 'ttb',
      drawIndex: 0,
      exChangeList: []
    }
  },
  components: {},
  methods: {
    tableJson(name) {
        let _this = this
        _this.exChangeList = []

        let postUrl = _this.data_api + `/minio/retrieve/` + _this.currentBucket + `/` + name
        axios.get(postUrl, {
           headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
           }
        }).then((response) => {
            let json = response.data
            if(json.status == 'success'){
              let dataAll = json.data
              if(dataAll.deals && dataAll.deals.length>0){
                dataAll.deals.map(item => {
                  if(item.data){
                    item.data.visible = false
                    item.data.visibleDataCid = false
                    item.data.timeStamp = Moment(new Date(item.data.timeStamp/1000)).format('YYYY-MM-DD HH:mm:ss')
                  }
                })
              }
              _this.exChangeList = dataAll.deals
            }else{
                _this.$message.error(json.message);
                return false
            }
        }).catch(function (error) {
            console.log(error);
            // console.log(error.message, error.request, error.response.headers);
        });
    },
  },
  watch: {},
  filters: {
    slideName: function (name) {
       if (!name) return '-';
       let retName = that.prefixName ? name.replace(that.prefixName+'/', "") : name
       return retName;
     }
  },
  mounted() {
    let _this = this
    that = _this
    document.onkeydown = function(e) {
      if (e.keyCode === 13) {
      }
    }
  }
}
</script>

<style lang="scss" scoped>
.landing{
  padding: 0 0 0.4rem;
  .table{
    width: 100%;
    margin: 0 0 .2rem;
    padding: 0 0 .2rem;
    // overflow-x: scroll;
    .el-table /deep/{
      // min-width: 500px;
      .el-table__header-wrapper{
        margin-bottom: 0.2rem;
      }
      th{
        >.cell{
          font-weight: 500;
          color: #818181;
          font-size: 0.15rem;
        }
      }
      th,td{
        &:nth-child(1){
          padding-left: 0.3rem;
        }
      }
      .descending {
        .sort-caret.descending{

        }
      }
      .cell{
        cursor: default;
        line-height: 0.35rem;
        .point{
          display: block;
          margin: auto;
          text-align: center;
          font-size: 0.18rem;
          color: #818181;
          cursor: pointer;
          &:hover{
            color: #333;
          }
        }
        .dropdown-menu {
          position: absolute;
          top: 100%;
          left: 0;
          z-index: 1000;
          display: none;
          float: left;
          min-width: 160px;
          padding: 5px 0;
          margin: 2px 0 0;
          list-style: none;
          font-size: 15px;
          text-align: left;
          background-color: #fff;
          border: 1px solid transparent;
          border-radius: 4px;
          box-shadow: 0 6px 12px rgba(0,0,0,0.18);
          background-clip: padding-box;

          .fiad-action {
              height: 0.35rem;
              width: 0.35rem;
              background: #ffc107;
              display: inline-block;
              border-radius: 50%;
              text-align: center;
              line-height: 0.35rem;
              font-weight: 400;
              position: relative;
              top: 12px;
              margin-left: 5px;
              animation-name: fiad-action-anim;
              transform-origin: center center;
              -webkit-backface-visibility: none;
              backface-visibility: none;
              box-shadow: 0 2px 4px rgba(0,0,0,0.1);
              display: flex;
              float: right;
              align-items: center;
              justify-content: center;
              i {
                  font-size: 0.18rem;
                  font-weight: bold;
                  color: #fff;
              }
              img {
                  display: block;
                  width: 100%;
                  height: 100%;
              }
          }
        }
        .dropdown-show{
          display: block;
          background-color: transparent;
          box-shadow: none;
          padding: 0;
          right: 0.6rem;
          top: 0;
          left: auto;
          margin: 0;
          height: 100%;
          text-align: right;
        }
        .iconBefore{
          line-height: 0.35rem;
          i{
            float: left;
            display: flex;
            justify-content: center;
            align-items: center;
            width: 0.35rem;
            height: 0.35rem;
            margin: 0 0.1rem 0 0;
            background-color: #32393f;
            font-size: 0.18rem;
            border-radius: 50%;
            color: #fff;
            cursor: pointer;
            font-weight: bold;
          }
          .iconfont{
            background-color: #afafaf;
            font-weight: normal;
            &:hover::before{
              content: "\e600";
              font-size: 0.16rem;
            }
          }
        }
      }

      .el-table__expanded-cell{
        padding: 0 !important;
      }
      .demo-table-expand {
        .el-table__header-wrapper{
          margin-bottom: 0;
        }
        th, td{
          &:first-child{
            padding-left: 0;
          }
        }
        .cell{
          cursor: default;
          text-align: center;
          word-break: break-word;
          line-height: 0.25rem;
          .hot-cold-box{
              .el-button{
                  width: 100%;
                  border: 0;
                  padding: 0;
                  background-color: transparent;
                  word-break: break-word;
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
                      font-weight: normal;
                      word-break: break-all;
                  }
                  i, img{
                      display: none;
                      float: left;
                      margin: 0 0.03rem;
                      font-size: 0.17rem;
                      line-height: 0.25rem;
                  }
              }
              .el-button:hover{
                  color: inherit;
                  i, img{
                      display: inline-block;
                  }
              }
          }

        }
      }
    }
    // &::-webkit-scrollbar{
    //     width: 7px;
    //     height: 7px;
    //     background-color: #F5F5F5;
    // }

    // /*定义滚动条轨道 内阴影+圆角*/
    // &::-webkit-scrollbar-track {
    //     box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
    //     -webkit-box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
    //     border-radius: 10px;
    //     background-color: #F5F5F5;
    // }

    // /*定义滑块 内阴影+圆角*/
    // &::-webkit-scrollbar-thumb{
    //     border-radius: 10px;
    //     box-shadow: inset 0 0 6px rgba(0, 0, 0, .1);
    //     -webkit-box-shadow: inset 0 0 6px rgba(0, 0, 0, .1);
    //     background-color: #c8c8c8;
    // }
  }
  .el-drawer__wrapper.drawStyle01 /deep/{
    bottom: auto;
    height: auto;
    padding: 20px 20px 20px 25px;
    background-color: #2298F6;
    z-index: 20;
    box-shadow: 0 0 10px rgba(0,0,0,0.3);
    text-align: center;
    .el-drawer.ltr, .el-drawer.rtl, .el-drawer__container{
      height: 0.35rem;
      line-height: 0.35rem;
    }
    .el-drawer{
        position: relative;
        height: 100% !important;
        background: transparent;
        box-shadow: none;
    }
    .el-drawer__header{
      display: none;
    }
    .el-drawer__body{
      font-size: 0.16rem;
      .draw_cont{
        display: flex;
        justify-content: space-between;
        align-items: center;
        color: #fff;
        line-height: 0.35rem;
        .draw_left{
          display: flex;
          justify-content: space-between;
          align-items: center;
          i{
            margin-right: 10px;
            font-size: 0.23rem;
          }
        }
        .draw_right{
          .el-button{
            font-size: 0.16rem;
            i{
              font-weight: bold;
            }
          }
          .btn{
            padding: 0.08rem;
            background-color: transparent;
            border: 2px solid hsla(0,0%,100%,.9);
            color: #fff;
            border-radius: 2px;
            padding: 5px 10px;
            font-size: 0.13rem;
            transition: all;
            transition-duration: .3s;
            margin-left: 10px;
            &:hover{
              color: #2298F6;
              background-color: #fff;
            }
          }
          .close{
            padding: 0.06rem;
            font-weight: bold;
            border-radius: 50%;
          }
        }
      }
    }
  }
  .model{
    display: flex;
    justify-content: center;
    align-items: center;
    position: fixed;
    top: 0;
    right: 0;
    left: 0;
    bottom: 0;
    z-index: 9;
    .model_bg{
      position: absolute;
      left: 0;
      top: 0;
      width: 100%;
      height: 100%;
      background-color: rgba(0,0,0,0.1);
      z-index: 10;
    }
    .model_cont{
      display: flex;
      position: relative;
      width: 600px;
      min-height: 300px;
      background-color: #00303f;
      z-index: 11;
      .model_close{
        right: 0.15rem;
        font-weight: 400;
        opacity: 1;
        font-size: 0.17rem;
        position: absolute;
        text-align: center;
        top: 0.15rem;
        z-index: 1;
        padding: 0;
        border: 0;
        background-color: hsla(0,0%,100%,.1);
        color: hsla(0,0%,100%,.8);
        width: 0.25rem;
        height: 0.25rem;
        display: block;
        border-radius: 50%;
        line-height: 0.25rem;
        text-shadow: none;
        cursor: pointer;
        &:hover{
          background-color: hsla(0,0%,100%,.2);
        }
      }
      .model_left{
        display: flex;
        justify-content: center;
        align-items: center;
        background-color: #022631;
        width: 150px;
        img{
          width: 70px;
        }
      }
      .model_right{
        display: flex;
        justify-content: center;
        align-items: center;
        width: calc(100% - 150px - 0.6rem);
        padding: 0.3rem;
        .el-row /deep/{
          width: 100%;
          .el-col{
            width: 100%;
            margin-bottom: 0.15rem;
            line-height: 1.42857143;
            h2{
              color: hsla(0,0%,100%,.8);
              text-transform: uppercase;
              font-size: 0.14rem;
              font-weight: normal;
              line-height: 2;
            }
            p{
              font-size: 0.13rem;
              color: hsla(0,0%,100%,.4);
            }
          }
        }
      }
    }
  }
  .el-dialog__wrapper /deep/{
    justify-content: center;
    display: flex;
    align-items: center;

    .el-dialog.customStyle{
        width: 400px;
        margin: 0 !important;
        position: absolute;
        bottom: 0.9rem;
        right: 0.2rem;
        .el-dialog__body{
            padding: 0.2rem 0.3rem 0.3rem;
            .el-input{
                .el-input__inner{
                    border: 0;
                    border-bottom: 1px solid #DCDFE6;
                    border-radius: 0;
                    text-align: center;
                    font-size: 0.13rem;
                    color: #32393f;
                }
            }
        }
    }
    .deleteStyle{
        width: 90%;
        max-width: 400px;
      .el-dialog__header{
          display: flex;
          .el-dialog__title{
              font-size: 0.15rem;
              color: #333;
          }
      }
      .el-dialog__body{
        padding: 0 0.2rem 0.2rem;
        img{
          display: block;
          width: 0.7rem;
          margin: 0 auto 0.05rem;
        }
        p{
          text-align: center;
          font-size: 0.15rem;
          line-height: 1.5;
          color: #333;
        }
        h6{
          font-size: 0.13rem;
          font-weight: normal;
          color: #bdbdbd;
          margin-top: 5px;
          text-align: center;
        }
        .btncompose{
              display: flex;
              align-items: center;
              justify-content: center;
              margin: 0.25rem auto 0.2rem;
              .el-button{
                padding: 0.05rem 0.1rem;
                margin: 0 0.03rem;
                font-size: 12px;
                color: #fff;
                border: 0;
                background-color: #ff726f;
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
  }
}
@media screen and (max-width:999px){
.landing{
  .fe-header{
    .feh-actions{
        top: 0.1rem;
        right: 0;
        position: fixed;
        .btn-group>button, >a{
          color: #fff;
        }
        .pcIcon{
          display: none !important;
        }
        .mobileIcon{
          display: block !important;
          i{
            font-size: 0.16rem !important;
          }
        }
    }
  }
}
}

@media screen and (max-width:600px){
.landing{
  .model{
    .model_cont{
      width: 90%;
      .model_left{
        display: none;
      }
      .model_right{
        width: 92%;
        padding: 4%;
      }
    }
  }
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
  }
  .table{
    overflow-x: auto;
    .el-table /deep/{
      min-width: 800px !important;
    }
  }
}
}
</style>

