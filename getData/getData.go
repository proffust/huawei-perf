package getData
import (
  "net/http"
  "strconv"
  "crypto/tls"
  "encoding/json"
  //"fmt"
  "io/ioutil"
  "strings"
  "github.com/proffust/huawei-perf/sendData"
  "reflect"
  "github.com/sirupsen/logrus"
  "errors"
  //"../sendData"
)
// TODO: конретезировать ошибки во всех методах
var sections = []string{"StoragePool", "Lun", "Controller", "fc_port", "disk"}
var statisticNameID = map[string]string{"22":"io_rate", "25":"read_io", "28":"write_io", "23":"read", "26":"write", "370":"resp_t", "384":"resp_t_r",
                                        "385":"resp_t_w", "93":"r_cache_hit", "95":"w_cache_hit", "68":"cpu_usage", "69":"cache_usage",
                                        "110":"r_cache_usage", "120":"w_cache_usage", "19":"queue_length", "182":"io_rate", "232":"read_io",
                                        "123":"read", "464":"resp_t_r", "233":"write_io", "124":"write", "465":"resp_t_w", "29":"resp_t", "18": "usage"}

func GetAllData(log *logrus.Logger, Groupname string, Devicename string, DeviceCookie *http.Cookie,
  DeviceToken *string, DeviceID *string, DevicePort *int, DeviceAddress string) {
  for _, section := range sections {
    metrics, err := getSectionData(log, Groupname, Devicename, &section, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress)
    if err==nil {
      go sendData.SendObjectPerfs(log, metrics)
    }
  }
}

func getSectionIDs(log *logrus.Logger, Section *string, Devicename string, DeviceCookie *http.Cookie,
  DeviceToken *string, DeviceID *string, DevicePort *int, DeviceAddress string) (int, (map[string]string), error) {

  tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  client := &http.Client{Transport: tr}

  urlString := "https://" + DeviceAddress + ":"+ strconv.Itoa(*DevicePort) + "/deviceManager/rest/" + *DeviceID +"/" + *Section
  req, err := http.NewRequest("GET", urlString, nil)
  if err!=nil {
    log.Warning("Failed to create GET http request, device: ", Devicename, ", section: ",  *Section, "; Error: ", err)
    return -1, nil, err
  }
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  req.Header.Add("iBaseToken", *DeviceToken)
  req.AddCookie(DeviceCookie)
  resp, err := client.Do(req)
  if err!=nil {
    log.Warning("Failed to do client GET request, device: ", Devicename, ", section: ",  *Section, "; Error: ", err)
    return -1, nil, err
  }

  result:= make(map[string]string)
  body, err := ioutil.ReadAll(resp.Body)
  if err!=nil {
    log.Warning("Failed to read response body, device: ", Devicename, ", section: ",  *Section, "; Error: ", err)
    return -1, nil, err
  }

  var ret map[string]interface{}
  json.Unmarshal(body, &ret)
  if ret["error"].(map[string]interface{})["code"].(float64)!=0 {
    err = errors.New(ret["error"].(map[string]interface{})["description"].(string))
    log.Warning("Device: ", Devicename, ", section: ",  *Section, "; Error: ", err)
    return -1, nil, err
  }

  if len(ret["data"].([]interface{}))==0 {
    err = errors.New("getSectionIDs: no data in the section "+*Section+" on the device "+DeviceAddress)
    log.Info("Error: ", err)
    return -1, nil, err
  }

  objectID := int(ret["data"].([]interface{})[0].(map[string]interface{})["TYPE"].(float64))
  for _, object := range ret["data"].([]interface{}) {
    if *Section=="disk" {
      ID, _ := object.(map[string]interface{})["ID"].(string)
      result[object.(map[string]interface{})["LOCATION"].(string)] = ID
    } else {
      if *Section=="Lun" {
        ID, _ := object.(map[string]interface{})["ID"].(string)
        result[object.(map[string]interface{})["PARENTNAME"].(string)+"."+object.(map[string]interface{})["NAME"].(string)] = ID
      } else {
        ID, _ := object.(map[string]interface{})["ID"].(string)
        result[object.(map[string]interface{})["NAME"].(string)] = ID
      }
    }
  }
  return objectID, result, nil
}

