import Vue from 'vue'
import Router from 'vue-router'
Vue.use(Router)
// Route lazy loading
const home = () => import("@/components/Home");
const minio = () => import("@/views/minio/index");
const login = () => import("@/components/login");


export default new Router({
	// mode: 'history', // Back end support can be opened
	mode: 'hash',
	routes: [
        {
            path: '/',
            redirect: '/minio'
        },
        {
            path: '/',
            component: home,
            children: [
                {
                    path: '/minio',
                    name: 'minio',
                    component: minio,
                    beforeEnter: (to, from, next) => {
                      if (!sessionStorage.getItem('oaxMinioLoginAccessToken')) {
                        next({
                          path: '/minio/login',
                          query: { redirect: to.fullPath }
                        })
                      } else {
                        next()
                      }
                    }
                },
            ]
        },
        {
          path: '/minio/login',
          name: 'login',
          component: login,
        },
        {
            path: '*',
            redirect: '/'
        }
	]
})
const originalPush = Router.prototype.push
	Router.prototype.push = function push(location) {
    return originalPush.call(this, location).catch(err => err)
}
