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
)
// TODO: конретезировать ошибки во всех методах
var sections = []string{"StoragePool", "Lun", "Controller", "fc_port"}
var statisticNameID = map[string]string{"22":"io_rate", "25":"read_io", "28":"write_io", "23":"read", "26":"write", "370":"resp_t", "384":"resp_t_r",
                                        "385":"resp_t_w", "93":"r_cache_hit", "95":"w_cache_hit", "68":"cpu_usage", "69":"cache_usage",
                                        "110":"r_cache_usage", "120":"w_cache_usage", "19":"queue_length", "182":"io_rate", "232":"read_io",
                                        "123":"read", "464":"resp_t_r", "233":"write_io", "124":"write", "465":"resp_t_w", "29":"resp_t"}

func GetAllData(Groupname string, Devicename string, DeviceCookie *http.Cookie,
  DeviceToken *string, DeviceID *string, DevicePort *int, DeviceAddress string) {
  for _, section := range sections {
    go sendData.SendObjectPerfs(getSectionData(Groupname, Devicename, &section, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress))
  }
}

func getSectionIDs(Devicename string, Section *string, DeviceCookie *http.Cookie,
  DeviceToken *string, DeviceID *string, DevicePort *int, DeviceAddress string) (int, (map[string]string)) {

  urlString := "https://" + DeviceAddress + ":"+ strconv.Itoa(*DevicePort) + "/deviceManager/rest/" + *DeviceID +"/" + *Section
  tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
  client := &http.Client{Transport: tr}
  req, err := http.NewRequest("GET", urlString, nil)
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  req.Header.Add("iBaseToken", *DeviceToken)
  req.AddCookie(DeviceCookie)
  resp, err := client.Do(req)
  if err!=nil {
    return -1, nil
  }
  result:= make(map[string]string)
  body, _ := ioutil.ReadAll(resp.Body)
  var ret map[string]interface{}
  json.Unmarshal(body, &ret)
  if ret["error"].(map[string]interface{})["code"].(float64)!=0 {
    return -1, nil
  }
  if len(ret["data"].([]interface{}))==0 {
    return -1, nil
  }
  objectID := int(ret["data"].([]interface{})[0].(map[string]interface{})["TYPE"].(float64))
  for _, object := range ret["data"].([]interface{}) {
    if *Section=="Disk" {
      ID, _ := object.(map[string]interface{})["ID"].(string)
      result[object.(map[string]interface{})["LOCATION"].(string)] = ID
    }else {
      if *Section=="Lun" {
        ID, _ := object.(map[string]interface{})["ID"].(string)
        result[object.(map[string]interface{})["PARENTNAME"].(string)+"."+object.(map[string]interface{})["NAME"].(string)] = ID
      }else {
        ID, _ := object.(map[string]interface{})["ID"].(string)
        result[object.(map[string]interface{})["NAME"].(string)] = ID
      }
    }
  }
  return objectID, result
}

func getSectionPerfData(Section *string, DeviceCookie *http.Cookie, DeviceToken *string,
  DeviceID *string, DevicePort *int, DeviceAddress string, PerfIDs *string, Object *string) ([]string, []string) {
  tr := &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  client := &http.Client{Transport: tr}
  urlString := "https://" + DeviceAddress + ":" + strconv.Itoa(*DevicePort) +
  "/deviceManager/rest/" + *DeviceID + "/performace_statistic/cur_statistic_data?CMO_STATISTIC_UUID=" + *Object +
  "&CMO_STATISTIC_DATA_ID_LIST=" + *PerfIDs
  req, _ := http.NewRequest("GET", urlString, nil)
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  req.Header.Add("iBaseToken", *DeviceToken)
  req.AddCookie(DeviceCookie)
  resp, err := client.Do(req)
  if err!=nil {
    return nil, nil
  }
  body, _ := ioutil.ReadAll(resp.Body)
  var ret map[string]interface{}
  json.Unmarshal(body, &ret)
  if reflect.TypeOf(ret["data"])==reflect.TypeOf(ret["error"]) {
    return nil, nil
  }
  perfArray := strings.Split(ret["data"].([]interface{})[0].(map[string]interface{})["CMO_STATISTIC_DATA_LIST"].(string), ",")
  metricArray := strings.Split(ret["data"].([]interface{})[0].(map[string]interface{})["CMO_STATISTIC_DATA_ID_LIST"].(string), ",")
  return perfArray, metricArray
}

func getSectionData(Groupname string, Devicename string, Section *string,
  DeviceCookie *http.Cookie, DeviceToken *string, DeviceID *string, DevicePort *int, DeviceAddress string) (map[string]int) {

  sectionID, sectionNameIDs := getSectionIDs(Devicename, Section, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress)
  if sectionID == -1 {
    return nil
  }
  result := make(map[string]int)
  switch *Section {
  case "StoragePool":
    perfIDs := "22,25,28,370,384,385,23,26"
    for name, id := range sectionNameIDs {
      object := strconv.Itoa(sectionID) + ":" + id
      perfArray, metricArray := getSectionPerfData(Section, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                      &object)
      for k,v := range metricArray {
        result[Groupname+"."+Devicename+"."+*Section+"."+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
      }
    }
    return result
  case "Lun":
    perfIDs := "22,25,28,370,384,385,23,26,93,95,19"
    for name, id := range sectionNameIDs {

      object := strconv.Itoa(sectionID) + ":" + id
      perfArray, metricArray := getSectionPerfData(Section, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                      &object)
      for k,v := range metricArray {
        result[Groupname+"."+Devicename+"."+*Section+"."+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
      }
    }
    return result
  case "Controller":
    perfIDs := "22,25,28,370,384,385,23,26,93,95,68,69,110,120,19"
    for name, id := range sectionNameIDs {
      object := strconv.Itoa(sectionID) + ":" + id
      perfArray, metricArray := getSectionPerfData(Section, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                      &object)
      for k,v := range metricArray {
        result[Groupname+"."+Devicename+"."+*Section+"."+Devicename +"_"+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
      }
    }
    return result
  // TODO: нули в тропуте, выяснить почему
  case "fc_port":
    perfIDs := "22,25,28,370,384,385,23,26"
    for name, id := range sectionNameIDs {
      object := strconv.Itoa(sectionID) + ":" + id
      perfArray, metricArray := getSectionPerfData(Section, DeviceCookie, DeviceToken, DeviceID, DevicePort, DeviceAddress, &perfIDs,
                                      &object)
      for k,v := range metricArray {
        result[Groupname+"."+Devicename+"."+*Section+"."+name+"."+statisticNameID[v]], _ = strconv.Atoi(perfArray[k])
      }
    }
    return result
  }
  return nil
}
