<template>
  <div>
    <div class="monitor">
      <div class="monitor-left">{{ left }}</div>
      <div class="monitor-right">{{ right }}</div>
    </div>
    <el-divider v-if="divider"></el-divider>
  </div>
</template>

<script>
export default {
  name: "Cell",
  props: {
    left: {},
    right: {},
    divider: {type: Boolean}
  },
}
</script>

<style scoped>
.monitor {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.monitor-left {
  font-size: 14px;
}

.monitor-right {
  font-size: 12px;
  color: #909399;
}
</style>