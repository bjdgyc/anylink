<template>
  <div class="home">
    <el-row :gutter="40" class="panel-group">
      <el-col :span="6" class="card-panel-col">
        <div class="card-panel" v-on:click="jump('/admin/user/online')">
          <i class="el-icon-user-solid" style="font-size:50px;color: #f4516c;"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">在线数</div>
            <countTo :startVal='0' :endVal='counts.online' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>

      <el-col :span="6" class="card-panel-col">
        <div class="card-panel" v-on:click="jump('/admin/user/list')">
          <i class="el-icon-user-solid" style="font-size:50px;color: #36a3f7"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">用户数</div>
            <countTo :startVal='0' :endVal='counts.user' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>

      <el-col :span="6" class="card-panel-col">
        <div class="card-panel" v-on:click="jump('/admin/group/list')">
          <i class="el-icon-wallet" style="font-size:50px;color:#34bfa3"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">用户组数</div>
            <countTo :startVal='0' :endVal='counts.group' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>

      <el-col :span="6" class="card-panel-col">
        <div class="card-panel" v-on:click="jump('/admin/user/ip_map')">
          <i class="el-icon-s-order" style="font-size:50px;color:#40c9c6"></i>
          <div class="card-panel-description">
            <div class="card-panel-text">IP映射数</div>
            <countTo :startVal='0' :endVal='counts.ip_map' :duration='2000' class="panel-num"></countTo>
          </div>
        </div>
      </el-col>    

    </el-row>

    <el-row class="line-chart-box" gutter="20">
        <el-col :span="12" class="line-chart-col">            
            <LineChart :chart-data="lineChart.online"/>
            <div class="chart-group-name">
                <el-select size="mini" v-model="lineChartGroup.online" @change="lineChartGroupChange('online')">
                    <el-option v-for="(item,index) in groupNames" :key="index" :label="item.text" :value="item.value">
                    </el-option>
                </el-select> 
            </div>           
            <div class="time-range">
                <el-radio-group v-model="lineChartScope.online" size="mini" @change="((label)=>{lineChartScopeChange('online', label)})">
                    <el-radio-button label="rt" >实时</el-radio-button>
                    <el-radio-button label="1h">1小时</el-radio-button>
                    <el-radio-button label="24h">24小时</el-radio-button>
                    <el-radio-button label="7d">7天</el-radio-button>
                    <el-radio-button label="30d">30天</el-radio-button>
                </el-radio-group>
            </div>
        </el-col>
        <el-col :span="12" class="line-chart-col">
            <LineChart :chart-data="lineChart.network"/>
            <div class="chart-group-name">
                <el-select size="mini" v-model="lineChartGroup.network" @change="lineChartGroupChange('network')">
                    <el-option v-for="(item,index) in groupNames" :key="index" :label="item.text" :value="item.value">
                    </el-option>
                </el-select> 
            </div>            
            <div class="time-range">
                <el-radio-group v-model="lineChartScope.network" size="mini" @change="((label)=>{lineChartScopeChange('network', label)})">
                    <el-radio-button label="rt" >实时</el-radio-button>
                    <el-radio-button label="1h">1小时</el-radio-button>
                    <el-radio-button label="24h">24小时</el-radio-button>
                    <el-radio-button label="7d">7天</el-radio-button>
                    <el-radio-button label="30d">30天</el-radio-button>
                </el-radio-group>
            </div>            
        </el-col>
    </el-row>

    <el-row class="line-chart-box" gutter="20">
        <el-col :span="12" class="line-chart-col">            
            <LineChart :chart-data="lineChart.cpu"/>
            <div class="time-range">
                <el-radio-group v-model="lineChartScope.cpu" size="mini" @change="((label)=>{lineChartScopeChange('cpu', label)})">
                    <el-radio-button label="rt" >实时</el-radio-button>
                    <el-radio-button label="1h">1小时</el-radio-button>
                    <el-radio-button label="24h">24小时</el-radio-button>
                    <el-radio-button label="7d">7天</el-radio-button>
                    <el-radio-button label="30d">30天</el-radio-button>
                </el-radio-group>
            </div>
        </el-col>
        <el-col :span="12" class="line-chart-col">
            <LineChart :chart-data="lineChart.mem"/>
            <div class="time-range">
                <el-radio-group v-model="lineChartScope.mem" size="mini" @change="((label)=>{lineChartScopeChange('mem', label)})">
                    <el-radio-button label="rt" >实时</el-radio-button>
                    <el-radio-button label="1h">1小时</el-radio-button>
                    <el-radio-button label="24h">24小时</el-radio-button>
                    <el-radio-button label="7d">7天</el-radio-button>
                    <el-radio-button label="30d">30天</el-radio-button>
                </el-radio-group>
            </div>            
        </el-col>
    </el-row>    
  </div>
