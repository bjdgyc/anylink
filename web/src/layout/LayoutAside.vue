<template>

  <!--background-color="#304156"-->
  <!--text-color="#bfcbd9"-->
  <!--active-text-color="#409EFF"-->
  <!--:unique-opened="false"-->

  <!--<div class="layout-aside" :style="aside_style">-->

  <el-menu :collapse="!is_active"
           :default-active="route_path"
           :style="is_active?'width:200px':''"
           router
           class="layout-menu"
           :collapse-transition="false"

           background-color="#545c64"
           text-color="#fff"
           active-text-color="#ffd04b"
  >
    <el-menu-item index="/admin/home">
      <i class="el-icon-s-home"></i>
      <span slot="title">首页</span>
    </el-menu-item>

    <el-submenu index="1">
      <template slot="title">
        <i class="el-icon-menu"></i>
        <span slot="title">基础信息</span>
      </template>

      <el-menu-item index="/admin/set/system">系统信息</el-menu-item>
      <el-menu-item index="/admin/set/soft">软件配置</el-menu-item>
      <el-menu-item index="/admin/set/other">其他设置</el-menu-item>
    </el-submenu>

    <el-submenu index="2">
      <template slot="title">
        <i class="el-icon-s-custom"></i>
        <span slot="title">用户信息</span>
      </template>

      <el-menu-item index="/admin/user/list">用户列表</el-menu-item>
      <el-menu-item index="/admin/user/online">在线用户</el-menu-item>
      <el-menu-item index="/admin/user/ip_map">IP映射</el-menu-item>
    </el-submenu>

    <el-submenu index="3">
      <template slot="title">
        <i class="el-icon-s-order"></i>
        <span slot="title">用户组信息</span>
      </template>

      <el-menu-item index="/admin/group/list">用户组列表</el-menu-item>
    </el-submenu>


  </el-menu>

</template>

<script>
export default {
  name: "LayoutAside",
  data() {
    return {}
  },
  props: ['is_active', 'route_path'],
  mounted() {
  }
}
</script>

<style scoped>
.layout-menu {
  height: 100%;
}
</style>
