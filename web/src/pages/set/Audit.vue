<template>
  <div>
    <el-card>
      <el-form :model="searchForm" :rules="rules" ref="searchForm" :inline="true" class="form-inner-error">
        <el-form-item label="用户名:" prop="username">
          <el-input size="small" v-model="searchForm.username" style="width: 130px" @keydown.enter.native="searchEnterFun"></el-input>
        </el-form-item>
        <el-form-item label="源IP地址:" prop="src">
          <el-input size="small" v-model="searchForm.src" style="width: 130px" @keydown.enter.native="searchEnterFun"></el-input>
        </el-form-item>    
        <el-form-item label="目的IP地址:" prop="dst">
          <el-input size="small" v-model="searchForm.dst" style="width: 130px" @keydown.enter.native="searchEnterFun"></el-input>
        </el-form-item> 
        <el-form-item label="目的端口:" prop="dst_port">
          <el-input size="small" v-model="searchForm.dst_port" style="width: 80px" @keydown.enter.native="searchEnterFun"></el-input>
        </el-form-item> 
        <el-form-item label="访问协议：">
            <el-select size="small" v-model="searchForm.access_proto" style="width: 100px">
                    <el-option v-for="(item,index) in access_proto" :key="index" :label="item.text" :value="item.value">
                    </el-option>
            </el-select>           
        </el-form-item>  
        <div>
        <el-form-item label="日期范围：">
            <el-date-picker
                v-model="searchForm.date"
                type="datetimerange"
                size="small"
                value-format="yyyy-MM-dd HH:mm:ss"
                range-separator="～"
                start-placeholder="开始日期"
                end-placeholder="结束日期">
            >
            </el-date-picker>
        </el-form-item>
        <el-form-item label="详情:">
          <el-input size="small" v-model="searchForm.info" placeholder="请输入关键字" style="width: 200px" @keydown.enter.native="searchEnterFun"></el-input>
        </el-form-item>         
        <el-form-item>
          <el-button
              size="small"
              type="primary"
              icon="el-icon-search"
              @click="handleSearch">搜索
          </el-button>
          <el-button
              size="small"
              icon="el-icon-refresh"
              @click="rest">重置搜索
          </el-button>
        </el-form-item>
        </div>
      </el-form>

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
    this.setSearchData()
  },
  data() {
    return {
      tableData: [],
      count: 10,
      nowIndex: 0,
      accessProtoArr:["", "UDP", "TCP", "HTTPS", "HTTP"], 
      defSearchForm: {username:'', src:'', dst:'', dst_port:'', access_proto:'', info:'', date:["",""]},
      searchForm: {},
      access_proto: [
            { text: '请选择', value: '' },
            { text: 'UDP', value: '1' },
            { text: 'TCP', value: '2' },
            { text: 'HTTPS', value: '3' },
            { text: 'HTTP', value: '4' },
      ],
      rules: {
        username: [
          {max: 30, message: '长度小于 30 个字符', trigger: 'blur'}
        ],
        src: [
          {  message: '请输入正确的IP地址', validator: this.validateIP, trigger: 'blur' },
        ],        
        dst: [
          { message: '请输入正确的IP地址', validator: this.validateIP, trigger: 'blur' },
        ],               
      },            
    }
  },
  methods: {
    setSearchData() {
        this.searchForm = JSON.parse(JSON.stringify(this.defSearchForm));
    },    
    handleSearch() {
      this.$refs["searchForm"].validate((valid) => {
        if (!valid) {
          console.log('error submit!!');
          return false;
        }
        this.getData(1)
      })          
    },
    searchEnterFun(e) {
        var keyCode = window.event ? e.keyCode : e.which;
        if (keyCode == 13) {
            this.handleSearch()
        }
    },        
    getData(p) {
      if (! this.searchForm.date) {
        this.searchForm.date = ["", ""];
      }        
      axios.get('/set/audit/list', {
        params: {
          page: p,
          search: this.searchForm,
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
    rest() {
        console.log("rest");
        this.setSearchData();
        this.handleSearch();
    }, 
    validateIP(rule, value, callback) {
        if (value === '' || typeof value === 'undefined' || value == null) {
            callback()
        } else {
            const reg = /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/
            if ((!reg.test(value)) && value !== '') {
            callback(new Error('请输入正确的IP地址'))
            } else {
            callback()
            }
        }
    },        
  },
}
</script>

<style scoped>

</style>