</template>

<script>

import countTo from 'vue-count-to';
import LineChart from "@/components/LineChart";
import axios from "axios";

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
      groupNames:[],
      lineChart: {
        online: {
            title: '用户在线数',
            xname: [],
            xdata: {
                '在线人数': [],
            },
            yminInterval: 1,
            yname:"人数"
        },
        network: {
            title: '网络吞吐量',
            xname: [],
            xdata: {
                '下行流量': [],
                '上行流量': [],                
            },
            yname:"Mbps"
        },
        cpu: {
            title: 'CPU占用比例',
            xname: [],
            xdata: {
                'CPU': [],
            },
            yname:"%"
        },
        mem: {
                title: '内存占用比例',
                xname: [],
                xdata: {
                    '内存': [],
                },
                yname:"%"
        }
      },
      lineChartScope : { 
            online: "rt",
            network : "rt",
            cpu : "rt",
            mem : "rt"  
      }, 
      lineChartGroup : {
            online: "",
            network: "",
      }  
    }
  },
  created() {
    this.$emit('update:route_path', this.$route.path)
    this.$emit('update:route_name', ['首页'])
  },
  mounted() {
    this.getData()
    this.getGroups()
    this.getAllStats() 
    const chartsTimer = setInterval(() => {
        this.getAllStats()                                      
    }, 10000);
    this.$once('hook:beforeDestroy', () => {
      clearInterval(chartsTimer);
    })    
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
    getAllStats() {
        for (var action in this.lineChartScope){
           if (this.lineChartScope[action] == "rt") {
               this.getStatsData(action);
           }
        }   
    },
    getStatsData(action, scope) {
        if (!scope) {
            scope = "rt"
        }
        let getData = {params:{"action": action, "scope": scope}}
        axios.get('/statsinfo/list', getData).then(resp => {
            var data = resp.data.data
            if (! data.datas) return ;
            data.action = action
            data.scope = scope
            switch(action) {
                case "online": this.formatOnline(data); break;
                case "network": this.formatNetwork(data); break;
                case "cpu": this.formatCpu(data); break;
                case "mem": this.formatMem(data); break;
            }
        }).catch(error => {
            this.$message.error('哦，请求出错');
            console.log(error);
        });
    },
    formatOnline(data) {
        let timeFormat = this.getTimeFormat(data.scope)
        let chartData = this.lineChart[data.action]
        let chooseGroup = this.lineChartGroup[data.action]
        let datas = data.datas
        let xnum = 0     
        chartData.xname = []
        chartData.xdata["在线人数"] = []
        for(var i=0; i<datas.length;i++){
            chartData.xname[i] = this.dateFormat(datas[i].created_at, timeFormat)  
            xnum = datas[i].num
            if (chooseGroup != "" && xnum > 0) {
                let num_groups = JSON.parse(datas[i].num_groups)
                xnum = ! num_groups[chooseGroup] ? 0 : num_groups[chooseGroup]
            }
            chartData.xdata["在线人数"][i] = xnum
        }
        // 实时更新在线数
        if (data.scope == "rt" && chooseGroup == "") {
            this.counts.online = datas[datas.length - 1].num
        }
        this.lineChart[data.action] = chartData
    },    
    formatNetwork(data) {
        let timeFormat = this.getTimeFormat(data.scope)
        let chartData = this.lineChart[data.action]
        let chooseGroup = this.lineChartGroup[data.action]
        let datas = data.datas
        let xnumUp = 0, xnumDown = 0
        chartData.xname = []
        chartData.xdata["上行流量"] = []
        chartData.xdata["下行流量"] = []
        for(var i=0; i<datas.length;i++){
            chartData.xname[i] = this.dateFormat(datas[i].created_at, timeFormat)
            xnumUp = datas[i].up
            xnumDown = datas[i].down             
            if (chooseGroup != "") {
                if (xnumUp > 0) {
                    let upGroups = JSON.parse(datas[i].up_groups)
                    xnumUp = ! upGroups[chooseGroup] ? 0 : upGroups[chooseGroup]
                }
                if (xnumDown > 0) {
                    let downGroups = JSON.parse(datas[i].down_groups)
                    xnumDown = ! downGroups[chooseGroup] ? 0 : downGroups[chooseGroup]
                }
            }
            chartData.xdata["上行流量"][i] = this.toMbps(xnumUp)
            chartData.xdata["下行流量"][i] = this.toMbps(xnumDown)
        }
        this.lineChart[data.action] = chartData
    },
    formatCpu(data) {
        let timeFormat = this.getTimeFormat(data.scope)
        let chartData = this.lineChart[data.action]
        let datas = data.datas        
        chartData.xname = []
        chartData.xdata["CPU"] = []        
        for(var i=0; i<datas.length;i++){
            chartData.xname[i] = this.dateFormat(datas[i].created_at, timeFormat)
            chartData.xdata["CPU"][i] = this.toDecimal(datas[i].percent)
        }
        this.lineChart[data.action] = chartData
    }, 
    formatMem(data) {
        let timeFormat = this.getTimeFormat(data.scope)
        let chartData = this.lineChart[data.action]
        let datas = data.datas        
        chartData.xname = []
        chartData.xdata["内存"] = []        
        for(var i=0; i<datas.length;i++){
            chartData.xname[i] = this.dateFormat(datas[i].created_at, timeFormat)
            chartData.xdata["内存"][i] = this.toDecimal(datas[i].percent)
        }
        this.lineChart[data.action] = chartData
    },  
    getTimeFormat(scope) {
        return (scope == "rt" || scope == "1h" || scope == "24h") ? "h:i:s" : "m/d h:i:s"
    },   
    toMbps(bytes) {
        if (bytes == 0) return 0
        return (bytes / Math.pow(1024, 2) * 8).toFixed(2) * 1
    },
    toDecimal(f) {
        return Math.floor(f * 100) / 100
    },
    lineChartScopeChange(action, label) {
        this.lineChartScope[action] = label;
        this.getStatsData(action, label);  
    },     
    dateFormat(p, format) {
        var da = new Date(p);
        var year = da.getFullYear();
        var month = da.getMonth() + 1;
        var dt = da.getDate();
        var h = ('0'+da.getHours()).slice(-2);
        var m = ('0'+da.getMinutes()).slice(-2)
        var s = ('0'+da.getSeconds()).slice(-2);
        switch (format) {
            case "h:i:s":  return h + ':' + m + ':' + s;
            case "m/d h:i:s":  return month + '/' + dt + ' ' + h + ':' + m + ':' + s;
        }
        return year + '-' + month + '-' + dt + ' ' + h + ':' + m + ':' + s;
    },    
    jump(path) {
        this.$router.push(path);
    },
    getGroups() {
      axios.get('/group/names_ids', {}).then(resp => {
        var data = resp.data.data
        var groupNames = []
        groupNames[0] = {text:"全部", value:""}
        for(var i=0; i<data.datas.length;i++){
            groupNames[i+1] = {text:data.datas[i].name, value:data.datas[i].id}
        }
        this.groupNames = groupNames
      }).catch(error => {
        this.$message.error('哦，请求出错');
        console.log(error);
      });
    }, 
    lineChartGroupChange(action) {
        this.getStatsData(action, this.lineChartScope[action]);
    }
  },
}
</script>

