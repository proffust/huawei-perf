package sendData
import (
  "gopkg.in/fgrosse/graphigo.v2"
  "github.com/proffust/huawei-perf/config"
  //"log"
  "github.com/sirupsen/logrus"
)

func SendObjectPerfs(log *logrus.Logger, PerfMap map[string]int) int {
  if PerfMap==nil {
    return -1
  }
  Connection := graphigo.NewClient(config.HuaweiPerfConfig.Default.Graphite.Address)
  Connection.Prefix = config.HuaweiPerfConfig.Default.Graphite.Prefix
  Connection.Connect()
  for name, value := range PerfMap {
    err := Connection.Send(graphigo.Metric{Name: name, Value: value})
    if err!=nil {
      log.Warning("Failed to send metric: ", name, " = ", value, " :Error: ", err)
      continue
    }
    log.Debug("Metric sent successfully: ", name, " = ", value)
  }
  //err := Connection.SendAll(metrics)
  return 0
}
