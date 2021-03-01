<template>
  <el-card>
      <el-table
          :data="soft_data"
          border>

        <el-table-column
            prop="info"
            label="信息"
            width="260">
        </el-table-column>

        <el-table-column
            prop="name"
            label="配置"
            width="200">
        </el-table-column>

        <el-table-column
            prop="data"
            label="数据">
          <template slot-scope="scope">
            {{ scope.row.data }}
          </template>
        </el-table-column>

      </el-table>
  </el-card>
</template>

<script>
import axios from "axios";

export default {
  name: "Soft",
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['基础信息', '软件配置'])
  },
  mounted() {
    this.getSoftInfo()
  },
  data() {
    return {soft_data: []}
  },

  methods: {
    getSoftInfo() {
      axios.get('/set/soft', {}).then(resp => {
        var data = resp.data
        console.log(data);
        this.soft_data = data.data;
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    }
  },
}
</script>

<style scoped>

</style>
