<template>
  <div>
    <el-card>
      <el-table
          ref="multipleTable"
          :data="tableData"
          border>

        <el-table-column
            sortable="true"
            type="index"
            label="序号"
            width="50">
        </el-table-column>

        <el-table-column
            prop="username"
            label="用户名">
        </el-table-column>

        <el-table-column
            prop="group"
            label="登陆组">
        </el-table-column>

        <el-table-column
            prop="mac_addr"
            label="MAC地址">
        </el-table-column>
        
        <el-table-column
            prop="unique_mac"
            label="唯一MAC">
            <template slot-scope="scope">
                <el-tag v-if="scope.row.unique_mac" type="success">是</el-tag>
            </template>
        </el-table-column>

        <el-table-column
            prop="ip"
            label="IP地址"
            width="140">
        </el-table-column>

        <el-table-column
            prop="remote_addr"
            label="远端地址">
        </el-table-column>

        <el-table-column
            prop="tun_name"
            label="虚拟网卡">
        </el-table-column>

        <el-table-column
            prop="mtu"
            label="MTU">
        </el-table-column>

        <el-table-column
            prop="is_mobile"
            label="客户端">
          <template slot-scope="scope">
            <i v-if="scope.row.client === 'mobile'" class="el-icon-mobile-phone" style="font-size: 20px;color: red"></i>
            <i v-else class="el-icon-s-platform" style="font-size: 20px;color: blue"></i>
          </template>
        </el-table-column>

        <el-table-column
            prop="status"
            label="实时 上行/下行"
            width="220">
          <template slot-scope="scope">
            <el-tag type="success">{{ scope.row.bandwidth_up }}</el-tag>
            <el-tag class="m-left-10">{{ scope.row.bandwidth_down }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
            prop="status"
            label="总量 上行/下行"
            width="200">
          <template slot-scope="scope">
            <el-tag effect="dark" type="success">{{ scope.row.bandwidth_up_all }}</el-tag>
            <el-tag class="m-left-10" effect="dark">{{ scope.row.bandwidth_down_all }}</el-tag>
          </template>
        </el-table-column>

        <el-table-column
            prop="last_login"
            label="登陆时间"
            :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column
            label="操作"
            width="150">
          <template slot-scope="scope">
            <el-button
                size="mini"
                type="primary"
                @click="handleReline(scope.row)">重连
            </el-button>

            <el-popconfirm
                class="m-left-10"
                @confirm="handleOffline(scope.row)"
                title="确定要下线用户吗？">
              <el-button
                  slot="reference"
                  size="mini"
                  type="danger">下线
              </el-button>
            </el-popconfirm>

          </template>
        </el-table-column>
      </el-table>

    </el-card>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "Online",
  components: {},
  mixins: [],
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['用户信息', '在线用户'])
  },
  mounted() {
    this.getData();

    const chatTimer = setInterval(() => {
      this.getData();
    }, 10000);

    this.$once('hook:beforeDestroy', () => {
      clearInterval(chatTimer);
    })

  },
  data() {
    return {
      tableData: [],
    }
  },
  methods: {
    handleOffline(row) {
      axios.post('/user/offline?token=' + row.token).then(resp => {
        var data = resp.data
        if (data.code === 0) {
          this.$message.success(data.msg);
          this.getData();
        } else {
          this.$message.error(data.msg);
        }
        console.log(data);
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },

    handleReline(row) {
      axios.post('/user/reline?token=' + row.token).then(resp => {
        var data = resp.data
        if (data.code === 0) {
          this.$message.success(data.msg);
          this.getData();
        } else {
          this.$message.error(data.msg);
        }
        console.log(data);
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },

    handleEdit(a, row) {
      console.log(a, row)
    },
    getData() {
      axios.get('/user/online').then(resp => {
        var data = resp.data.data
        console.log(data);
        this.tableData = data.datas;
        this.count = data.count
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },
  },
}
</script>

<style scoped>

</style>