<style scoped>
.panel-group {
    margin-bottom: 20px;
}

.card-panel {
  display: flex;
  border-radius: 12px;
  justify-content: space-around;
  border: 1px solid red;
  padding: 20px 0;

  color: #666;
  background: #fff;
  /*box-shadow: 4px 4px 40px rgba(0, 0, 0, .05);*/
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, .12), 0 0 3px 0 rgba(0, 0, 0, .04);
  border-color: rgba(0, 0, 0, .05);
}

.card-panel:hover {
    box-shadow: 0 3px 5px 2px rgba(0, 0, 0, .12), 0 0 5px 0 rgba(0, 0, 0, .04);
    cursor:pointer;
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

.line-chart-box {
  padding: 0px 0px;
  margin-top: 10px; 
}

.line-chart-col {
    position: relative;
}

.line-chart {
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, .12), 0 0 3px 0 rgba(0, 0, 0, .04);
  border-color: rgba(0, 0, 0, .05);
  padding:5px 5px;
}

.time-range {
    position: absolute;
    right: 5px;
    top: 5px;
}

.chart-group-name {
    position: absolute;
    left: 110px;
    top: 5px;
    width: 130px;
    font-size: 10px;
}

/deep/ .el-radio-button--mini .el-radio-button__inner {
    padding: 7px 8px;    
    font-size: 10px;
}
</style>
