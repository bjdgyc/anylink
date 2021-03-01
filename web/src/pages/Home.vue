<template>
  <div class="home">
    <el-row :gutter="40" class="panel-group">
      <el-col :span="6" class="card-panel-col">
        <div class="card-panel">
          <i class="el-icon-user-solid" style="font-size:50px;color: #f4516c;"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">在线数</div>
            <countTo :startVal='0' :endVal='counts.online' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>

      <el-col :span="6" class="card-panel-col">
        <div class="card-panel">
          <i class="el-icon-user-solid" style="font-size:50px;color: #36a3f7"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">用户数</div>
            <countTo :startVal='0' :endVal='counts.user' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>

      <el-col :span="6" class="card-panel-col">
        <div class="card-panel">
          <i class="el-icon-wallet" style="font-size:50px;color:#34bfa3"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">用户组数</div>
            <countTo :startVal='0' :endVal='counts.group' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>

      <el-col :span="6" class="card-panel-col">
        <div class="card-panel">
          <i class="el-icon-s-order" style="font-size:50px;color:#40c9c6"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">IP映射数</div>
            <countTo :startVal='0' :endVal='counts.ip_map' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>

    </el-row>

    <el-row class="line-chart">
      <LineChart :chart-data="lineChartUser"/>
    </el-row>

    <el-row class="line-chart">
      <LineChart :chart-data="lineChartOrder"/>
    </el-row>

  </div>
</template>

<script>

import countTo from 'vue-count-to';
import LineChart from "@/components/LineChart";
import axios from "axios";

const lineChartUser = {
  title: '每日在线统计',
  xname: ['2019-12-13', '2019-12-14', '2019-12-15', '2019-12-16', '2019-12-17', '2019-12-18', '2019-12-19'],
  xdata: {
    'test1': [10, 120, 11, 134, 105, 10, 15],
    'test2': [10, 82, 91, 14, 162, 10, 15]
  }
}

const lineChartOrder = {
  title: '每日流量统计',
  xname: ['2019-12-13', '2019-12-14', '2019-12-15', '2019-12-16', '2019-12-17', '2019-12-18', '2019-12-19'],
  xdata: {
    'test1': [100, 120, 161, 134, 105, 160, 165],
    'test2': [120, 82, 91, 154, 162, 140, 145]
  }
}

export default {
  name: "Home",
  components: {
    LineChart,
    countTo,
  },
  data() {
    return {
      counts: {
        online: 0,
        user: 0,
        group: 0,
        ip_map: 0,
      },
      lineChartUser: lineChartUser,
      lineChartOrder: lineChartOrder,
    }
  },
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['首页'])
  },
  mounted() {
    this.getData()
  },
  methods: {
    getData() {
      axios.get('/set/home').then(resp => {
        var data = resp.data.data
        console.log(data);
        this.counts = data.counts
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    },
  },
}
</script>

<style scoped>
.card-panel {
  display: flex;
  justify-content: space-around;
  border: 1px solid red;
  padding: 30px 0;

  color: #666;
  background: #fff;
  /*box-shadow: 4px 4px 40px rgba(0, 0, 0, .05);*/
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, .12), 0 0 3px 0 rgba(0, 0, 0, .04);
  border-color: rgba(0, 0, 0, .05);
}

.card-panel-description {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: center;
}

.card-panel-text {
  line-height: 18px;
  color: rgba(0, 0, 0, .45);
  font-size: 16px;
}

.panel-num {
  font-size: 20px;
  font-weight: 700;
}

.line-chart {
  background: #fff;
  padding: 0 16px;
  margin-top: 40px;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, .12), 0 0 3px 0 rgba(0, 0, 0, .04);
  border-color: rgba(0, 0, 0, .05);
}
</style>
