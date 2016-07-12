fontFamily = "Helvetica, Arial, sans-serif"
mainColor = '#4477aa'
lightColor = '#ababab'

Highcharts.theme =
  credits:
    enabled: false
  legend:
    enabled: false
  subtitle:
    style:
      color: lightColor
      fontFamily: fontFamily
      fontSize: '12px'
  title:
    style:
      color: mainColor
      fontFamily: fontFamily
  tooltip:
    borderColor: '#000'
    borderWidth: 1
    borderRadius: 0
    shadow: false
    style:
      color: '#000'
      fontFamily: fontFamily
      fontSize: '12px'
  xAxis:
    labels:
      style:
        color: lightColor
        fontFamily: fontFamily
  yAxis:
    gridLineColor: 'rgba(255, 255, 255, .1)'
    labels:
      style:
        color: lightColor
        fontFamily: fontFamily

# Apply the theme
highchartsOptions = Highcharts.setOptions(Highcharts.theme);
