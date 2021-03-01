<template>
  <div id="line-chart" :style="{height:height,width:width}"/>
</template>

<script>
  import echarts from 'echarts'

  export default {
    name: 'LineChart',
    props: {
      width: {
        type: String,
        default: '100%'
      },
      height: {
        type: String,
        default: '350px'
      },
      // title,xAxis,series
      chartData: {
        type: Object,
        required: true
      }
    },
    data() {
      return {}
    },
    mounted() {
      this.initChart()
    },
    beforeDestroy() {
    },
    methods: {
      initChart() {
        let chart = echarts.init(this.$el)

        const option = {
          title: {
            text: this.chartData.title || '折线图'
          },
          tooltip: {
            trigger: 'axis'
          },
          legend: {},
          grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
          },
          // toolbox: {
          //   feature: {
          //     saveAsImage: {}
          //   }
          // },
          xAxis: {
            type: 'category',
            boundaryGap: false,
            data: this.chartData.xname
          },
          yAxis: {
            type: 'value'
          },
          series: [],
        };

        let xdata = this.chartData.xdata
        for (let key in xdata) {
          // window.console.log(key);
          let a = {
            name: key,
            type: 'line',
            data: xdata[key]
          };
          option.series.push(a)
        }

        chart.setOption(option)
      },
    }
  }
</script>
