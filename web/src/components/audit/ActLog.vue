<template>
  <div>
      <el-form  :model="searchForm" ref="searchForm" :inline="true" class="search-form">
        <el-form-item>
          <el-input size="mini" v-model="searchForm.username" clearable placeholder="请输入用户名" style="width: 130px" @keydown.enter.native="searchEnterFun"></el-input>
        </el-form-item>
        <el-form-item>
                <el-date-picker
                    v-model="searchForm.sdate"
                    type="date"
                    size="mini"
                    placeholder="开始日期"
                    format="yyyy-MM-dd"
                    value-format="yyyy-MM-dd"  
                    style="width: 130px"
                >
                </el-date-picker>
            </el-form-item>
            <el-form-item>    
                <el-date-picker
                    v-model="searchForm.edate"
                    type="date"
                    size="mini"
                    placeholder="结束日期"
                    format="yyyy-MM-dd"
                    value-format="yyyy-MM-dd"  
                    style="width: 130px"              
                >
            </el-date-picker>
        </el-form-item>
        <el-form-item >
            <el-select size="mini" v-model="searchForm.status" clearable placeholder="操作类型" style="width: 130px">
                    <el-option v-for="(item,index) in statusOps" :key="index" :label="item.value" :value="item.key+1">
                    </el-option>
            </el-select>           
        </el-form-item>
        <el-form-item>
            <el-select size="mini" v-model="searchForm.os" clearable placeholder="操作系统" style="width: 130px">
                    <el-option v-for="(value,item,index) in osOps" :key="index" :label="value" :value="item+1">
                    </el-option>
            </el-select>           
        </el-form-item>        
        <el-form-item>
          <el-button
              size="mini"
              type="primary"
              icon="el-icon-search"
              @click="handleSearch">搜索
          </el-button>
          <el-button
              size="mini"
              icon="el-icon-refresh"
              @click="rest">重置搜索
          </el-button>
        </el-form-item>
      </el-form>

      <el-table
          ref="multipleTable"
          :data="tableData"
          :default-sort="{ prop: 'id', order: 'descending' }"
          @sort-change="sortChange"
          :header-cell-style="{backgroundColor:'#fcfcfc'}"
          border>

        <el-table-column
            prop="id"
            label="ID"
            sortable="custom"
            width="100">
        </el-table-column>
        <el-table-column
            prop="username"
            label="用户名"
            width="140">
        </el-table-column>
        <el-table-column
            prop="group_name"
            label="登陆组"
            width="100">
        </el-table-column>
        <el-table-column
            prop="status"
            label="操作类型"
            width="92">
                <template slot-scope="{ row }">
                    <span v-for="(item, index) in statusOps" :key="index">
                        <el-tag size="small" v-if="row.status == item.key" disable-transitions :type="item.tag">{{item.value}}</el-tag>
                    </span>
                </template>            
        </el-table-column> 
        <el-table-column
            prop="info"
            label="操作详情"
            min-width="200">
        </el-table-column> 
        <el-table-column
            prop="created_at"
            label="操作时间"
            width="150"
            :formatter="tableDateFormat">
        </el-table-column>                                          
        <el-table-column
            prop="os"
            label="操作系统"
            min-width="210">
                <template slot-scope="{ row }">
                    <span v-for="(value, item, index) in osOps" :key="index">
                    {{ row.os == item? value: "" }}
                    </span>
                    <div class="sub_txt">型号: 
                        <span v-if="row.device_type != ''">{{ row.device_type }} / {{ row.platform_version }}</span>
                        <span v-else> - </span>
                    </div>
                </template>             
        </el-table-column>         
        <el-table-column
            prop="client"
            label="客户端"
            width="150">
                <template slot-scope="{ row }">
                    <span v-for="(value, item, index) in clientOps" :key="index">
                    {{ row.client == item? value: "" }}
                    </span>
                    {{ row.version }} 
                </template>                           
        </el-table-column>  
        <el-table-column
            prop="ip_addr"
            label="内网IP"
            width="120">
        </el-table-column>
        <el-table-column
            prop="remote_addr"
            label="外网IP"
            width="120">
        </el-table-column>                                                  
      </el-table>
      <div class="sh-20"></div>
        <el-pagination
            background
            layout="prev, pager, next"  
            :pager-count="11"
            @current-change="pageChange"
            :current-page="page"
            :total="count">
        </el-pagination>
</div>
</template>

<script>
import axios from "axios";

export default {
  name: "List",
  components: {},
  mixins: [],
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['用户信息', '登入日志'])
  },
  data() {
    return {
      page: 1,
      grouNames: [],
      tableData: [],
      idSort: 1,
      count: 10,
      searchForm: {username:'', sdate:'', edate:'', status:'', os:''},
      statusOps:[],
      osOps:[],
      clientOps:[],                  
    }
  },
  watch: {
    idSort: {
        handler(newValue, oldValue) {
            if (newValue != oldValue) {
                this.getData(1);
            }
        },
    },
  },
  methods: {    
    handleSearch() {
      this.getData(1)
    },
    pageChange(p) {
      this.getData(p)
    },
    searchEnterFun(e) {
        var keyCode = window.event ? e.keyCode : e.which;
        if (keyCode == 13) {
            this.handleSearch()
        }
    },    
    getData(page) {
      console.log(this.searchForm)
      this.page = page
      axios.get('/set/audit/act_log_list', {
        params: {
          page: page,
          username: this.searchForm.username || '',
          sdate: this.searchForm.sdate || '',
          edate: this.searchForm.edate || '',
          status: this.searchForm.status || '',
          os: this.searchForm.os || '',
          sort: this.idSort,
        }
      }).then(resp => {
        var data = resp.data.data
        console.log(data);
        this.tableData = data.datas;
        this.count = data.count
        this.statusOps = data.statusOps
        this.osOps = data.osOps
        this.clientOps = data.clientOps
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },
    rest() {
        console.log("rest");
        this.searchForm.username = "";
        this.searchForm.sdate = "";
        this.searchForm.edate = "";
        this.searchForm.status = "";
        this.searchForm.os = "";
        this.handleSearch();
    },
    sortChange(column) {
        let { order } = column;
        if (order === 'ascending') {
            this.idSort = 2;
        } else {
            this.idSort = 1;
        }
    },    
  }
}
</script>

<style scoped>
.el-form-item {
    margin-bottom: 8px;
}
.el-table {
    font-size: 12px;
}
.search-form >>> .el-form-item__label {
  font-size: 12px;
}
/deep/ .el-table th {
    padding: 5px 0;
}
/deep/ .el-table td {
    padding: 5px 0;
}
.sub_txt {
    color: #88909B;
}
</style>