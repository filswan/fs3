<template>
    <div class="wrapper" @click="wrapperClick">
        <v-slide :class="{'sliMobile': slideShow}"
            :minioListBuckets="minioListBuckets" :currentBucket="currentBucket"
            :homeClick="homeClick" @homeClickFun="homeClickFun" @getshareHome="getshareHome" @getretrievalHome="getretrievalHome"
            @getminioListBucket="getminioListBucket" @getListBuckets="getListBuckets" @getMenuStretch="getMenuStretch"></v-slide>
        <div class="content" :class="{'content_stretch': menuStretch}">
            <div class="content_body">
                <el-row class="headStyle">
                    <el-col :span="6">
                        <el-button class="iconfont icon-ziyuan" @click.stop="slideBtn" v-if="!slideShow"></el-button>
                        <el-button class="el-icon-back" style="background-color: #484b4e;" @click.stop="slideBtn" v-else></el-button>
                    </el-col>
                    <el-col :span="12">
                        <img :src="logo" />
                    </el-col>
                    <el-col :span="6"></el-col>
                </el-row>
                <transition name="move" mode="out-in">
                    <router-view
                    :aboutServer="aboutServer" :aboutListObjects="aboutListObjects"
                    :slideListClick="slideListClick" :addFileClick="addFileClick" :uploadClick="uploadClick"
                    :dialogFormVisible="dialogFormVisible" :currentBucket="currentBucket" :userd="userd" :allDealShow="allDealShow"
                    @getDialogClose="getDialogClose"
                    @getaboutServer="getMakeBucket"
                    @getRemoveObject="getRemoveObject"
                    @getListObjects="getListObjects"></router-view>
                </transition>
            </div>
            <div class="fes-icon">
                <div class="fes-icon-logo">
                    <a href="https://filswan.medium.com/" target="_block"><img :src="share_img1" alt=""></a>
                    <a href="https://discord.com/invite/KKGhy8ZqzK" target="_block"><img :src="share_img10" alt=""></a>
                    <a href="https://twitter.com/0xfilswan" target="_block"><img :src="share_img2" alt=""></a>
                    <a href="https://github.com/filswan" target="_block"><img :src="share_img3" alt=""></a>
                    <!-- <a href="https://www.facebook.com/filswan.technology" target="_block"><img :src="share_img5" alt=""></a>
                    <a href="https://filswan.slack.com" target="_block"><img :src="share_img7" alt=""></a>
                    <a href="https://youtube.com/channel/UCcvrZdNqFWYl3FwfcHS9xIg" target="_block"><img :src="share_img8" alt=""></a> -->
                    <a href="https://t.me/filswan" target="_block"><img :src="share_img9" alt=""></a>
                </div>
                <div class="fes-icon-copy">
                    <span>© 2022 FilSwan Canada</span>
                    <el-divider direction="vertical"></el-divider>
                    <a href="https://www.filswan.com/" target="_block">filswan.com</a>

                </div>
            </div>
            <div class="addFile">
                <el-row v-if="addFileShow">
                    <el-col :span="24">
                        <el-upload
                            class="upload-demo"
                            action="customize"
                            ref="uploadFile"
                            :http-request="httpRequest"
                            :on-change="onChange"
                            multiple
                            :auto-upload="false"
                            >
                            <el-tooltip class="item" effect="dark" content="Upload file" placement="left">
                                <i class="iconfont icon-shangchuan"></i>
                            </el-tooltip>
                        </el-upload>
                    </el-col>
                    <el-col :span="24">
                        <el-tooltip class="item" effect="dark" content="Create bucket" placement="left" @click.native="createHomeBuck">
                            <i class="iconfont icon-harddriveyingpan"></i>
                        </el-tooltip>
                    </el-col>
                </el-row>
                <i class="el-icon-plus" :class="{'el-icon-plus-new': addFileShow}" @click.stop="addToggle"></i>
            </div>

            <div class="progressStyle" id="progressStyle"></div>
            <el-backtop target=".wrapper"></el-backtop>
        </div>

        <share-dialog  v-if="isRouterAlive"
          :shareDialog="shareDialog" :shareObjectShow="shareObjectShow"
          :shareFileShow="shareFileShow" :postAdress="currentBucket" :sendApi="sendApi"
          @getshareDialog="getshareDialog">
        </share-dialog>

        <retrieval-dialog
          :retrievalDialog="retrievalDialog" :currentBucket="currentBucket"
          @getretrievalDialog="getretrievalDialog">
        </retrieval-dialog>
    </div>
