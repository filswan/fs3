import { login, logout } from '@/api/login'
import { Message } from 'element-ui'

const user = {
  state: {
    name: sessionStorage.oaxLoginName || '',
    userId: sessionStorage.oaxLoginUserId || '',
    accessToken: sessionStorage.fs3MinioLoginAccessToken || ''
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
    // login
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
              _this.$router.push("/fs3")
              return false
            }
            _this.loginLoad = false
            if (response.success === true) {
              sessionStorage.oaxLoginUserId = response.data.userId
              sessionStorage.fs3MinioLoginAccessToken = response.data.accessToken
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
              message: 'Login failed',
              type: 'error',
              duration: 5 * 1000
            })
            console.log(error)
            reject(error)
          })
      })
    },

    // Get user information
    SetTime({ commit }, time) {
      return new Promise((resolve, reject) => {
        commit('SET_LINKTIME', time)
      })
    },
    // logout
    FedLogOut({ commit }) {
      // var _this = this
      return new Promise(resolve => {
        sessionStorage.removeItem('oaxLoginUserId')
        sessionStorage.removeItem('fs3MinioLoginAccessToken')
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

