<template>
      <el-dialog title="Retrieval" :visible.sync="retrievalDialog" :custom-class="{'ShareObjectMobile': shareFileShowMobile, 'ShareObject': 1 === 1}" :width='width' top="50px" :before-close="getDiglogChange">
          <div class="shareContent">
              <div class="tableStyle">
                <el-table :data="exChangeList" stripe style="width: 100%" class="demo-table-expand">
                    <el-table-column prop="data.timeStamp" label="Date">
                      <template slot-scope="scope">
                        {{exChangeList[scope.$index].data.timeStamp}}
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
            exChangeList: [],
            width: document.documentElement.clientWidth < 1024 ? '90%' : '75%'
        }
    },
    props: ['retrievalDialog', 'currentBucket'],
    watch: {
      'retrievalDialog': function(){
        let _this = this
        if(_this.retrievalDialog){
          _this.getData()
        }
      }
    },
    methods: {
      getDiglogChange() {
        this.$emit('getretrievalDialog', false)
      },
      getData(){
        let _this = this
        _this.exChangeList = []

        let postGetUrl = _this.data_api + `/minio/bucket/retrieve/` + _this.currentBucket
        axios.get(postGetUrl, {
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
            display: flex;
            margin: 0 0 0.3rem;
            .el-dialog__title{
                color: #333;
            }
        }
        .el-dialog__body{
          padding: 0;
          .shareContent{
            padding: 0 0.2rem 0.5rem;

            h4{
              width: 100%;
              font-weight: normal;
              display: block;
              margin: 20px 0 10px;;
              line-height: 2;
              font-size: .17rem;
              color: #333;
            }
            .tableStyle{
              width: 100%;
              .demo-table-expand {
                margin: 0.2rem auto;
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
