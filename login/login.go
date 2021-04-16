package login
import (
  "net/http"
  "crypto/tls"
  //"fmt"
  "strconv"
  "bytes"
  "encoding/json"
  "io/ioutil"
  "github.com/sirupsen/logrus"
  "errors"
)
// TODO: продумать передачу дефолтных и уникальных для массива параметров
func Login(log *logrus.Logger, ArrayUsername *string, ArrayPassword *string, ArrayAddress string, ArrayPort *int) (*http.Cookie, string, string, error) {
  urlString := "https://" + ArrayAddress + ":"+ strconv.Itoa(*ArrayPort) + "/deviceManager/rest/xxxxx/sessions"
  //отключение проверки безопасности для client
  tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  client := &http.Client{Transport: tr}

  data := map[string]string{"username": *ArrayUsername, "password": *ArrayPassword, "scope": "0"}
  jsonValue, _ := json.Marshal(data)
  req, err := http.NewRequest("POST", urlString, bytes.NewBuffer(jsonValue))
  if err!=nil {
    log.Warning("Failed to create new http request: array address - ", ArrayAddress, ": Error: ", err)
    return nil, "", "", err
  }
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  resp, err := client.Do(req)
  if err!=nil {
    log.Warning("Failed to do client request: array address - ", ArrayAddress, ": Error: ", err)
    return nil, "", "", err
  }

  var buf []byte
  buf, err = ioutil.ReadAll(resp.Body)
  if err!=nil{
    log.Warning("Failed to read response body: array address - ", ArrayAddress, ": Error: ", err)
    return nil, "", "", err
  }

  var raw map[string]interface{}
  json.Unmarshal(buf, &raw)
  if raw["error"].(map[string]interface{})["code"].(float64)==0 {
    return resp.Cookies()[0], raw["data"].(map[string]interface{})["iBaseToken"].(string), raw["data"].(map[string]interface{})["deviceid"].(string), nil
  }
  err = errors.New(raw["error"].(map[string]interface{})["description"].(string))
  log.Warning("Array address - ", ArrayAddress, ": Error: ", err)
  return nil, "", "", err
}

func Logout(log *logrus.Logger, ArrayAddress string, ArrayPort *int, Token *string, DeviceID *string, reqCookie *http.Cookie) error {
  urlString := "https://" + ArrayAddress + ":"+ strconv.Itoa(*ArrayPort) + "/deviceManager/rest/" + *DeviceID + "/sessions"
  tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
  client := &http.Client{Transport: tr}
  req, err := http.NewRequest("DELETE", urlString, nil)
  if err!=nil{
    log.Warning("Failed to do client request: array address - ", ArrayAddress, ": Error: ", err)
    return err
  }
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  req.Header.Add("iBaseToken", *Token)
  req.AddCookie(reqCookie)
  resp, err := client.Do(req)
  if err!=nil {
    log.Warning("Failed to do client request: array address - ", ArrayAddress, ": Error: ", err)
    return err
  }

  buf, err := ioutil.ReadAll(resp.Body)
  if err!=nil {
    log.Warning("Failed to read response body: array address - ", ArrayAddress, ": Error: ", err)
    return err
  }

  var raw map[string]interface{}
  json.Unmarshal(buf, &raw)
  if raw["error"].(map[string]interface{})["code"].(float64)!=0 {
    err = errors.New(raw["error"].(map[string]interface{})["description"].(string))
    log.Warning("Array address - ", ArrayAddress, ": Error: ", err)
    return err
  }
  return nil
}
