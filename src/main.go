package main

import (
  "fmt"
  "bytes"
  "encoding/json"
  "encoding/base64"
  "net/http"
  "errors"
  "io/ioutil"
  "time"
)

type Config struct {
  appkey            string    `json:"appkey"`
  appsecret         string    `json:"appsecret"`
  authManagerURL    string    `json:"authManagerURL"`
  chronosURL        string    `json:"chronosURL"`
  instanceID        string    `json:"instanceID"`
  status            string    `json:status"`
}

// file, _ := os.Open("conf.json")
// decoder := json.NewDecoder(file)
// config := Config{}
// err := decoder.Decode(&config)
// if err != nil {
//   fmt.Println("Error reading config: ", err)
//   return
// }

func getAuthString(config *Config) (string) {
  //var plainAuth = config.appkey + ":" + config.appsecret
  var plainAuth = config.appkey + ":" + config.appsecret
  authString := base64.URLEncoding.EncodeToString([]byte(plainAuth))
  return authString
}

func getToken(config *Config) (string, error) {
  payload := []byte(`{
      "Type": "Server"
  }`)

  client := http.Client{
    Timeout: time.Second * 2,
  }

  base64Auth := getAuthString(config)
  req, err := http.NewRequest(http.MethodPost,
                              config.authManagerURL + "/v1/login/",
                              bytes.NewBuffer(payload))
  if err != nil {
    fmt.Println("Failed to create request: ", err.Error())
    return "", err
  }

  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Basic " + base64Auth)

  res, getErr := client.Do(req)
  if getErr != nil {
    fmt.Println("Failed to execute request")
    return "", getErr
  }

  if res.StatusCode != 200 {
    return "", errors.New("Unauthorized")
  }

  body, readErr := ioutil.ReadAll(res.Body)
  if readErr != nil {
    fmt.Println("Failed to read response body")
    return "", readErr
  }

  result := make(map[string]interface{})
  err = json.Unmarshal(body, &result)
  if err != nil {
    return "", err
  }

  token, ok := result ["Token"].(string)
  if !ok {
    return "", errors.New("Invalid response type")
  }

  return token, nil
}

func updateJobStatus(config *Config) (string, error) {
  //chronos currently doesn't return anything when status is updated
  //the string return value can be used later if chronos responds with any data
  payload := []byte(`{
      "instance_id": "` + config.instanceID + `",
      "status": "` + config.status + `"
  }`)

  client := http.Client{
    Timeout: time.Second * 2,
  }

  token, err := getToken(config)
  if err != nil {
    fmt.Println("Failed to get token: ", err.Error())
    return "", err
  }

  req, err := http.NewRequest(http.MethodPost,
                              config.chronosURL + "/v1/jobcustomstatus",
                              bytes.NewBuffer(payload))
  if err != nil {
    fmt.Println("Failed to create request: ", err.Error())
    return "", err
  }

  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Bearer " + token)

  res, getErr := client.Do(req)
  if getErr != nil {
    fmt.Println("Failed to execute request")
    return "", getErr
  }

  //check if error code is in 400s
  if res.StatusCode/100 == 4 {
    fmt.Println("Error updating job status: unauthorized")
    return "", errors.New("Unauthorized")
  }

  body, readErr := ioutil.ReadAll(res.Body)
  if readErr != nil {
    fmt.Println("Failed to read response body")
    return "", readErr
  }

  result := make(map[string]interface{})
  err = json.Unmarshal(body, &result)
  if err != nil {
    return "", err
  }

  return "", nil
}

func getJobArgs(config *Config) (string, error) {
  client := http.Client{
    Timeout: time.Second * 2,
  }

  token, err := getToken(config)
  if err != nil {
    fmt.Println("Failed to get token: ", err.Error())
    return "", err
  }

  getArgsURL := config.chronosURL + "/v1/getargs/instanceid/" + config.instanceID
  req, err := http.NewRequest(http.MethodGet, getArgsURL, nil)
  if err != nil {
    fmt.Println("Failed to create request: ", err.Error())
    return "", err
  }
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Authorization", "Bearer " + token)

  res, getErr := client.Do(req)
  if getErr != nil {
    fmt.Println("Failed to execute request")
    return "", getErr
  }

  //check if error code is in 400s
  if res.StatusCode/100 == 4 {
    fmt.Println("Error updating job status: unauthorized")
    return "", errors.New("Unauthorized")
  }

  body, readErr := ioutil.ReadAll(res.Body)
  if readErr != nil {
    fmt.Println("Failed to read response body")
    return "", readErr
  }

  result := make(map[string]interface{})
  err = json.Unmarshal(body, &result)
  if err != nil {
    return "", err
  }

  args, ok := result["args"].(string)
  if !ok {
    return "", errors.New("Invalid response type")
  }
  return args, nil
}

func main() {
  config := Config{} //fill this in for testing 

  au := getAuthString(&config)
  fmt.Println("authtest: ", au)
  tok, _ := getToken(&config)
  fmt.Println("gettoken test: ", tok)
  updateJobStatus(&config)
  args, _ := getJobArgs(&config)
  fmt.Println(args)
  //fmt.Println("updatestatus: ", updateJobStatus(&config))
}
