<template>
  <div>
    <el-row :gutter="10" class="mb10">
      <el-col :span="8">
        <el-card v-if="system.cpu" body-style="text-align: center;">
          <div slot="header">
            <span>CPU使用率</span>
          </div>

          <el-progress type="circle" :percentage="system.cpu.percent" style="margin-bottom: 20px"/>

          <Cell left="CPU主频" :right="system.cpu.ghz" divider/>
          <Cell left="系统负载" :right="system.sys.load"/>
        </el-card>
      </el-col>


      <el-col :span="8">
        <el-card v-if="system.mem" body-style="text-align: center;">
          <div slot="header">
            <span>内存使用率</span>
          </div>

          <el-progress type="circle" :percentage="system.mem.percent" style="margin-bottom: 20px"/>

          <Cell left="总内存" :right="system.mem.total" divider/>
          <Cell left="剩余内存" :right="system.mem.free"/>
        </el-card>
      </el-col>


      <el-col :span="8">
        <el-card v-if="system.disk" body-style="text-align: center;">
          <div slot="header">
            <span>磁盘信息</span>
          </div>

          <el-progress type="circle" :percentage="system.disk.percent" style="margin-bottom: 20px"/>

          <Cell left="总存储" :right="system.disk.total" divider/>
          <Cell left="剩余存储" :right="system.disk.free"/>
        </el-card>
      </el-col>

    </el-row>

    <el-card v-if="system.sys" style="margin-top: 10px">
      <div slot="header">
        <span>运行环境</span>
      </div>
      <Cell left="软件版本" :right="system.sys.appVersion" divider/>
      <Cell left="软件CommitId" :right="system.sys.appCommitId" divider/>
      <Cell left="软件BuildDate" :right="system.sys.appBuildDate" divider/>
      <Cell left="GO系统" :right="system.sys.goOs" divider/>
      <Cell left="GoArch" :right="system.sys.goArch" divider/>
      <Cell left="GO版本" :right="system.sys.goVersion" divider/>
      <Cell left="Goroutine" :right="system.sys.goroutine"/>
    </el-card>


    <el-card v-if="system.sys" style="margin-top: 10px">
      <div slot="header">
        <span>服务器信息</span>
      </div>

      <Cell left="机器名称" :right="system.sys.hostname" divider/>
      <Cell left="操作系统" :right="system.sys.platform" divider/>
      <Cell left="内核版本" :right="system.sys.kernel" divider/>
      <Cell left="CPU核心" :right="system.cpu.core" divider/>
      <Cell left="CPU" :right="system.cpu.modelName"/>
    </el-card>

  </div>
</template>

<script>

import Cell from "@/components/Cell";
import axios from "axios";

export default {
  name: 'Monitor',
  components: {Cell},
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['基础信息', '系统信息'])
  },
  mounted() {
    this.getData();

    // const chatTimer = setInterval(() => {
    //   this.getData();
    // }, 2000);
    //
    // this.$once('hook:beforeDestroy', () => {
    //   clearInterval(chatTimer);
    // });

  },
  data() {
    return {system: {}}
  },
  methods: {
    getData() {
      axios.get('/set/system', {}).then(resp => {
        var data = resp.data.data
        console.log(data);
        this.system = data;
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    }
  }
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


