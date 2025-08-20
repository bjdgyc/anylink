<template>
  <div class="login">
    <el-card style="width: 550px;">
      <div class="issuer">VPN账号密码重置申请</div>

      <el-form
        :model="form"
        status-icon
        :rules="rules"
        ref="form"
        label-width="120px"
        class="ruleForm"
      >
        <el-form-item label="注册邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入注册邮箱"></el-input>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="submitForm('form')"
          >
            发送重置邮件
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import axios from "axios";  // 移除了 qs 的导入

export default {
  name: "ForgotPassword",
  data() {
    return {
      loading: false,
      form: {
        email: ""
      },
      rules: {
        email: [
          { required: true, message: "请输入邮箱地址", trigger: "blur" },
          { type: "email", message: "请输入正确的邮箱格式", trigger: "blur" }
        ]
      }
    };
  },
  methods: {
    submitForm(formName) {
      this.$refs[formName].validate(valid => {
        if (!valid) return false;

        this.loading = true;
        // 修改后的 POST 请求
        axios.post("/user/reset/forgotPassword",
          this.form,  // 直接发送对象
          {
            headers: {
              'Content-Type': 'application/json'  // 明确指定 JSON 格式
            }
          }
        )
          .then(res => {
            if (res.data.code === 0) {
              this.$message.success("重置邮件已发送，请检查您的邮箱");
            } else {
              this.$message.error(res.data.msg);
            }
          })
          .catch(() => {
            this.$message.error("服务暂时不可用");
          })
          .finally(() => {
            this.loading = false;
          });
      });
    }
  }
};
</script>

<style scoped>
/* 样式保持不变 */
.login {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
}
.issuer {
  font-size: 24px;
  text-align: center;
  margin-bottom: 30px;
  color: #409EFF;
}
</style>
