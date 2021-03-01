<template>
  <div id="app">
    <router-view></router-view>
  </div>
</template>

<script>

export default {
  name: 'app',
  components: {},
  created() {
    const token = sessionStorage.getItem('jwtToken');
    console.log("App created", token)
    if (token) {
      this.$root.isLogin = true
    }
  },
  mounted() {
    console.log("App mounted")
  },
  data() {
    return {}
  },
  computed: {
    currentComponent: function () {
      var isLogin = this.$root.isLogin
      console.log("App isLogin", isLogin)
      if (isLogin) {
        return "layout";
      }
      return "login";
    },
  },
}
</script>

<style>
html, body {
  height: 100%;
  margin: 0;
}

#app {
  font-family: "Helvetica Neue", Helvetica, "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "微软雅黑", Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  /*color: #2c3e50;*/
  /*border: 1px solid red;*/
  height: 100%;
  /*width:100%;*/
  /*box-sizing: border-box;*/
  /*padding: 4px;*/
}

.hide {
  display: none;
}

/*space vertical*/
.sh-10 {
  height: 10px;
}

.sh-20 {
  height: 20px;
}

/*space horizontal*/
.sw-10 {
  height: 1px;
  width: 10px;
}

.sw-20 {
  height: 1px;
  width: 20px;
}

.m-left-10 {
  margin-left: 10px;
}

</style>