</template>

<script>
import axios from 'axios'
import vSlide from './Slide.vue';
import Moment from "moment"
import shareDialog from '@/components/shareDialog.vue';
import retrievalDialog from '@/components/retrievalDialog.vue';
export default {
    provide () {
        return {
            reload: this.reload
        }
    },
    data() {
        return {
            postUrl: this.data_api + `/minio/webrpc`,
            logo: require("@/assets/images/logo.png"),
            share_img1: require('@/assets/images/landing/medium.png'),
            share_img2: require('@/assets/images/landing/twitter.png'),
            share_img3: require('@/assets/images/landing/github-fill.png'),
            share_img5: require('@/assets/images/landing/facebook-fill.png'),
            share_img7: require('@/assets/images/landing/slack.png'),
            share_img8: require('@/assets/images/landing/youtube.png'),
            share_img9: require('@/assets/images/landing/telegram.png'),
            share_img10: require('@/assets/images/landing/discord.png'),
            share_logo: require('@/assets/images/landing/logo_small.png'),
            bodyWidth: document.body.clientWidth<=1024?true:false,
            addFileShow: false,
            dialogFormVisible: false,
            form: {
                name: ''
            },
            slideShow: false,
            minioListBuckets: {
                buckets: [],
                uiVersion: ""
            },
            currentBucket: '',
            minioStorageInfo: {
                storageInfo: {},
                uiVersion: ""
            },
            userd: 0,
            aboutServer: {
                MinioVersion:"",
                MinioMemory:"",
                MinioPlatform:"",
                MinioRuntime:"",
                MinioGlobalInfo:{}
            },
            aboutListObjects: {
                objects: [],
                uiVersion: "",
                writable: true
            },
            fileList: [],
            fileListIndexNow: 0,
            actionUrl: '',
            prefixData: '',
            homeClick: false,
            addArr: [],
            progressArr: {
              ot: 0,
              oloaded: 0,
              percentage_new: 0,
            },
            percentage_new: 0,
            drawer: false,
            customColor: '#5cb87a',
            shareDialog: false,
            shareObjectShow: true,
            shareFileShow: false,
            slideListClick: 0,
            addFileClick: 0,
            uploadClick: 0,
            sendApi: 1,
            allDealShow: true,
            retrievalDialog: false,
            isRouterAlive: true,
            menuStretch: false
        }
    },
    components: {
        vSlide,
        shareDialog,
        retrievalDialog
    },
    computed: {
        headertitle() {
            return this.$store.getters.headertitle
        },
        routerMenu() {
            return this.$store.getters.routerMenu
        },
    },
    methods: {
        getMenuStretch(stretch) {
            this.menuStretch = stretch
        },
        reload () {
            this.isRouterAlive = false;
            this.$nextTick(function () {
                this.isRouterAlive = true;
            })
        },
        getshareDialog(shareDialog) {
          this.shareDialog = shareDialog
        },
        getshareHome(shareDialog, shareObjectShow, shareFileShow){
          this.shareDialog = shareDialog
          this.shareObjectShow = shareObjectShow
          this.shareFileShow = shareFileShow
        },
        getretrievalHome(retrievalDialog) {
          this.retrievalDialog = retrievalDialog
        },
        getretrievalDialog(retrievalDialog) {
          this.retrievalDialog = retrievalDialog
        },
        getData() {
            this.getListBuckets()
            this.getStorageInfo()
            this.getServerInfo()
        },
        getListBuckets(name) {
            let _this = this
            let dataListBuckets = {
                id: 1,
                jsonrpc: "2.0",
                method: "web.ListBuckets",
                params:{}
            }

            axios.post(_this.postUrl, dataListBuckets, {headers: {
                'Authorization': "Bearer "+ _this.$store.getters.accessToken
            }}).then((response) => {
                let json = response.data
                let error = json.error
                let result = json.result
                if (error) {
                    _this.$message.error(error.message);
                    return false
                }
                _this.minioListBuckets = result
                _this.currentBucket = _this.minioListBuckets && _this.minioListBuckets.buckets?_this.minioListBuckets.buckets[0].name:''

                if(name) {
                  _this.getListObjects(name)
                  return false
                }
                if(_this.minioListBuckets.buckets){
                  _this.getListObjects()
                }

            }).catch(function (error) {
                // console.log(error.request.status);
                if(error.request.status == '401'){
                  _this.$store.dispatch("FedLogOut").then(() => {
                    _this.$router.push("/fs3/login")
                  })
                }
            });
        },
        getStorageInfo() {
            let _this = this;
            let dataStorageInfo = {
                id: 1,
                jsonrpc: "2.0",
                method: "web.StorageInfo",
                params:{}
            }
            axios.post(_this.postUrl, dataStorageInfo, {headers: {
                'Authorization': "Bearer "+ _this.$store.getters.accessToken
            }}).then((response) => {
                let json = response.data
                let error = json.error
                let result = json.result
                if (error) {
                    _this.$message.error(error.message);
                    return false
                }
                _this.minioStorageInfo = result
                _this.userd = result.used

            }).catch(function (error) {
                console.log(error);
            });
        },
        getServerInfo() {
            let _this = this
            let dataServerInfo = {
                id: 1,
                jsonrpc: "2.0",
                method: "web.ServerInfo",
                params:{}
            }
            axios.post(_this.postUrl, dataServerInfo, {headers: {
                'Authorization': "Bearer "+ _this.$store.getters.accessToken
            }}).then((response) => {
                let json = response.data
                let error = json.error
                let result = json.result
                if (error) {
                    _this.$message.error(error.message);
                    return false
                }
                _this.aboutServer = result

            }).catch(function (error) {
                console.log(error);
            });
        },
        getListObjects(listName, prefixData) {
            let _this = this
            _this.prefixData = prefixData
            let dataListObjects = {
                id: 1,
                jsonrpc: "2.0",
                method: "web.ListObjects",
                params:{
                    bucketName: listName?listName:_this.minioListBuckets.buckets?_this.minioListBuckets.buckets[0].name:'',
                    prefix: _this.prefixData?_this.prefixData + '/':""
                }
            }
            _this.currentBucket = listName?listName:_this.minioListBuckets.buckets?_this.minioListBuckets.buckets[0].name:''
            axios.post(_this.postUrl, dataListObjects, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
            }}).then((response) => {
                let json = response.data
                let error = json.error
                let result = json.result
                if (error) {
                    _this.$message.error(error.message);
                    return false
                }
                _this.aboutListObjects = result

            }).catch(function (error) {
                console.log(error);
            });
        },
        getDialogClose(dialogFormVisible, closeModule) {
            this.dialogFormVisible = dialogFormVisible
            if(!closeModule){
              this.addFileShow = false
              this.homeClick = false
            }
        },
        getMakeBucket(name, dialogFormVisible, prefix, oldName) {
            let _this = this
            let dataMakeBucket = {
                id: 1,
                jsonrpc: "2.0",
                method: "web.MakeBucket",
                params:{
                    bucketName: name
                }
            }
            _this.dialogFormVisible = dialogFormVisible
            axios.post(_this.postUrl, dataMakeBucket, {headers: {
                'Authorization':"Bearer "+ _this.$store.getters.accessToken
            }}).then((response) => {
                let json = response.data
                let error = json.error
                let result = json.result
                if (error) {
                    _this.$message.error(error.message);
                    if(oldName) {
                      _this.currentBucket = oldName
                    }
                    return false
                }
                _this.currentBucket = name
                if(_this.minioListBuckets && _this.minioListBuckets.buckets) {
                  _this.minioListBuckets.buckets.map(item => {
                    if(item.name.indexOf(name) >= 0){
                      _this.getListBuckets()
                      return false
                    }
                  })
                }
                _this.getListBuckets(name)
                _this.getListObjects(name)

                if(prefix){
                    _this.getListObjects(name, false, prefix)
                }

            }).catch(function (error) {
                console.log(error);
            });

        },
        getRemoveObject(data){
            let _this = this
            _this.aboutListObjects.objects = JSON.parse(JSON.stringify(data))
        },
        getminioListBucket(listName, all, silde, push) {
          if(push) this.$router.push({name: 'fs3'})
          if(listName){
            this.$router.push({name: 'fs3'})
            this.getListObjects(listName)
            this.slideListClick += 1
          }
          this.allDealShow = all
          if(silde) this.slideShow=false
        },
        addToggle() {
           this.addFileShow = !this.addFileShow
           this.homeClick = false
           this.addFileClick += 1
        },
        slideBtn() {
            this.slideShow = !this.slideShow
        },
        wrapperClick() {
            this.addFileShow = false
            this.homeClick = false
            this.slideShow = false
        },
        homeClickFun(now) {
            this.homeClick = now
            this.addFileShow = false
        },
        createHomeBuck(){
            let _this = this
            let path = _this.$route.path
            if(path.indexOf('/fs3') < 0){
                this.$router.push({name: 'fs3'})
                this.allDealShow = true
            }
            this.dialogFormVisible = true
        },

    //File upload
    httpRequest(file) {
      // console.log('httpRequest', file);
    },
    onChange(file, fileList) {
      let _this = this
    //   console.log('onChange', file, fileList, fileList.indexOf(fileList.filter(d=>d.name == file.name)[0]));
      _this.fileListIndexNow += 1
      let indexNow = _this.fileListIndexNow
      let progressArr = {
            ot: 0,
            oloaded: 0,
            percentage_new: 0
        }
      let regexp = /[#\\?]/
      if(regexp.test(file.name)){
        _this.$message.error('The filename cannot contain any of the following characters # ? \\');
        return false
      }

      let reg=new RegExp(" ","g");
      if(file.name.indexOf(" ") > -1){
        file.name=file.name.replace(reg,"_");
        file.raw = new File([file.raw], file.name)
      }

        let $hgh
        if(!_this.minioListBuckets.buckets || _this.minioListBuckets.buckets.length < 1){
            _this.$message({
                message: 'Please choose a bucket before trying to upload files.',
                type: 'error',
                showClose: true,
                duration: 0
            });
            $hgh = true
            return false
        }

        if(!$hgh) {
          let prefix = _this.prefixData ? _this.prefixData + '/': ''
          let postUrl = _this.data_api + '/minio/upload/' + _this.currentBucket + '/' + prefix + file.name
          let formData = new FormData();  //Create Empty

            var evs = {};
            evs.indexNow = document.createElement('div');
            var div1 = document.createElement('div');
            var progress = document.createElement('progress');
            var span = document.createElement('span');
            var span1 = document.createElement('span1');
            evs.indexNow.id= "div" + indexNow
            evs.indexNow.className= "div"
            evs.indexNow.style.width = '100%'
            evs.indexNow.style.margin = '0.1rem 0.2rem'

            div1.id = 'speed' + indexNow
            div1.className = 'speed'

            progress.id = 'progressBar0' + indexNow
            progress.value = 0
            progress.max = 100
            progress.style.width = '100%'

            span.id= "time" + indexNow
            span1.id= "percentage" + indexNow

            document.getElementById('progressStyle').appendChild(evs.indexNow);
            document.getElementById('div' + indexNow).appendChild(progress);
            document.getElementById('div' + indexNow).appendChild(div1);
            document.getElementById('speed' + indexNow).appendChild(span);
            document.getElementById('speed' + indexNow).appendChild(span1);

            //   document.getElementById("progressBar01").value = 0
              let xhr
              xhr = new XMLHttpRequest()
              xhr.open("PUT", postUrl, true)
              xhr.withCredentials = false
              const token = _this.$store.getters.accessToken
              if (token) {
                xhr.setRequestHeader(
                  "Authorization",
                  "Bearer " + _this.$store.getters.accessToken
                )
              }
              xhr.setRequestHeader(
                "x-amz-date",
                Moment()
                  .utc()
                  .format("YYYYMMDDTHHmmss") + "Z"
              )


              xhr.onload = function(event) {
                if (xhr.status == 401 || xhr.status == 403) {
                  _this.$message({
                      message: "Unauthorized request.",
                      type: 'danger'
                  });
                }
                if (xhr.status == 500) {
                  _this.$message({
                      message: xhr.responseText,
                      type: 'danger'
                  });
                }
                if (xhr.status == 200) {
                    _this.$message({
                        message: "File '" + file.name + "' uploaded successfully.",
                        type: 'success'
                    });

                    _this.getListObjects(_this.currentBucket, _this.prefixData)
                    _this.uploadClick += 1
                }

                xhr.upload.addEventListener("error", event => {
                    _this.$message({
                        message: "Error occurred uploading '" + file.name + "'.",
                        type: 'danger'
                    });
                })

                xhr.upload.addEventListener("progress", event => {
                  if (event.lengthComputable) {
                    let loaded = event.loaded
                    let total = event.total
                    // Update the counter
                    //dispatch(updateProgress(slug, loaded))
                  }
                })

                //xhr.send(file.raw)
             }

            //  xhr.upload.onprogress = _this.progressFunction;//Implementation of upload progress call method
             xhr.upload.onprogress = function(evt){
                    let progressBar = document.getElementById("progressBar0"+indexNow);
                    let percentageDiv = document.getElementById("percentage"+indexNow);
                    if (evt.lengthComputable) {//
                        progressBar.max = evt.total;
                        progressBar.value = evt.loaded;
                        progressArr.percentage_new = Math.round(evt.loaded / evt.total * 100);
                        percentageDiv.innerHTML = "(" + Math.round(evt.loaded / evt.total * 100) + "%)";
                    }

                    let time = document.getElementById("time"+indexNow);
                    let nt = new Date().getTime();//Get current time
                    var pertime = (nt - progressArr.ot)/1000; //Calculate the time difference from the last time this method was called to the present, unit: s
                    progressArr.ot = new Date().getTime(); //Reassign time for next calculation

                    var perload = evt.loaded - progressArr.oloaded; //Calculate the file size uploaded by this segment, unit B
                    progressArr.oloaded = evt.loaded;//Reassign the uploaded file size, calculated with the following times

                    //Upload speed calculation
                    var speed = perload/pertime;//unit b/s
                    var bspeed = speed;
                    var units = 'b/s'; //unit
                    if(speed/1024>1){
                        speed = speed/1024;
                        units = 'k/s';
                    }
                    if(speed/1024>1){
                        speed = speed/1024;
                        units = 'M/s';
                    }
                    speed = speed.toFixed(1);
                    //Remaining time
                    var resttime = ((evt.total-evt.loaded)/bspeed).toFixed(1);
                    time.innerHTML = speed+units;
                    if(bspeed==0)
                        time.innerHTML = 'Upload cancelled';
                    if(!resttime || resttime <= 0){
                        //Notification.closeAll()
                        document.getElementById('progressStyle').removeChild(evs.indexNow)
                        return true
                    }
             };//Implementation of upload progress call method
             xhr.upload.onloadstart = function(){//Upload start execution method
                 progressArr.ot = new Date().getTime();   //Set upload start time
                 progressArr.oloaded = 0;//Set the file size to 0 when uploading starts
                 progressArr.percentage_new = 0
                 _this.drawer = true
             };
             xhr.send(file.raw)
        }
      },
      //Upload progress implementation method, which will be called frequently during the upload process
      progressFunction(evt) {
           let _this = this
           let progressBar = document.getElementById("progressBar01");
           let percentageDiv = document.getElementById("percentage");
           if (evt.lengthComputable) {//
               progressBar.max = evt.total;
               progressBar.value = evt.loaded;
               _this.percentage_new = Math.round(evt.loaded / evt.total * 100);
               percentageDiv.innerHTML = "(" + Math.round(evt.loaded / evt.total * 100) + "%)";
           }

          let time = document.getElementById("time");
          let nt = new Date().getTime();//Get current time
          var pertime = (nt - _this.progressArr.ot)/1000; //Calculate the time difference from the last time this method was called to the present, unit: s
          _this.progressArr.ot = new Date().getTime(); //Reassign time for next calculation

          var perload = evt.loaded - _this.progressArr.oloaded; //Calculate the file size uploaded by this segment, unit B
          _this.progressArr.oloaded = evt.loaded;//Reassign the uploaded file size, calculated with the following times

          //Upload speed calculation
          var speed = perload/pertime;//unit b/s
          var bspeed = speed;
          var units = 'b/s'; //unit
          if(speed/1024>1){
              speed = speed/1024;
              units = 'k/s';
          }
          if(speed/1024>1){
              speed = speed/1024;
              units = 'M/s';
          }
          speed = speed.toFixed(1);
          //Remaining time
          var resttime = ((evt.total-evt.loaded)/bspeed).toFixed(1);
          time.innerHTML = speed+units;
          if(bspeed==0)
              time.innerHTML = 'Upload cancelled';
          if(!resttime || resttime <= 0){
            //Notification.closeAll()
            _this.drawer = false
          }
      }


    },
    mounted() {
        let _this = this
        _this.getData()
        localStorage.removeItem('addrWeb')
        console.log('update time: 2022-03-25')
    },
};
</script>

