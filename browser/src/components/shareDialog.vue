<template>
  <div>
      <el-dialog title="" top="50px" :visible.sync="shareDialog" :custom-class="{'ShareObjectMobile': shareFileShowMobile, 'ShareObject': 1 === 1}" :before-close="getDiglogChange">
          <div class="shareContent" v-loading="loadShare">
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

              <!-- <el-tabs v-model="activeOn" @tab-click="handleClick" tab-position="left" v-if="shareFileShow && sendApi == 1">
                <el-tab-pane label="online" name="online"></el-tab-pane>
                <el-tab-pane label="offline" name="offline"></el-tab-pane>
              </el-tabs> -->

              <el-row class="share_right" v-if="shareFileShow && activeOn == 'online'">
                <el-button 
                  class="shareFileCoinSend" @click="submitForm('ruleForm')" 
                  :disabled="ruleForm.duration_tip" :style="{'opacity':ruleForm.duration_tip?'0.3':'1'}">Send</el-button>
                <el-col :span="24">
                  <!--h4 v-if="shareObjectShow">Share to Filecoin</h4-->
                  <h4>Backup to Filecoin</h4>
                </el-col>
                <el-col :span="24">
                  <el-form :model="ruleForm" :rules="rules" ref="ruleForm" label-width="110px" class="demo-ruleForm">
                    <el-form-item label="Provider ID:" prop="minerId" v-if="activeOn == 'online'">
                      <!--el-input v-model="ruleForm.minerId"></el-input-->
                      <el-menu :default-active="'1'" menu-trigger="click hover" class="el-menu-demo" mode="horizontal" @open="handleOpen" @close="handleClose" :unique-opened="true"  v-show="activeOn == 'online'" >
                          <el-submenu index="1" popper-class="myMenu" :popper-append-to-body="false">
                              <template slot="title">
                                 <!-- {{ name }} -->
                                  <el-input v-model="ruleForm.minerId" @blur="inputBlur(ruleForm.minerId, 2)" placeholder="Please select Provider ID"></el-input>
                                  <p class="el-form-item__error" v-if="ruleForm.minerId_tip">{{ruleForm.minerId_tip}}</p>
                              </template>
                              <el-submenu :index="'1-'+n" v-for="(item, n) in locationOptions" :key="n" :attr="'1-'+n">
                                  <template slot="title">
                                      <span class="span" @mouseover="handleOpenSubmenu(item.value)">{{ item.value }}</span>
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
                      <el-input v-model="ruleForm.price" onkeyup="value=value.replace(/^\D*(\d*(?:\.\d{0,18})?).*$/g, '$1')" @blur="inputBlur(ruleForm.price, 1)" @input="inputChange"></el-input> FIL
                      <p class="el-form-item__error" v-if="ruleForm.price_tip">The minimum price is 0.000000000000000001FIL</p>
                    </el-form-item>
                    <el-form-item prop="duration">
                      <template slot="label">
                          <span style="color: #F56C6C;margin-right: 4px;">*</span>Duration:
                      </template>
                      <el-input v-model="ruleForm.duration" onkeyup="value=value.replace(/^(0+)|[^\d]+/g,'')" @blur="calculation"></el-input> Day
                      <p class="_error" v-if="ruleForm.duration_tip">Duration must be in the range of 180-540 days</p>
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


              <!-- <el-row class="share_right" v-if="shareFileShow && activeOn == 'offline'">
                <el-button class="shareFileCoinSend" @click="submitofflineForm('offlineForm')">Send</el-button>
                <el-col :span="24">
                  <h4>Backup to Filecoin</h4>
                </el-col>
                <el-col :span="24">
                  <el-form :model="offlineForm" :rules="ruleOfflines" ref="offlineForm" label-width="115px" class="demo-ruleForm">
                    <el-form-item label="Task Name:" prop="task_name">
                      <el-input v-model="offlineForm.task_name"></el-input>
                    </el-form-item>
                    <el-form-item label="Description:" prop="desc">
                      <el-input type="textarea" v-model="offlineForm.desc"></el-input>
                    </el-form-item>
                    <el-form-item label="Tags:" prop="tags">
                        <el-tag :key="tag" v-for="tag in offlineForm.tags" closable :disable-transitions="false" @close="handleTagsClose(tag)">
                            {{tag}}
                        </el-tag>
                        <el-input class="input-new-tag" v-if="inputVisibleTask" v-model="inputValueTask" ref="saveTagInput" size="small"
                                  @blur="handleInputConfirmTask"  @keyup.enter.native="handleInputConfirmTask"
                                  maxlength="15"
                        >
                        </el-input>
                        <el-button v-else class="button-new-tag" size="small" @click="showInputTask">+ New Tag</el-button>

                    </el-form-item>
                    <el-form-item label="Curated Dataset:" prop="curated_dataset" class="lineHeight">
                      <el-input v-model="offlineForm.curated_dataset"></el-input>
                    </el-form-item>
                    <el-form-item label="Type:" prop="type">
                      <el-radio v-model="offlineForm.type" label="regular">Regular</el-radio>
                      <el-radio v-model="offlineForm.type" label="verified">Verified</el-radio>
                    </el-form-item>
                    <el-form-item label="Open Bid:" prop="OpenBidType">
                      <el-radio v-model="offlineForm.OpenBidType" :label="1">True</el-radio>
                      <el-radio v-model="offlineForm.OpenBidType" :label="0">False</el-radio>
                    </el-form-item>
                    <el-form-item prop="bidDay" v-if="offlineForm.OpenBidType == 1">
                        <h4 style="margin: 0px 0px 0px -75px;color: #606266;font-size: 14px;">Expect Complete in <el-input v-model="offlineForm.bidDay" style="width:60px"></el-input> days.</h4>
                    </el-form-item>
                    <el-form-item label="Estimated Budget" prop="bidprice" v-if="offlineForm.OpenBidType == 1" class="lineHeight">
                        <h4 style="margin: 0;">
                            <el-input v-model="offlineForm.min_price" style="width:100px" placeholder="Min"></el-input>
                            -
                            <el-input v-model="offlineForm.max_price" style="width:100px" placeholder="Max"></el-input>
                        </h4>
                    </el-form-item>
                    <el-form-item label="Provider ID" prop="providerId" v-show="offlineForm.OpenBidType != 1">
                      <el-menu :default-active="'1'" menu-trigger="click hover" class="el-menu-demo" mode="horizontal" @open="handleOpen" @close="handleClose" :unique-opened="true" v-if="offlineForm.OpenBidType != 1">
                          <el-submenu index="1" popper-class="myMenu" :popper-append-to-body="false">
                              <template slot="title">
                                  <el-input v-model="offlineForm.providerId" placeholder="Please select Provider ID"></el-input>
                              </template>
                              <el-submenu :index="'1-'+n" v-for="(item, n) in locationOptions" :key="n" :attr="'1-'+n">
                                  <template slot="title">
                                      <span class="span" @mouseover="handleOpenSubmenu(item.value)">{{ item.value }}</span>
                                  </template>
                                  <el-menu-item :index="'1-'+n+'-1'" :attr="'1-'+n+'-1'">
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
                      <h4 style="margin: 0;color: #f56c6c;font-size: 12px;line-height: 1;clear: both;" v-if="offlineForm.providerId_tips">Please enter Provider ID</h4>
                    </el-form-item>
                  </el-form>
                </el-col>
              </el-row> -->
          </div>
      </el-dialog>

      <el-dialog title="" :visible.sync="finishTransaction" :width="width" class="completed">
        <h1>Completed!</h1>
        <h3>Please check FilSwan platform, the task {{taskName}} has been successfully added to your FilSwan account.</h3>
      </el-dialog>
  </div>
