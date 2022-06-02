<template>
  <div>
    <el-card>

      <el-form :inline="true">
        <el-form-item>
          <el-button
              size="small"
              type="primary"
              icon="el-icon-plus"
              @click="handleEdit('')">添加
          </el-button>
        </el-form-item>
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
            prop="name"
            label="组名">
        </el-table-column>

        <el-table-column
            prop="note"
            label="备注">
        </el-table-column>

        <el-table-column
            prop="allow_lan"
            label="本地网络">
          <template slot-scope="scope">
            <el-switch
                v-model="scope.row.allow_lan"
                disabled>
            </el-switch>
          </template>
        </el-table-column>

        <el-table-column
            prop="bandwidth"
            label="带宽限制">
        </el-table-column>

        <el-table-column
            prop="client_dns"
            label="客户端DNS"
            width="160">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.client_dns" :key="inx">{{ item.val }}</el-row>
          </template>
        </el-table-column>

        <el-table-column
            prop="route_include"
            label="路由包含"
            width="200">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.route_include" :key="inx">{{ item.val }}</el-row>
          </template>
        </el-table-column>

        <el-table-column
            prop="route_exclude"
            label="路由排除"
            width="200">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.route_exclude" :key="inx">{{ item.val }}</el-row>
          </template>
        </el-table-column>

        <el-table-column
            prop="link_acl"
            label="LINK-ACL"
            min-width="200">
          <template slot-scope="scope">
            <el-row v-for="(item,inx) in scope.row.link_acl" :key="inx">
              {{ item.action }} => {{ item.val }} : {{ item.port }}
            </el-row>
          </template>
        </el-table-column>

        <el-table-column
            prop="status"
            label="状态"
            width="70">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.status === 1" type="success">可用</el-tag>
            <el-tag v-else type="danger">停用</el-tag>
          </template>

        </el-table-column>

        <el-table-column
            prop="updated_at"
            label="更新时间"
            :formatter="tableDateFormat">
        </el-table-column>

        <el-table-column
            label="操作"
            width="150">
          <template slot-scope="scope">
            <el-button
                size="mini"
                type="primary"
                @click="handleEdit(scope.row)">编辑
            </el-button>

            <el-popconfirm
                style="margin-left: 10px"
                @confirm="handleDel(scope.row)"
                title="确定要删除用户组吗？">
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
          :current-page="page"
          :total="count">
      </el-pagination>

    </el-card>

    <!--新增、修改弹出框-->
    <el-dialog
        :close-on-click-modal="false"
        title="用户组"
        :visible.sync="user_edit_dialog"
        width="750px"
        center>

      <el-form :model="ruleForm" :rules="rules" ref="ruleForm" label-width="100px" class="ruleForm">
        <el-tabs v-model="activeTab">
           <el-tab-pane label="通用" name="general">      
                <el-form-item label="用户组ID" prop="id">
                <el-input v-model="ruleForm.id" disabled></el-input>
                </el-form-item>

                <el-form-item label="组名" prop="name">
                <el-input v-model="ruleForm.name" :disabled="ruleForm.id > 0"></el-input>
                </el-form-item>

                <el-form-item label="备注" prop="note">
                <el-input v-model="ruleForm.note"></el-input>
                </el-form-item>

                <el-form-item label="带宽限制" prop="bandwidth">
                <el-input v-model.number="ruleForm.bandwidth">
                    <template slot="append">BYTE/S</template>
                </el-input>
                </el-form-item>
                <el-form-item label="本地网络" prop="allow_lan">
                <el-switch
                    v-model="ruleForm.allow_lan">
                </el-switch>
                </el-form-item>

                <el-form-item label="客户端DNS" prop="client_dns">
                <el-row class="msg-info">
                    <el-col :span="20">输入IP格式如: 192.168.0.10</el-col>
                    <el-col :span="4">
                    <el-button size="mini" type="success" icon="el-icon-plus" circle
                                @click.prevent="addDomain(ruleForm.client_dns)"></el-button>
                    </el-col>
                </el-row>
                <el-row v-for="(item,index) in ruleForm.client_dns"
                        :key="index" style="margin-bottom: 5px" :gutter="10">
                    <el-col :span="10">
                    <el-input v-model="item.val"></el-input>
                    </el-col>
                    <el-col :span="12">
                    <el-input v-model="item.note" placeholder="备注"></el-input>
                    </el-col>
                    <el-col :span="2">
                    <el-button size="mini" type="danger" icon="el-icon-minus" circle
                                @click.prevent="removeDomain(ruleForm.client_dns,index)"></el-button>
                    </el-col>
                </el-row>
                </el-form-item>
                <el-form-item label="状态" prop="status">
                    <el-radio-group v-model="ruleForm.status">
                        <el-radio :label="1" border>启用</el-radio>
                        <el-radio :label="0" border>停用</el-radio>
                    </el-radio-group>
                </el-form-item>            
            </el-tab-pane>

            <el-tab-pane label="认证方式" name="authtype">
                <el-form-item label="认证" prop="authtype">
                    <el-radio-group v-model="ruleForm.auth.type">
                        <el-radio label="local" border>本地</el-radio>
                        <el-radio label="radius" border>Radius</el-radio>
                    </el-radio-group>
                </el-form-item>               
                <el-form-item label="Radius密钥" v-if="ruleForm.auth.type == 'radius'">
                    <el-col :span="10">
                        <el-input v-model="ruleForm.auth.radius.secret"></el-input>
                    </el-col>
                </el-form-item>               
                <el-form-item label="Radius服务器" v-if="ruleForm.auth.type == 'radius'">
                    <el-col :span="10">
                        <el-input v-model="ruleForm.auth.radius.addr" placeholder="输入IP和端口 192.168.2.1:1812"></el-input>
                    </el-col>
                </el-form-item>                
            </el-tab-pane>            

            <el-tab-pane label="路由设置" name="route">
                <el-form-item label="包含路由" prop="route_include">
                <el-row class="msg-info">
                    <el-col :span="20">输入CIDR格式如: 192.168.1.0/24</el-col>
                    <el-col :span="4">
                    <el-button size="mini" type="success" icon="el-icon-plus" circle
                                @click.prevent="addDomain(ruleForm.route_include)"></el-button>
                    </el-col>
                </el-row>
                <el-row v-for="(item,index) in ruleForm.route_include"
                        :key="index" style="margin-bottom: 5px" :gutter="10">
                    <el-col :span="10">
                    <el-input v-model="item.val"></el-input>
                    </el-col>
                    <el-col :span="12">
                    <el-input v-model="item.note" placeholder="备注"></el-input>
                    </el-col>
                    <el-col :span="2">
                    <el-button size="mini" type="danger" icon="el-icon-minus" circle
                                @click.prevent="removeDomain(ruleForm.route_include,index)"></el-button>
                    </el-col>
                </el-row>
                </el-form-item>

                <el-form-item label="排除路由" prop="route_exclude">
                <el-row class="msg-info">
                    <el-col :span="20">输入CIDR格式如: 192.168.2.0/24</el-col>
                    <el-col :span="4">
                    <el-button size="mini" type="success" icon="el-icon-plus" circle
                                @click.prevent="addDomain(ruleForm.route_exclude)"></el-button>
                    </el-col>
                </el-row>
                <el-row v-for="(item,index) in ruleForm.route_exclude"
                        :key="index" style="margin-bottom: 5px" :gutter="10">
                    <el-col :span="10">
                    <el-input v-model="item.val"></el-input>
                    </el-col>
                    <el-col :span="12">
                    <el-input v-model="item.note" placeholder="备注"></el-input>
                    </el-col>
                    <el-col :span="2">
                    <el-button size="mini" type="danger" icon="el-icon-minus" circle
                                @click.prevent="removeDomain(ruleForm.route_exclude,index)"></el-button>
                    </el-col>
                </el-row>
                </el-form-item>
            </el-tab-pane>
            <el-tab-pane label="权限控制" name="link_acl">
                <el-form-item label="权限控制" prop="link_acl">
                <el-row class="msg-info">
                    <el-col :span="20">输入CIDR格式如: 192.168.3.0/24 端口0表示所有端口</el-col>
                    <el-col :span="4">
                    <el-button size="mini" type="success" icon="el-icon-plus" circle
                                @click.prevent="addDomain(ruleForm.link_acl)"></el-button>
                    </el-col>
                </el-row>

                <el-row v-for="(item,index) in ruleForm.link_acl"
                        :key="index" style="margin-bottom: 5px" :gutter="5">
                    <el-col :span="11">
                    <el-input placeholder="请输入CIDR地址" v-model="item.val">
                        <el-select v-model="item.action" slot="prepend">
                        <el-option label="允许" value="allow"></el-option>
                        <el-option label="禁止" value="deny"></el-option>
                        </el-select>
                    </el-input>
                    </el-col>
                    <el-col :span="3">
                    <el-input v-model.number="item.port" placeholder="端口"></el-input>
                    </el-col>
                    <el-col :span="8">
                    <el-input v-model="item.note" placeholder="备注"></el-input>
                    </el-col>
                    <el-col :span="2">
                    <el-button size="mini" type="danger" icon="el-icon-minus" circle
                                @click.prevent="removeDomain(ruleForm.link_acl,index)"></el-button>
                    </el-col>
                </el-row>
                </el-form-item>
            </el-tab-pane>

            <el-tab-pane label="域名拆分隧道" name="ds_domains">
                <el-form-item label="包含域名" prop="ds_include_domains">
                    <el-input type="textarea" :rows="5" v-model="ruleForm.ds_include_domains" placeholder="输入域名用,号分隔，默认匹配所有子域名, 如baidu.com,163.com"></el-input>
                </el-form-item>                
                <el-form-item label="排除域名" prop="ds_exclude_domains">
                    <el-input type="textarea" :rows="5" v-model="ruleForm.ds_exclude_domains" placeholder="输入域名用,号分隔，默认匹配所有子域名, 如baidu.com,163.com"></el-input>
                </el-form-item>
            </el-tab-pane>
            <el-form-item>
            <el-button type="primary" @click="submitForm('ruleForm')">保存</el-button>
            <el-button @click="disVisible">取消</el-button>
            </el-form-item>
        </el-form>        
    </el-dialog>

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
    this.$emit('update:route_name', ['用户组信息', '用户组列表'])
  },
  mounted() {
    this.getData(1);
    this.setAuthData();
  },
  data() {
    return {
      page: 1,
      tableData: [],
      count: 10,
      activeTab : "general",

      ruleForm: {
        bandwidth: 0,
        status: 1,
        allow_lan: true,
        client_dns: [{val: '114.114.114.114'}],
        route_include: [{val: 'all', note: '默认全局代理'}],
        route_exclude: [],
        link_acl: [],
        auth : {"type":'local'}
      },
      rules: {
        name: [
          {required: true, message: '请输入组名', trigger: 'blur'},
          {max: 30, message: '长度小于 30 个字符', trigger: 'blur'}
        ],
        bandwidth: [
          {required: true, message: '请输入带宽限制', trigger: 'blur'},
          {type: 'number', message: '带宽限制必须为数字值'}
        ],
        status: [
          {required: true}
        ],
      },
    }
  },
  methods: {
    setAuthData(row) {
        var defAuthData = {"type":'local', 
                           "radius":{"addr":"", "secret":""},
                          }
        if (this.ruleForm.auth.type == "local" || !row) {
            this.ruleForm.auth = defAuthData;
        }
    },
    handleDel(row) {
      axios.post('/group/del?id=' + row.id).then(resp => {
        const rdata = resp.data;
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
    handleEdit(row) {
      !this.$refs['ruleForm'] || this.$refs['ruleForm'].resetFields();
      console.log(row)
      this.activeTab = "general"
      this.user_edit_dialog = true
      if (!row) {
        this.setAuthData(row)
        return;
      }
      axios.get('/group/detail', {
        params: {
          id: row.id,
        }
      }).then(resp => {
        this.ruleForm = resp.data.data;
        this.setAuthData(resp.data.data);
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },
    pageChange(p) {
      this.getData(p)
    },
    getData(page) {
      this.page = page
      axios.get('/group/list', {
        params: {
          page: page,
        }
      }).then(resp => {
        const rdata = resp.data.data;
        console.log(rdata);
        this.tableData = rdata.datas;
        this.count = rdata.count
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },
    removeDomain(arr, index) {
      console.log(index)
      if (index >= 0 && index < arr.length) {
        arr.splice(index, 1)
      }
      // let index = arr.indexOf(item);
      // if (index !== -1 && arr.length > 1) {
      //   arr.splice(index, 1)
      // }
      // arr.pop()
    },
    addDomain(arr) {
      arr.push({val: "", action: "allow", port: 0});
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid, obj) => {
        if (!valid) {
          console.log('error submit!!');
          return false;
        }
        axios.post('/group/set', this.ruleForm).then(resp => {
          const rdata = resp.data;
          if (rdata.code === 0) {
            this.$message.success(rdata.msg);
            this.getData(1);
            this.user_edit_dialog = false
          } else {
            this.$message.error(rdata.msg);
          }
          console.log(rdata);
        }).catch(error => {
          this.$message.error('哦，请求出错');
          console.log(error);
        });
      });
    },
    resetForm(formName) {
      this.$refs[formName].resetFields();
    }
  },

}
</script>

<style scoped>
.msg-info {
  background-color: #f4f4f5;
  color: #909399;
  padding: 0 5px;
  margin: 0;
  box-sizing: border-box;
  border-radius: 4px;
  font-size: 12px;
}

.el-select {
  width: 80px;
}
</style>
