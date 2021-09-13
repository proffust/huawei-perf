package config

import(
  "io/ioutil"
  "fmt"
  "gopkg.in/yaml.v2"
)

type TGraphiteConfig struct {
  Address string `yaml:"address"`
}

type TDefaultHuaweiPerfConfig struct {
  Username string `yaml:"username"`
  Password string `yaml:"password"`
  Port int `yaml:"port"`
  Interval int `yaml:"interval"`
  Graphite TGraphiteConfig `yaml:"graphite"`
}

type THuaweiArray struct {
  Name string `yaml:"name"`
  Addresses []string `yaml:"addresses"`
}

type TGroupConfig struct {
  Groupname string `yaml:"groupname"`
  Arrays []THuaweiArray `yaml:"arrays"`
}

type TLoggerConfig struct {
  Name string `yaml:"logger"`
  File string `yaml:"file"`
  Level string `yaml:"level"`
  Encoding string `yaml:"encoding"`
}

type THuaweiPerfConfig struct {
  Default TDefaultHuaweiPerfConfig `yaml:"default"`
  Groups []TGroupConfig `yaml:"groups"`
  Loggers []TLoggerConfig `yaml:"logging"`
}

var HuaweiPerfConfig = THuaweiPerfConfig {
  Default: TDefaultHuaweiPerfConfig {
    Username: "",
    Password: "",
    Port: 8443,
    Interval: 60,
    Graphite: TGraphiteConfig {
      Address: "0.0.0.0:2003",
    },
  },
}

func CreateConfig(configPath *string) (err error){
  if *configPath!="" {
    buff, err := ioutil.ReadFile(*configPath)
    if err!=nil {
      fmt.Println("Error while read file ", err)
      return err
    }
    err = yaml.Unmarshal(buff, &HuaweiPerfConfig)
    if err!=nil {
      fmt.Println("Error while parse config ", err)
      return err
    }
  }
  return nil
}
