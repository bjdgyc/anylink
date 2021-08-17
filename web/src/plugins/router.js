import Vue from "vue";
import VueRouter from "vue-router";
import {getToken} from "./token";

Vue.use(VueRouter)


const routes = [
    {path: '/login', component: () => import('@/pages/Login')},
    {
        path: '/admin',
        component: () => import('@/layout/Layout'),
        redirect: '/admin/home',
        children: [
            {path: 'home', component: () => import('@/pages/Home')},

            {path: 'set/system', component: () => import('@/pages/set/System')},
            {path: 'set/soft', component: () => import('@/pages/set/Soft')},
            {path: 'set/other', component: () => import('@/pages/set/Other')},
            {path: 'set/audit', component: () => import('@/pages/set/Audit')},

            {path: 'user/list', component: () => import('@/pages/user/List')},
            {path: 'user/online', component: () => import('@/pages/user/Online')},
            {path: 'user/ip_map', component: () => import('@/pages/user/IpMap')},

            {path: 'group/list', component: () => import('@/pages/group/List')},

        ],
    },

    {path: '*', redirect: '/admin/home'},
]

// 3. 创建 router 实例，然后传 `routes` 配置
// 你还可以传别的配置参数, 不过先这么简单着吧。
const router = new VueRouter({
    routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
    // 判断要进入的路由是否需要认证

    const token = getToken();

    console.log("beforeEach", from.path, to.path, token)
    // console.log(from)

    // 没有token,全都跳转到login
    if (!token) {
        if (to.path === "/login") {
            next();
            return;
        }

        next({
            path: '/login',
            query: {
                redirect: to.path
            }
        });
        return;
    }

    if (to.path === "/login") {
        next({path: '/admin/home'});
        return;
    }

    // 有token情况下
    next();
});

export default router;

