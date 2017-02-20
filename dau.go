package main

import (
  "fmt"
  "strings"
  "github.com/pborman/getopt"
  "path/filepath"
  "os"
  "time"
  "net/http"
  "log"
  "io"
  "bytes"
  "mime/multipart"
  // "json"
)

var last_check = time.Now()
var new_last_check = time.Now()
var webhook_url string

type webhook_response struct {
	Test string
}

func main() {
  webhook, path, watch := parse_options()
  webhook_url = webhook

  // wander the path, forever
  for {
    err := filepath.Walk(path, check_file)
    if err != nil { log.Fatal("oh dear") }
    //fmt.Printf("filepath.Walk() returned %v\n", err)
    last_check = new_last_check
    time.Sleep(time.Duration(watch)*time.Second)
  }
}

func parse_options() (webhook_url string, path string, watch int) {

  // Declare the flags to be used
  // helpFlag    := getopt.Bool('h', "display help")
  webhookFlag := getopt.StringLong("webhook",   'w', "", "webhook URL")
  pathFlag    := getopt.StringLong("directory", 'd', "", "directory")
  watchFlag   := getopt.Int16Long("watch", 's', 10, "time between scans")

  getopt.Parse()

  return *webhookFlag, *pathFlag, int(*watchFlag)
}

func check_file(path string, f os.FileInfo, err error) error {
  // fmt.Println("Comparing", f.ModTime(), "to", last_check, "for", path)

  if f.ModTime().After(last_check) && f.Mode().IsRegular() {

    if file_eligible(path) {
      // process file
      process_file(path)
    }

    if new_last_check.Before(f.ModTime()) {
      new_last_check = f.ModTime()
    }
  }

  return nil
}

func file_eligible(file string) (bool) {
  extension := strings.ToLower(filepath.Ext(file))
  if extension == ".png" || extension == ".jpg" || extension == ".gif" {
    return true
  }
  return false
}

func process_file(file string) {
  fmt.Println("Uploading", file)
  // resp, err := http.Post("http://example.com/upload", "image/jpeg", &buf)

  extraParams := map[string]string{
  //  "username":    "Some username",
  }

  request, err := newfileUploadRequest(webhook_url, extraParams, "file", file)
  if err != nil {
    log.Fatal(err)
  }
  client := &http.Client{}
  resp, err := client.Do(request)
  if err != nil {
    log.Fatal(err)
  } else {
    body := &bytes.Buffer{}
    _, err := body.ReadFrom(resp.Body)
    if err != nil {
      log.Fatal(err)
    }
    resp.Body.Close()
    fmt.Println(resp.StatusCode)
    // fmt.Println(resp.Header)
    // fmt.Println(body)

  }
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
  file, err := os.Open(path)
  if err != nil {
      return nil, err
  }
  defer file.Close()

  body := &bytes.Buffer{}
  writer := multipart.NewWriter(body)
  part, err := writer.CreateFormFile(paramName, filepath.Base(path))
  if err != nil {
    return nil, err
  }
  _, err = io.Copy(part, file)

  for key, val := range params {
    _ = writer.WriteField(key, val)
  }
  err = writer.Close()
  if err != nil {
    return nil, err
  }

  req, err := http.NewRequest("POST", uri, body)
  req.Header.Set("Content-Type", writer.FormDataContentType())
  return req, err
}
