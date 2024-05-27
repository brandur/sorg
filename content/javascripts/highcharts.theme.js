const fontFamily = "Helvetica, Arial, sans-serif";
const mainColor = '#4477aa';
const lightColor = '#ababab';

Highcharts.theme = {
  credits: {
    enabled: false
  },
  legend: {
    enabled: false
  },
  chart: {
     backgroundColor: null
  },
  subtitle: {
    style: {
      color: lightColor,
      fontFamily: fontFamily,
      fontSize: '12px'
    }
  },
  title: {
    style: {
      color: mainColor,
      fontFamily: fontFamily
    }
  },
  tooltip: {
    borderColor: '#000',
    borderWidth: 1,
    borderRadius: 0,
    shadow: false,
    style: {
      color: '#000',
      fontFamily: fontFamily,
      fontSize: '12px'
    }
  },
  xAxis: {
    labels: {
      style: {
        color: lightColor,
        fontFamily: fontFamily
      }
    },
    tickLength: 0,
  },
  yAxis: {
    gridLineColor: 'rgba(255, 255, 255, .1)',
    labels: {
      style: {
        color: lightColor,
        fontFamily: fontFamily
      }
    }
  }
};

highchartsOptions = Highcharts.setOptions(Highcharts.theme);
