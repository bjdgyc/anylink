<template>
  <div class="login">
    <el-card style="width: 550px;">
      <div class="issuer">VPN个人账号密码重置</div>
      <el-form
        :model="form"
        status-icon
        :rules="rules"
        ref="form"
        label-width="120px"
        class="ruleForm"
      >
        <el-form-item label="新密码" prop="password">
          <el-input
            type="password"
            v-model="form.password"
            autocomplete="off"
            show-password
            placeholder="请输入新密码,需要大小写并且8位以上"
          ></el-input>
        </el-form-item>

        <el-form-item label="确认密码" prop="checkPass">
          <el-input
            type="password"
            v-model="form.checkPass"
            autocomplete="off"
            show-password
            placeholder="请再次输入密码"
          ></el-input>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :loading="loading"
            @click="submitForm('form')"
          >
            确认重置
          </el-button>
<!--          <el-button @click="$router.push('/login')">返回登录</el-button>-->
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script>
import axios from "axios";

export default {
  name: "ResetPassword",
  data() {
    const validatePass = (rule, value, callback) => {
      if (value === "") {
        callback(new Error("请输入密码"));
      } else if (value.length < 8) {
        callback(new Error("密码长度不能小于8位"));
      } else if (
        !/[A-Z]/.test(value) ||  // 必须包含大写字母
        !/[a-z]/.test(value) ||  // 必须包含小写字母
        !/\d/.test(value)        // 必须包含数字
      ) {
        callback(new Error("密码必须包含大小写字母和数字"));
      } else {
        if (this.form.checkPass !== "") {
          this.$refs.form.validateField("checkPass");
        }
        callback();
      }
    };
    const validatePass2 = (rule, value, callback) => {
      if (value === "") {
        callback(new Error("请再次输入密码"));
      } else if (value !== this.form.password) {
        callback(new Error("两次输入密码不一致!"));
      } else {
        callback();
      }
    };

    return {
      loading: false,
      form: {
        password: "",
        checkPass: "",
        token: ""
      },
      rules: {
        password: [
          { validator: validatePass, trigger: "blur" }
        ],
        checkPass: [
          { validator: validatePass2, trigger: "blur" }
        ]
      }
    };
  },
  created() {
    // 从URL获取token参数
    this.form.token = this.$route.query.token;
    if (!this.form.token) {
      this.$message.error("无效的验证令牌");
    }
  },
  methods: {
    submitForm(formName) {
      this.$refs[formName].validate(valid => {
        if (!valid) return false;

        this.loading = true;
        axios.post("/user/reset/resetPassword",
          {
            password: this.form.password,
            token: this.form.token
          },
          {
            headers: {
              'Content-Type': 'application/json'
            }
          }
        )
          .then(res => {
            if (res.data.code === 0) {
              this.$message.success("密码重置成功,关闭此页面即可!");
              this.form.password = "";
              this.form.checkPass = "";
            } else {
              this.$message.error(res.data.msg || "重置失败");
            }
          })
          .catch(error => {
            let message = "服务暂时不可用";
            if (error.response) {
              switch (error.response.status) {
                case 401:
                  message = "验证令牌已过期";
                  break;
                case 400:
                  message = "无效的请求参数";
                  break;
              }
            }
            this.$message.error(message);
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
/* 保持相同样式 */
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
