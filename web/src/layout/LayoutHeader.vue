<template>

  <div class="layout-header">
    <div>
      <i @click="toggleClick" :class="is_active ? 'el-icon-s-fold' : 'el-icon-s-unfold'" class="toggle-icon"
         style="font-size: 26px;"></i>

      <el-breadcrumb separator="/" class="app-breadcrumb">
        <el-breadcrumb-item v-for="(item, index) in route_name" :key="index">{{ item }}</el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <el-dropdown trigger="click" @command="handleCommand">
      <i class="el-icon-setting" style="margin-right: 15px"></i>
      <el-dropdown-menu slot="dropdown">
        <el-dropdown-item command="logout">退出</el-dropdown-item>
      </el-dropdown-menu>
      <span style="font-size: 12px;">{{ admin_user }}</span>
    </el-dropdown>

  </div>

</template>

<script>
import {getUser, removeToken} from "@/plugins/token";

export default {
  name: "Layoutheader",
  props: ['route_name'],
  data() {
    return {
      is_active: true
    }
  },
  computed: {
    admin_user() {
      return getUser();
    },
  },
  methods: {
    // 菜单栏开关按钮
    toggleClick() {
      this.is_active = !this.is_active
      // 触发事件,抛出到上层
      this.$emit('update:is_active', this.is_active)
    },
    handleCommand() {
      console.log("handleCommand")
      // 退出 删除登录信息
      removeToken()
      this.$router.push("/login");
    },
  }
}
</script>

<style scoped>
.layout-header {
  display: flex;
  justify-content: space-between;
  align-items: center
}

.toggle-icon {
  cursor: pointer;
  transition: background .3s;
  -webkit-tap-highlight-color: transparent;
}

.toggle-icon:hover {
  background: rgba(0, 0, 0, .025)
}

.app-breadcrumb {
  display: inline-block;
  font-size: 14px;
  /*line-height: 20;*/
  margin-left: 20px;
}

</style>
