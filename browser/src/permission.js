import router from './router'
import NProgress from 'nprogress' // Progress
import 'nprogress/nprogress.css' // Progress style
import store from './store'

NProgress.configure({
  showSpinner: false
})

const whiteList = ['/minio/login'] //Do not redirect whitelist
router.beforeEach((to, from, next) => {
  NProgress.start()
  store.state.user.linkPageName = null
  store.state.user.linkPageName = to.name
  if (sessionStorage.oaxLoginUserId) {
    if (to.path === '/minio/login') {
      next({
        path: '/'
      })
      NProgress.done()
    } else {
      next()
      NProgress.done()
    }
  } else {
    if (whiteList.indexOf(to.path) !== -1) { // In the login free white list, enter directly
      next()
    } else {
      next()
      NProgress.done()
    }
  }
})

router.afterEach(() => {
  NProgress.done() // Progress end
})
