import axios from 'axios' // axios  npm install axios
import QS from 'qs'
import store from '../store'
import router from '../router'
import { Message } from 'element-ui'


const service = axios.create({
	baseURL: process.env.NODE_ENV === 'production' ? process.env.BASE_API : '/api',
	timeout: 15000
})
service.interceptors.request.use(function (config) {
    if (sessionStorage.oaxLoginUserId) {
      config.baseURL === process.env.BASE_API
    }
    if (config.method === 'post') {
      // config.data = qs.stringify(config.data)
      // config.content-type = 'application/x-www-form-urlencoded'
    }
		config.headers['Authorization'] = "Bearer "+ store.getters.accessToken;
    sessionStorage.time = 70
    return config
}, function (error) {
	// Do something with request error
    console.log(error) // for debug
	return Promise.reject(error);
})
service.interceptors.response.use(response => {
	const res = response
    if (!res.data.success) {
      // -1:User not logged in;
      if (res.data.code === '-1') {
        store.dispatch('FedLogOut').then(() => {
			    router.push('/minio/login')
        })
      }
      return response.data
    } else {
      return response.data
	}
	return response.data
}, function (error) {
  // Failure handling
	console.log('responseError:' + error) // for debug
	Message({
		message: 'Error',
		type: 'error',
		duration: 5 * 1000
	  })
	return Promise.reject(error);
});

export default service;
