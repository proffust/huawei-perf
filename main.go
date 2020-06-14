package main
import (
  "github.com/proffust/huawei-perf/config"
  "github.com/proffust/huawei-perf/login"
  "github.com/proffust/huawei-perf/getData"
  "flag"
  "time"
  "runtime"
  "log"
//  "net/http"
//  _ "net/http/pprof"
  "strconv"
)

const Version = "0.1.1"
// TODO: проблема при передаче по ссылке, выяснить почему, продумать передачу дефолтных и уникальных для массива параметров
func worker(username *string, password *string, address string, port *int, arrayname string, groupname string) {
  for {
    log.Println("Starting work with array "+ arrayname)
    deviceCookie, deviceToken, deviceID := login.Login(username, password, address, port)
    if deviceCookie!=nil {
      log.Println("login succesful on "+arrayname+" with username "+*username)
      getData.GetAllData(groupname, arrayname, deviceCookie, &deviceToken,
                               &deviceID, port, address)
      if login.Logout(address, port, &deviceToken, &deviceID, deviceCookie)!=0 {
        log.Println("succesful logout from array "+arrayname)
      }
    }else {
      log.Println("error while login on "+arrayname)
    }
    log.Println("end loop for array "+arrayname+" wait "+strconv.Itoa(config.HuaweiPerfConfig.Default.Interval)+" seconds")
    time.Sleep(time.Second*time.Duration(config.HuaweiPerfConfig.Default.Interval))
  }
}

func main() {
  configPath := flag.String("config", "", "Path to the `config file`.")
  flag.Parse()
  config.CreateConfig(configPath)
  //runtime.GOMAXPROCS(runtime.NumCPU())
  runtime.Gosched()
  log.Println("Starting performance monitoring, loaded config, graphite endpoint "+config.HuaweiPerfConfig.Default.Graphite.Address)
  var exit = make(chan bool)
  for _, group := range config.HuaweiPerfConfig.Groups {
    for _, array := range group.Arrays {
      go worker(&config.HuaweiPerfConfig.Default.Username, &config.HuaweiPerfConfig.Default.Password, array.Address,
                &config.HuaweiPerfConfig.Default.Port, array.Name, group.Groupname)
    }
  }
  // TODO: сделать более простой и легкий способ считать горутины
  //log.Println(http.ListenAndServe("localhost:6060", nil))
  <-exit
}
