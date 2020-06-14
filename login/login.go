package login
import (
  "net/http"
  "crypto/tls"
  "fmt"
  "strconv"
  "bytes"
  "encoding/json"
  "io/ioutil"
)
// TODO: продумать передачу дефолтных и уникальных для массива параметров
func Login(ArrayUsername *string, ArrayPassword *string, ArrayAddress string, ArrayPort *int) (*http.Cookie, string, string) {
  urlString := "https://" + ArrayAddress + ":"+ strconv.Itoa(*ArrayPort) + "/deviceManager/rest/xxxxx/sessions"
  tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
  client := &http.Client{Transport: tr}
  data := map[string]string{"username": *ArrayUsername, "password": *ArrayPassword, "scope": "0"}
  jsonValue, _ := json.Marshal(data)
  req, err := http.NewRequest("POST", urlString, bytes.NewBuffer(jsonValue))
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  resp, err := client.Do(req)
  if err!=nil {
    fmt.Println(err)
    return nil, "", ""
  }
  var buf []byte
  buf, _ = ioutil.ReadAll(resp.Body)
  var raw map[string]interface{}
  json.Unmarshal(buf, &raw)
  if raw["error"].(map[string]interface{})["code"].(float64)==0 {
    return resp.Cookies()[0], raw["data"].(map[string]interface{})["iBaseToken"].(string), raw["data"].(map[string]interface{})["deviceid"].(string)
  }
  fmt.Println(raw["error"].(map[string]interface{})["description"].(string))
  return nil, "", ""
}

func Logout(ArrayAddress string, ArrayPort *int, Token *string, DeviceID *string, reqCookie *http.Cookie) float64 {
  urlString := "https://" + ArrayAddress + ":"+ strconv.Itoa(*ArrayPort) + "/deviceManager/rest/" + *DeviceID + "/sessions"
  tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
  client := &http.Client{Transport: tr}
  req, err := http.NewRequest("DELETE", urlString, nil)
  req.Header.Add("Content-Type", "application/json")
  req.Header.Add("Accept", "application/json")
  req.Header.Add("Connection", "keep-alive")
  req.Header.Add("iBaseToken", *Token)
  req.AddCookie(reqCookie)
  resp, err := client.Do(req)
  if err!=nil {
    return -1
  }
  buf, err := ioutil.ReadAll(resp.Body)
  // TODO: конретезировать ошибки
  if err!=nil {
    return -1
  }
  var raw map[string]interface{}
  json.Unmarshal(buf, &raw)
  return raw["error"].(map[string]interface{})["code"].(float64)
}
