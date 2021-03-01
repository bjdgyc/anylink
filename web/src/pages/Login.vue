<template>

  <div class="login">
    <el-card style="width: 550px;">

      <div class="issuer">AnyLink SSL VPN管理后台</div>

      <el-form :model="ruleForm" status-icon :rules="rules" ref="ruleForm" label-width="100px" class="ruleForm">
        <el-form-item label="管理用户名" prop="admin_user">
          <el-input v-model="ruleForm.admin_user"></el-input>
        </el-form-item>
        <el-form-item label="管理密码" prop="admin_pass">
          <el-input type="password" v-model="ruleForm.admin_pass" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="isLoading" @click="submitForm('ruleForm')">登录</el-button>
          <el-button @click="resetForm('ruleForm')">重置</el-button>
        </el-form-item>
      </el-form>

    </el-card>
  </div>

</template>

<script>
import axios from "axios";
import qs from "qs";
import {setToken, setUser} from "@/plugins/token";

export default {
  name: "Login",
  mounted() {
    // 进入login，删除登录信息
    console.log("login created")
    //绑定事件
    window.addEventListener('keydown', this.keyDown);
  },
  destroyed(){
    window.removeEventListener('keydown',this.keyDown,false);
  },
  data() {
    return {
      ruleForm: {},
      rules: {
        admin_user: [
          {required: true, message: '请输入用户名', trigger: 'blur'},
          {max: 50, message: '长度小于 50 个字符', trigger: 'blur'}
        ],
        admin_pass: [
          {required: true, message: '请输入密码', trigger: 'blur'},
          {min: 6, message: '长度大于 6 个字符', trigger: 'blur'}
        ],
      },
    }
  },
  methods: {
    keyDown(e) {
      //如果是回车则执行登录方法
      if (e.keyCode === 13) {
        this.submitForm('ruleForm');
      }
    },
    submitForm(formName) {
      this.$refs[formName].validate((valid) => {
        if (!valid) {
          console.log('error submit!!');
          return false;
        }
        this.isLoading = true

        // alert('submit!');
        axios.post('/base/login', qs.stringify(this.ruleForm)).then(resp => {
          var rdata = resp.data
          if (rdata.code === 0) {
            this.$message.success(rdata.msg);
            setToken(rdata.data.token)
            setUser(rdata.data.admin_user)
            this.$router.push("/home");
          } else {
            this.$message.error(rdata.msg);
          }
          console.log(rdata);
        }).catch(error => {
          this.$message.error('哦，请求出错');
          console.log(error);
        }).finally(() => {
              this.isLoading = false
            }
        );

      });
    },
    resetForm(formName) {
      this.$refs[formName].resetFields();
    },
  },
}
</script>

<style scoped>
.login {
  /*border: 1px solid red;*/
  height: 100%;
  /*margin: 0 auto;*/
  text-align: center;

  display: flex;
  justify-content: center;
  align-items: center;
}

.issuer {
  font-size: 26px;
  font-weight: bold;
  margin-bottom: 50px;
}
</style>