<style lang="scss" scoped>
.wrapper{
    display: flex;
    flex-wrap: wrap;
    .content{
        position: relative;
        width: calc(100% - 3.2rem);
        height: 100%;
        overflow-y: scroll;
        transition: all;
        transition-duration: .3s;
        .headStyle{
            display: none;
        }
        .el-backtop{
            background-color: #45a2ff;
        }
        .el-backtop, .el-calendar-table td.is-today{
            color: #fff;
        }
        .content_body{
            min-height: calc(100% - 65px);
        }
        .fes-icon{
            background-color: #fff;
            z-index: 8;
            .fes-icon-logo{
                display: flex;
                justify-content: center;
                align-items: center;
                padding: 10px 0;
                img{
                    display: block;
                    height: 20px;
                    margin: 0 0.05rem;
                    @media screen and (max-width: 999px) {
                        height: 20px;
                    }
                }
            }
            .fes-icon-copy{
                display: flex;
                justify-content: center;
                align-items: center;
                padding: 0 0 10px;
                span, a{
                    font-size: 12px;
                    color: #333;
                    line-height: 15px;
                }
                a{
                    &:hover{
                        color: #409eff;
                    }
                }
                .el-divider--vertical /deep/{
                    height: 15px;
                }
            }
        }
        .addFile{
            display: flex;
            flex-wrap: wrap;
            position: fixed;
            right: 0.3rem;
            bottom: 0.2rem;
            width: 0.55rem;
            z-index: 9;
            @media screen and (max-width:600px){
                width: 40px;
            }
            .el-icon-plus{
                width: 0.55rem;
                height: 0.55rem;
                line-height: 0.55rem;
                border-radius: 50%;
                background: #ff726f;
                box-shadow: 0 2px 3px rgba(0,0,0,0.15);
                display: inline-block;
                text-align: center;
                border: 0;
                padding: 0;
                color: #fff;
                font-size: 0.2rem;
                font-weight: bold;
                cursor: pointer;
                transition: all;
                transition-duration: .3s;
                @media screen and (max-width:600px){
                    width: 40px;
                    height: 40px;
                    display: flex;
                    justify-content: center;
                    align-items: center;
                    line-height: 40px;
                    font-size: 15px;
                }
            }
            .el-icon-plus-new{
                background-color: #ff403c;
                transform: rotate(45deg);
            }
            .el-row /deep/{
                width: 100%;
                .el-col{
                    width: 100%;
                    display: flex;
                    justify-content: center;
                    i{
                        width: 0.4rem;
                        margin: 0 auto 0.15rem;
                        height: 0.4rem;
                        background-color: #ffc107;
                        border-radius: 50%;
                        text-align: center;
                        display: inline-block;
                        line-height: 40px;
                        box-shadow: 0 2px 3px rgba(0,0,0,0.15);
                        transform: scale(0);
                        position: relative;
                        animation-name: feba-btn-anim;
                        animation-duration: .3s;
                        animation-fill-mode: forwards;
                        color: #fff;
                        cursor: pointer;
                        font-size: 0.18rem;
                        @media screen and (max-width:999px){
                            width: 30px;
                            height: 30px;
                            display: flex;
                            justify-content: center;
                            align-items: center;
                            line-height: 30px;
                            font-size: 15px;
                        }
                    }
                }
                @-webkit-keyframes feba-btn-anim {
                    from {
                        transform: scale(0);
                        opacity: 0;
                    }
                    to {
                        transform: scale(1);
                        opacity: 1;
                    }
                }

                @keyframes feba-btn-anim {
                    from {
                        transform: scale(0);
                        opacity: 0;
                    }
                    to {
                        transform: scale(1);
                        opacity: 1;
                    }
                }
            }
        }
        &::-webkit-scrollbar{
            width: 1px;
            height: 1px;
            background-color: #F5F5F5;
        }

        &::-webkit-scrollbar-track {
            box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
            -webkit-box-shadow: inset 0 0 6px rgba(0, 0, 0, 0.3);
            border-radius: 10px;
            background-color: #F5F5F5;
        }

        &::-webkit-scrollbar-thumb{
            border-radius: 10px;
            box-shadow: inset 0 0 6px rgba(0, 0, 0, .1);
            -webkit-box-shadow: inset 0 0 6px rgba(0, 0, 0, .1);
            background-color: #c8c8c8;
        }
        .progressStyle{
          position: fixed;
          z-index: 999999;
          bottom: 0px;
          right: 50px;
          background: #00b7ff none repeat scroll 0% 0%;
          font-size: 14px;
          color: #fff;
        //   padding: 0.2rem 0.4rem;
          width: 360px;
          display: flex;
          flex-wrap: wrap;
          div{
              margin: 0.1rem 0.2rem;
          }
          .el-progress /deep/{
            width: 100%;
            .el-progress-bar{
              width: 100%;
              .el-progress-bar__inner{
                display: none;
              }
            }
            .el-progress__text{
                display: none;
                opacity: 0;
            }
          }
          .speed{
            display: flex;
            justify-content: center;
            align-items: center;
            width: 100%;
            margin: 0;
          }
        }
    }
    .content_stretch{
        width: calc(100% - 0.65rem);
    }
}
@media screen and (max-width:999px){
.wrapper{
    .content{
        width: 100%;
        height: calc(100% - 65px);
        padding-top: 65px;
        .headStyle.el-row /deep/{
            display: block;
            background-color: #32393f;
            padding: 10px 12px 9px 12px;
            text-align: center;
            position: fixed;
            z-index: 9999;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.3);
            left: 0;
            top: 0;
            width: 100%;
            .el-col{
                display: flex;
                img{
                    display: block;
                    height: 35px;
                    margin: 5px auto 0;
                }
                .el-button{
                    display: block;
                    height: 45px;
                    min-width: 45px;
                    text-align: center;
                    border-radius: 50%;
                    padding: 0;
                    border: 0;
                    background: none;
                    color: #fff;
                    font-size: 21px;
                    font-family: inherit;
                    line-height: 45px;
                    -webkit-transition: all;
                    transition: all;
                    -webkit-transition-duration: .3s;
                    transition-duration: .3s;
                    cursor: pointer;
                }
            }
        }

    }
}
}
@media screen and (max-width:600px){
.wrapper{
    .el-dialog__wrapper /deep/{
        .el-dialog.customStyle{
            width: 300px;
        }
    }
}
}
</style>