func getSectionPerfData(log *logrus.Logger, Section *string, Devicename string, DeviceCookie *http.Cookie, DeviceToken *string,
  DeviceID *string, DevicePort *int, DeviceAddress string, PerfIDs *string, Object *string) ([]string, []string, error) {
  tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  client := &http.Client{Transport: tr}

  urlString := "https://" + DeviceAddress + ":" + strconv.Itoa(*DevicePort) +
  "/deviceManager/rest/" + *DeviceID + "/performace_statistic/cur_statistic_data?CMO_STATISTIC_UUID=" + *Object +
  "&CMO_STATISTIC_DATA_ID_LIST=" + *PerfIDs
  req, err := http.NewRequest("GET", urlString, nil)
  if err!=nil {
    log.Warning("Failed to create GET http request: Error: ", err)
    return nil, nil, err
  }
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  req.Header.Add("iBaseToken", *DeviceToken)
  req.AddCookie(DeviceCookie)
  resp, err := client.Do(req)
  if err!=nil {
    log.Warning("Failed to do client request, device: ", Devicename, ", section: ",  *Section, "; Error: ", err)
    return nil, nil, err
  }

  body, err := ioutil.ReadAll(resp.Body)
  if err!=nil {
    log.Warning("Failed to read response body, device: ", Devicename, ", section: ",  *Section, "; Error: ", err)
    return nil, nil, err
  }

  var ret map[string]interface{}
  json.Unmarshal(body, &ret)
  if reflect.TypeOf(ret["data"])==reflect.TypeOf(ret["error"]) {
    err = errors.New("getSectionPerfData: no static data in the section "+*Section+" on the device "+Devicename)
    log.Info("Error: ", err)
    return nil, nil, err
  }

  perfArray := strings.Split(ret["data"].([]interface{})[0].(map[string]interface{})["CMO_STATISTIC_DATA_LIST"].(string), ",")
  metricArray := strings.Split(ret["data"].([]interface{})[0].(map[string]interface{})["CMO_STATISTIC_DATA_ID_LIST"].(string), ",")
  return perfArray, metricArray, nil
}

func getSectionData(log *logrus.Logger, Groupname string, Devicename string, Section *string,
  DeviceCookie *http.Cookie, DeviceToken *string, DeviceID *string, DevicePort *int, DeviceAddress string) (map[string]int, error) {

  result := make(map[string]int)
  sectionID, sectionNameIDs, err := getSectionIDs(log, Section, Devicename, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress)
  if err!=nil {
    return result, err
  }
  
  switch *Section {
    case "StoragePool":
      perfIDs := "22,25,28,370,384,385,23,26"
      for name, id := range sectionNameIDs {
        object := strconv.Itoa(sectionID) + ":" + id
        perfArray, metricArray, err := getSectionPerfData(log, Section, Devicename, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                        &object)
        if err!=nil{
          return result, err
        }

        for k,v := range metricArray {
          result[Groupname+"."+Devicename+"."+*Section+"."+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
        }
      }
      return result, nil
    case "Lun":
      perfIDs := "22,25,28,370,384,385,23,26,93,95,19"
      for name, id := range sectionNameIDs {
        object := strconv.Itoa(sectionID) + ":" + id
        perfArray, metricArray, err := getSectionPerfData(log, Section, Devicename, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                      &object)
        if err!=nil {
          return result, err
        }

        for k,v := range metricArray {
          result[Groupname+"."+Devicename+"."+*Section+"."+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
        }
      }
      return result, nil
    case "Controller":
      perfIDs := "22,25,28,370,384,385,23,26,93,95,68,69,110,120,19"
      for name, id := range sectionNameIDs {
        object := strconv.Itoa(sectionID) + ":" + id
        perfArray, metricArray, err := getSectionPerfData(log, Section, Devicename, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                        &object)
        if err!=nil {
          return result, err
        }

        for k,v := range metricArray {
          result[Groupname+"."+Devicename+"."+*Section+"."+Devicename +"_"+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
        }
      }
      return result, nil
    // TODO: нули в тропуте, выяснить почему
    case "fc_port":
      perfIDs := "22,25,28,370,384,385,23,26"
      for name, id := range sectionNameIDs {
        object := strconv.Itoa(sectionID) + ":" + id
        perfArray, metricArray, err := getSectionPerfData(log, Section, Devicename, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                        &object)
        if err!=nil {
          return result, err
        }

        for k,v := range metricArray {
          result[Groupname+"."+Devicename+"."+*Section+"."+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
        }
      }
      return result, nil
    case "disk":
      perfIDs := "18"
      for name, id := range sectionNameIDs {
        object := strconv.Itoa(sectionID) + ":" + id
        perfArray, metricArray, err := getSectionPerfData(log, Section, Devicename, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                        &object)
        if err!=nil {
          return result, err
        }
        name := strings.Replace(".", "_", -1)
        name := strings.Replace(" ", "_", -1)
        for k,v := range metricArray {
          result[Groupname+"."+Devicename+"."+*Section+"."+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
        }
      }
      return result, nil
  }
  return result, nil
}