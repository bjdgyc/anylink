// http://www.zhangwj.com/
// 全局的 axios 默认值
import axios from "axios";
import {getToken, removeToken} from "./token";
// axios.defaults.headers.common['Jwt'] = AUTH_TOKEN;
axios.defaults.headers.post['Content-Type'] = 'application/x-www-form-urlencoded';

if (process.env.NODE_ENV !== 'production') {
    // 开发环境
    axios.defaults.baseURL = 'http://172.23.83.233:8800';
}

function request(vm) {
    // HTTP 请求拦截器
    axios.interceptors.request.use(config => {
        // 在发送请求之前做些什么
        // 获取token, 并添加到 headers 请求头中
        const token = getToken();
        if (token) {
            config.headers.Jwt = token;
        }
        return config;
    });

    console.log(vm)

    // HTTP 响应拦截器
    // 统一处理 401 状态，token 过期的处理，清除token跳转login
    // 参数 1， 表示成功响应
    axios.interceptors.response.use(null, err => {
        // 没有登录或令牌过期
        if (err.response.status === 401) {
            // 注销，情况状态和token
            // vm.$store.dispatch("logout");
            // 跳转的登录页
            removeToken();
            vm.$router.push('/login');
            // 注意: 这里的 vm 实例需要外部传入
        }
        return Promise.reject(err);
    });
}

export default request