</template>

<script>
let that
import axios from 'axios'
import QRCode from 'qrcodejs2'
export default {
    inject:['reload'],
    data() {
        var validateDuration = (rule, value, callback) => {
            value = value.replace(/[^\d.]/g,'')
            if (!value) {
                return callback(new Error('Please enter Duration'));
            }
            setTimeout(() => {
                callback()
            }, 100);
        };
        return {
            postUrl: this.data_api + `/minio/webrpc`,
            shareFileShowMobile: false,
            ruleForm: {
              minerId: '',
              minerId_tip: '',
              price: '',
              price_tip: false,
              duration: '',
              duration_tip: false,
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
                 { required: true, message: 'Please enter Provider ID', trigger: ['blur', 'change'] }
               ],
               price: [
                 { required: true, message: 'Please enter Price', trigger: ['blur', 'change'] }
               ],
               duration: [
                 { validator: validateDuration, trigger: 'blur' }
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
            loadShare: false,
            bodyWidth: document.documentElement.clientWidth < 1024 ? true : false,
            name: 'Please select Provider ID',
            nameOffline: 'Please select Provider ID',
            parentLi: '',
            parentName: '',
            radio: '1',
            offlineForm: {
              task_name: '',
              desc: '',
              tags: [],
              curated_dataset: '',
              type: 'regular',
              OpenBidType: 1,
              bidDay: '',
              min_price: '',
              max_price: '',
              providerId: '',
              providerId_tips: false
            },
            ruleOfflines: {
               task_name: [
                 { required: true, message: 'Please enter Task Name', trigger: 'blur' }
               ]
            },
            activeOn: 'online',
            inputVisibleTask: false,
            inputValueTask: '',
            width: document.body.clientWidth>600?'400px':'95%',
            finishTransaction: false,
            taskName: ''
        }
    },
    props: ['shareDialog','shareObjectShow','shareFileShow', 'num', 'share_input', 'postAdress', 'sendApi'],
    watch: {
      'shareDialog': function(){
        let _this = this
        if(!_this.shareDialog){
          _this.reload();
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
      async calculation(type){
        let _this = this
          _this.ruleForm.duration = _this.ruleForm.duration.replace(/[^\d.]/g,'')
          if(Number(_this.ruleForm.duration) > 540){
              _this.ruleForm.duration = '540'
          }else if(Number(_this.ruleForm.duration) < 180){
              _this.ruleForm.duration = '180'
          }else{
            return false
          }
          _this.ruleForm.duration_tip = true
          await _this.timeout(3000)
          _this.ruleForm.duration_tip = false

      },
      timeout (delay) {
          return new Promise((res) => setTimeout(res, delay))
      },
      async inputChange(val){
        if(val.indexOf('.') > -1){
          let array = val.split('.')
          this.ruleForm.price_tip = array[1].toString().length>18?true:false
        }else{
          this.ruleForm.price_tip = false
        }
      },
      async inputBlur(val, type){
        if(type == 1){
          const regexp=/(?:\.0*|(\.\d+?)0+)$/
          val = val.replace(/[^\d.]/g,'').replace(regexp,'$1')
          if(val.indexOf('.') > -1){
            let array = val.split('.')
            array[0] = array[0]>0 ? array[0].replace(/\b(0+)/gi,"") : '0'
            this.ruleForm.price =  array[0] + '.' + array[1]
          }else{
            this.ruleForm.price =  val.replace(/\b(0+)/gi,"")
          }
          this.ruleForm.price_tip = false
        }else if(type == 2){
          this.ruleForm.minerId_tip = ''
        }
      },
      handleClick(tab, event) {
        this.activeOn = tab.name
        this.$refs['ruleForm'].resetFields();
      },
      handleTagsClose(tag) {
          this.offlineForm.tags.splice(this.offlineForm.tags.indexOf(tag), 1);
      },
      showInputTask() {
          let _this = this
          if(_this.offlineForm.tags.length<5){
              _this.inputVisibleTask = true;
              _this.$nextTick(_ => {
                  _this.$refs.saveTagInput.$refs.input.focus();
              });
          }
      },
      handleInputConfirmTask() {
          let inputValue = this.inputValueTask;
          if (inputValue) {
              this.offlineForm.tags.push(inputValue);
              this.offlineForm.tags = this.uniqueNew(this.offlineForm.tags)
          }
          this.inputVisibleTask = false;
          this.inputValueTask = '';
      },
      uniqueNew(arr) {
          const res = new Map();
          return arr.filter((arr) => !res.has(arr) && res.set(arr, 1));
      },
      submitofflineForm(formName) {
          let _this = this;
          _this.$refs[formName].validate((valid) => {
              if (valid) {
                  let params = {
                    'Task_Name': _this.offlineForm.task_name,
                    'Curated_Dataset': _this.offlineForm.curated_dataset,
                    'Description': _this.offlineForm.desc,
                    'Is_Public': String(_this.offlineForm.OpenBidType),
                    'Type': _this.offlineForm.type,
                    'Tags': _this.offlineForm.tags.join(',')
                  }
                  if(_this.offlineForm.OpenBidType == 1){
                    params.Expire_Days = _this.offlineForm.bidDay
                    params.Max_Price = _this.offlineForm.max_price
                    params.Min_Price = _this.offlineForm.min_price
                  }else{
                    if(!_this.offlineForm.providerId){
                      _this.offlineForm.providerId_tips = true
                      return false;
                    }else{
                      _this.offlineForm.providerId_tips = false
                    }
                    params.Miner_Id = _this.offlineForm.providerId
                  }
                  // Initiate request
                  let postUrl = _this.data_api + `/minio/offlinedeals/` + _this.postAdress
                  axios.post(postUrl, params, {headers: {
                       'Authorization':"Bearer "+ _this.$store.getters.accessToken
                  }}).then((response) => {
                      let json = response.data
                      if (json.status == 'success') {
                        _this.taskName = json.data.deals.data.taskname
                        _this.finishTransaction = true
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
      async sendGetRequest(apilink, jsonObject) {
          try {
              const response = await axios.get(apilink)
              return response.data
          } catch (err) {
              console.error(err)
          }
      },
      submitForm(formName) {
        this.$refs[formName].validate(async (valid) => {
          if (valid) {
            let _this = this
            _this.loadShare = true

            const minerIDResponse = await _this.sendGetRequest(`${process.env.BASE_API}miner/validate/${this.ruleForm.minerId}`)
            _this.ruleForm.minerId_tip = minerIDResponse.status == 'fail'?minerIDResponse.message:''
            
            if(_this.ruleForm.duration_tip || _this.ruleForm.minerId_tip) {
              _this.loadShare = false
              return false
            }

            let postUrl = ''
            if(_this.sendApi == 1){
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
                "Duration": String(_this.ruleForm.duration.replace(/[^\d.]/g,'')*24*60*2)   //（The number of days entered by the user on the UI needs to be converted into epoch to the backend. For example, 10 days is 10*24*60*2）
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
                _this.loadShare = false
            }).catch(function (error) {
                console.log(error);
                _this.loadShare = false
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
      },
      setCurrent(row) {
          this.$refs.singleTable.setCurrentRow(row);
      },
      handleCurrentChange(val) {
          let _this = this
          _this.currentRow = val;
          if(val && val.miner_id){
              //this.name = this.parentName + ' / ' + val.miner_id
              if(_this.activeOn == 'online'){
                _this.ruleForm.minerId =  val.miner_id
                _this.name = val.miner_id
                _this.ruleForm.minerId_tip = ''
              }else{
                _this.offlineForm.providerId =  val.miner_id
                _this.nameOffline = val.miner_id
              }
          }
      },
      handleOpenSubmenu(key) {
          let _this = this
          _this.tableData = []
          _this.loading = true
          _this.page = 0
          _this.total = 1
          _this.parentLi = key
          _this.parentName = key
          let postURL = process.env.BASE_API+'miners?location='+key+'&status=&sort_by=score&order=ascending&limit=20&offset=0'

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
              let postURL = process.env.BASE_API+'miners?location='+_this.parentName+'&status=&sort_by=score&order=ascending&limit=20&offset=0'
              // let postURL = _this.data_api + '/miners?location='+_this.parentName+'&status=&sort_by=score&order=ascending&limit=20&offset=0'
              // let postURL = 'https://api.filswan.com/miners?location='+_this.parentName+'&status=&sort_by=score&order=ascending&limit=20&offset=0'

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
                  return false
              }

              _this.loading = true
              // let postURL = _this.data_api + '/miners?location='+_this.parentName+'&status=&sort_by=score&order=ascending&limit=20&offset='+_this.page*20
              let postURL = process.env.BASE_API+'miners?location='+_this.parentName+'&status=&sort_by=score&order=ascending&limit=20&offset='+_this.page*20
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
    mounted() { that=this },
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
            position: relative;
            display: flex;
            //align-items: center;
            justify-content: center;
            flex-wrap: wrap;
            width: 100%;
            .el-tabs{
              position:absolute;
              top:0;
              right: 100%;
              .el-tabs__header.is-left{
                float: none;
                margin: 0;
                .el-tabs__active-bar.is-left{
                  width: 0;
                }

              }
              .el-tabs__item{
                background: #eee;
                border-top-left-radius: 0.05rem;
                border-bottom-left-radius: 0.05rem;
              }
              .el-tabs__item.is-active {
                   background: #fff;
               }
            }
            .el-row{
              position: relative;
              width: 450px;
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
                padding: 0 0.15rem;
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
                    font-family: inherit;
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
                    cursor: text;
                    transition: border-color;
                    transition-duration: .3s;
                    background-color: transparent;
                    font-size: 14px;
                    color: #303133;
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
                      .el-form-item__error{
                        padding: 0;
                        word-break: break-word;
                      }
                      ._error{
                        color: #F56C6C;
                        font-size: 12px;
                        line-height: 1.2;
                        padding: 0;
                        word-break: break-word;
                      }
                    }
                  }
                  .lineHeight{
                    .el-form-item__label{
                      line-height: 20px;
                      word-break: break-word;
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
                          // display: flex;
                          // justify-content: space-between;
                          // align-items: center;
                          position: relative;
                          border: 0;
                          height: 35px;
                          line-height: 35px;
                          padding: 0;
                          .el-input{
                             .el-input__inner{
                                padding: 0 0 0 0.1rem;
                                border: 0;
                                font-size: 14px;
                                color: #303133;
                                height: 35px;
                                line-height: 35px;
                             }
                          }
                          .span{
                            position: relative;
                            padding: 0 0.1rem;
                            display: block;
                            z-index: 9;
                          }
                          .el-submenu__icon-arrow{
                            position: absolute;
                            right: 0.1rem;
                            z-index: 8;
                          }
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
                width: 0;
                height: 100%;
                background-color: #eee;
              }
            }
          }
        }
      }
    }


.completed /deep/{
  .el-dialog{
    margin-top: 0 !important;
  }
  text-align: center;
  .el-dialog__header{
    display: none;
  }
  img{
    display: block;
    max-width: 100px;
    margin: auto;
  }
  h1{
    margin: 0rem auto 0.1rem;
    font-size: 0.32rem;
    font-weight: 500;
    line-height: 1.2;
    color: #191919;
    word-break: break-word;
  }
  h3, a{
    font-size: 0.16rem;
    font-weight: 500;
    line-height: 1.2;
    color: #191919;
    word-break: break-word;
  }
  a{
    text-decoration: underline;
    color: #007bff;
  }
  a.a-close{
    padding: 5px 45px;
    background: #5c3cd3;
    color: #fff;
    border-radius: 10px;
    cursor: pointer;
    margin: 0.2rem auto 0;
    display: block;
    width: max-content;
    text-decoration: unset;
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
          .el-tabs{
             top: auto;
             bottom: 100%;
             right: auto;
             left: 0;
            .el-tabs__header.is-left{
              display: flex;
              border-bottom: 1px solid #ccc;
            }
          }
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
