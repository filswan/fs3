import { login, logout } from '@/api/login'
import { Message } from 'element-ui'

const user = {
  state: {
    name: sessionStorage.oaxLoginName || '',
    userId: sessionStorage.oaxLoginUserId || '',
    accessToken: sessionStorage.oaxMinioLoginAccessToken || ''
  },

  mutations: {
    SET_NAME: (state, name) => {
      state.name = name
    },
    SET_USERID: (state, userId) => {
      state.userId = userId
    },
    SET_ACCESSTOKEN: (state, accessToken) => {
      state.accessToken = accessToken
    }
  },

  actions: {
    // 登录
    Login({ commit }, userInfo) {
      var _this = this
      return new Promise((resolve, reject) => {
        login(userInfo)
          .then(response => {
            if (response.code === '10020') {
              Message({
                message: response.msg,
                type: 'error',
                duration: 5 * 1000
              })
              sessionStorage.oaxRegisterMail = userInfo.username
              _this.$router.push("/minio")
              return false
            }
            _this.loginLoad = false
            if (response.success === true) {
              sessionStorage.oaxLoginUserId = response.data.userId
              sessionStorage.oaxMinioLoginAccessToken = response.data.accessToken
              sessionStorage.oaxLoginName = response.data.name
              sessionStorage.oaxLoginEmail = response.data.email
              const data = response.data
              commit('SET_NAME', data.name)
              commit('SET_USERID', data.userId)
              commit('SET_ACCESSTOKEN', data.accessToken)
              newFunction(data)
              resolve()
            } else {
              Message({
                message: response.msg,
                type: 'error',
                duration: 5 * 1000
              })
            }
          })
          .catch(error => {
            _this.loginLoad = false
            Message({
              message: '登录失败',
              type: 'error',
              duration: 5 * 1000
            })
            console.log(error)
            reject(error)
          })
      })
    },

    // 获取用户信息
    SetTime({ commit }, time) {
      return new Promise((resolve, reject) => {
        commit('SET_LINKTIME', time)
      })
    },
    // 前端 登出
    FedLogOut({ commit }) {
      // var _this = this
      return new Promise(resolve => {
        sessionStorage.removeItem('oaxLoginUserId')
        sessionStorage.removeItem('oaxMinioLoginAccessToken')
        sessionStorage.removeItem('oaxLoginName')
        sessionStorage.removeItem('oaxLoginEmail')
        commit('SET_ACCESSTOKEN', '')
        resolve()
      })
    }
  }
}

export default user
function newFunction(data) {
  console.log(data.name)
}

