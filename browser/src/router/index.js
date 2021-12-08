import Vue from 'vue'
import Router from 'vue-router'
Vue.use(Router)
// Route lazy loading
const home = () => import("@/components/Home");
const minio = () => import("@/views/minio/index");
const login = () => import("@/components/login");

const fs3_backup = () => import("@/components/fs3Backup");

const my_account = () => import("@/views/myAccount/index");
const my_account_dashboard = () => import("@/views/myAccount/dashboard/index");
const my_account_dashboard_detail = () => import("@/views/myAccount/dashboard/details");
const backupPlans = () => import("@/views/myAccount/backupPlans/index");
const myPlans = () => import("@/views/myAccount/myPlans/index");
const myJobs = () => import("@/views/myAccount/myJobs/index");

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
                {
                  path: '/fs3_backup',
                  name: 'fs3_backup',
                  component: fs3_backup
                },
                {
                  path: '/my_account',
                  component: my_account,
                  children: [
                    {
                      path: '/',
                      redirect: '/my_account/dashboard'
                    },
                    {
                      path: '/my_account/dashboard',
                      name: 'my_account_dashboard',
                      component: my_account_dashboard
                    },
                    {
                      path: '/my_account/dashboard_detail/:type',
                      name: 'my_account_dashboard_detail',
                      component: my_account_dashboard_detail
                    },
                    {
                      path: '/my_account/backupPlans',
                      name: 'my_account_backupPlans',
                      component: backupPlans
                    },
                    {
                      path: '/my_account/myPlans',
                      name: 'my_account_myPlans',
                      component: myPlans
                    },
                    {
                      path: '/my_account/jobs',
                      name: 'my_account_jobs',
                      component: myJobs
                    }
                  ]
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
