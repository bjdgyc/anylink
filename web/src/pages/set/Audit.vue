<template>
  <div>
    <el-card>    
    <el-tabs v-model="activeName" @tab-click="handleClick">
        <el-tab-pane label="用户活动日志" name="act_log">
            <AuditActLog ref="auditActLog"></AuditActLog>
        </el-tab-pane>        
        <el-tab-pane label="用户访问日志" name="access_audit">
            <AuditAccess ref="auditAccess"></AuditAccess>
        </el-tab-pane>
    </el-tabs>
    </el-card>      
  </div>
</template>

<script>
import AuditAccess from "../../components/audit/Access";
import AuditActLog from "../../components/audit/ActLog";

export default {
  name: "Audit",
  components:{
    AuditAccess,
    AuditActLog
  },
  mixins: [],
  mounted() {    
    this.upTab();
  },  
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['基础信息', '审计日志'])        
  },
  data() {
    return {
      activeName: "act_log",
    }
  },
  methods: {  
    upTab() {
      var tabname = this.$route.query.tabname
      if (tabname) {
        this.activeName = tabname
      }
      this.handleClick(this.activeName)      
    },
    handleClick() {
        switch (this.activeName) {
        case "access_audit":
            this.$refs.auditAccess.setSearchData()
            this.$refs.auditAccess.getData(1)            
            break
        case "act_log":
            this.$refs.auditActLog.getData(1)
            break          
        }
        this.$router.push({path: this.$route.path, query: {tabname: this.activeName}})
    },    
  }
}
</script>
