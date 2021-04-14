package main
import (
  "github.com/proffust/huawei-perf/config"
  "github.com/proffust/huawei-perf/login"
  "github.com/proffust/huawei-perf/getData"
  "flag"
  "time"
  "runtime"
  //"log"
//  "net/http"
//  _ "net/http/pprof"
  "strconv"
  "github.com/sirupsen/logrus"
  "os"
)

const Version = "0.1.1"
// TODO: проблема при передаче по ссылке, выяснить почему, продумать передачу дефолтных и уникальных для массива параметров
func worker(log *logrus.Logger, username *string, password *string, address string, port *int, arrayname string, groupname string) {
  log.Info("Starting work with array ", arrayname)
  deviceCookie, deviceToken, deviceID, err := login.Login(log, username, password, address, port)
  if err==nil {
    log.Info("Login succesful on ", arrayname, " with username ", *username)
    getData.GetAllData(log, groupname, arrayname, deviceCookie, &deviceToken,
                             &deviceID, port, address)
    if err:=login.Logout(log, address, port, &deviceToken, &deviceID, deviceCookie); err==nil {
      log.Info("Succesful logout from array ", arrayname)
    }
  } else {
    log.Warning("Error while login on ", arrayname)
  }
  log.Info("End loop for array ", arrayname, " wait ", strconv.Itoa(config.HuaweiPerfConfig.Default.Interval), " seconds")
}

func main() {
  configPath := flag.String("config", "", "Path to the `config file`.")
  flag.Parse()
  log := logrus.New()

  //чтение конфигурационного файла
  if err:=config.CreateConfig(configPath); err!=nil {
    log.Fatal("Failed to get config file: Error: ", err)
    return
  }

  logLevels := map[string]logrus.Level{"trace": logrus.TraceLevel, "debug": logrus.DebugLevel, "info": logrus.InfoLevel, "warn": logrus.WarnLevel, "error": logrus.ErrorLevel, "fatal": logrus.FatalLevel, "panic": logrus.PanicLevel}
  formatters := map[string]logrus.Formatter{"json": &logrus.JSONFormatter{TimestampFormat: "02-01-2006 15:04:05"}, "text": &logrus.TextFormatter{TimestampFormat: "02-01-2006 15:04:05", FullTimestamp: true}}
  var writers []io.Writer
  var level logrus.Level
  var format logrus.Formatter

  for i, _ := range(config.HuaweiPerfConfig.Loggers){
    if config.HuaweiPerfConfig.Loggers[i].Name=="FILE"{
      file, err  := os.OpenFile(config.HuaweiPerfConfig.Loggers[i].File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
      if err!=nil{
        log.Warning("Failed to initialize log file: Error: ", err)
      }
      defer file.Close()
      writers = append(writers, file)
      level = logLevels[config.HuaweiPerfConfig.Loggers[i].Level]
      format = formatters[config.HuaweiPerfConfig.Loggers[i].Encoding]
    } else {
      writers = append(writers, os.Stdout)
      level = logLevels[config.HuaweiPerfConfig.Loggers[i].Level]
      format = formatters[config.HuaweiPerfConfig.Loggers[i].Encoding]
    }
  }

  //runtime.GOMAXPROCS(runtime.NumCPU())
  runtime.Gosched()
  log.Info("Starting performance monitoring, loaded config, graphite endpoint ", config.HuaweiPerfConfig.Default.Graphite.Address)

  var exit = make(chan bool)
  for {
    for _, group := range config.HuaweiPerfConfig.Groups {
      for _, array := range group.Arrays {
        go worker(log, &config.HuaweiPerfConfig.Default.Username, &config.HuaweiPerfConfig.Default.Password, array.Address,
                  &config.HuaweiPerfConfig.Default.Port, array.Name, group.Groupname)
      }
    }
    time.Sleep(time.Second*time.Duration(config.HuaweiPerfConfig.Default.Interval))
  }
  // TODO: сделать более простой и легкий способ считать горутины
  //log.Println(http.ListenAndServe("localhost:6060", nil))
  <-exit
}

func setValuesLogrus(log *logrus.Logger, level logrus.Level, output io.Writer, formatter logrus.Formatter){
  log.SetLevel(level)
  log.SetOutput(output)
  log.SetFormatter(formatter)
}