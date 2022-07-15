<template>
  <div>
    <el-card>

      <el-table
          ref="multipleTable"
          :data="tableData"
          border>

        <el-table-column
            sortable="true"
            prop="id"
            label="ID"
            width="60">
        </el-table-column>

        <el-table-column
            prop="username"
            label="用户名"
            width="120">
        </el-table-column>

        <el-table-column
            prop="src"
            label="源IP地址"
            width="140">
        </el-table-column>

        <el-table-column
            prop="dst"
            label="目的IP地址"
            width="140">
        </el-table-column>

        <el-table-column
            prop="dst_port"
            label="目的端口"
            width="85">
        </el-table-column>

        <el-table-column
            prop="access_proto"
            label="访问协议"
            width="85"
            :formatter="protoFormat">
        </el-table-column>

        <el-table-column
            prop="info"
            label="详情">
        </el-table-column>        

        <el-table-column
            prop="created_at"
            label="创建时间"
            width="150"
            :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column
            label="操作"
            width="100">
          <template slot-scope="scope">
            <el-popconfirm
                class="m-left-10"
                @confirm="handleDel(scope.row)"
                title="确定要删除审计日志吗？">
              <el-button
                  slot="reference"
                  size="mini"
                  type="danger">删除
              </el-button>
            </el-popconfirm>

          </template>
        </el-table-column>
      </el-table>

      <div class="sh-20"></div>

      <el-pagination
          background
          layout="prev, pager, next"
          :pager-count="11"
          @current-change="pageChange"
          :total="count">
      </el-pagination>

    </el-card>

  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "Audit",
  components: {},
  mixins: [],
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['基础信息', '审计日志'])
  },
  mounted() {
    this.getData(1)
  },
  data() {
    return {
      tableData: [],
      count: 10,
      nowIndex: 0,
      accessProtoArr:["", "UDP", "TCP", "HTTPS", "HTTP"], 
    }
  },
  methods: {
    getData(p) {
      axios.get('/set/audit/list', {
        params: {
          page: p,
        }
      }).then(resp => {
        var data = resp.data.data
        console.log(data);
        this.tableData = data.datas;
        this.count = data.count
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },
    pageChange(p) {
      this.getData(p)
    },

    handleDel(row) {
      axios.post('/set/audit/del?id=' + row.id).then(resp => {
        var rdata = resp.data
        if (rdata.code === 0) {
          this.$message.success(rdata.msg);
          this.getData(1);
        } else {
          this.$message.error(rdata.msg);
        }
        console.log(rdata);
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },
    protoFormat(row) {
        var access_proto = row.access_proto
        if (row.access_proto == 0) {
            switch (row.protocol) {
                case 6: access_proto = 2; break;
                case 17: access_proto = 1; break;
            }
        }
        return this.accessProtoArr[access_proto]
    },
  },
}
</script>

<style scoped>

</style>
