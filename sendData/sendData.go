package sendData
import (
  "gopkg.in/fgrosse/graphigo.v2"
  "github.com/proffust/huawei-perf/config"
  "log"
)

func SendObjectPerfs(PerfMap map[string]int) int {
  if PerfMap==nil {
    return -1
  }
  Connection := graphigo.NewClient(config.HuaweiPerfConfig.Default.Graphite.Address)
  Connection.Prefix = config.HuaweiPerfConfig.Default.Graphite.Prefix
  Connection.Connect()
  for name, value := range PerfMap {
    err := Connection.Send(graphigo.Metric{Name: name, Value: value})
    if err!=nil {
      log.Println(err.Error())
    }
  }
  //err := Connection.SendAll(metrics)
  return 0
}
