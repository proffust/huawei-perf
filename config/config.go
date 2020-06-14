package config

import(
  "bytes"
  "io/ioutil"
  "fmt"
  "github.com/spf13/viper"
)

type TGraphiteConfig struct {
  Address string
  Prefix string
}

type TDefaultHuaweiPerfConfig struct {
  Username string
  Password string
  Port int
  Interval int
  Graphite TGraphiteConfig
}

type THuaweiArray struct {
  Name string
  Address string
}

type TGroupConfig struct {
  Groupname string
  Arrays []THuaweiArray
}

type TLoggerConfig struct {
  Name string
  File string
  Level string
  Encoding string
}

type THuaweiPerfConfig struct {
  Default TDefaultHuaweiPerfConfig
  Groups []TGroupConfig
  LoggerConfig []TLoggerConfig
}

var HuaweiPerfConfig = THuaweiPerfConfig {
  Default: TDefaultHuaweiPerfConfig {
    Username: "",
    Password: "",
    Port: 8443,
    Interval: 60,
    Graphite: TGraphiteConfig {
      Address: "0.0.0.0:2003",
      Prefix: "huawei.perf",
    },
  },
}

func CreateConfig(configPath *string) {
  if *configPath!="" {
    buff, err := ioutil.ReadFile(*configPath)
    if err!=nil {
      fmt.Println("error while read file ", err)
    }
    viper.SetConfigType("YAML")
    err = viper.ReadConfig(bytes.NewBuffer(buff))
    if err!=nil {
      fmt.Println("error while parse file ", err)
    }
    err = viper.Unmarshal(&HuaweiPerfConfig)
    if err!=nil {
      fmt.Println("error while parse config ", err)
    }
  }
}
