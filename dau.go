package main

import (
  "fmt"
  "strings"
  "github.com/pborman/getopt"
  "path/filepath"
  "os"
  "time"
  "net/http"
)

var last_check = time.Now()
var new_last_check = time.Now()

func main() {
  webhook_url, path := parse_options()
  fmt.Println(webhook_url, path)

  // wander the path, forever
  for {
    err := filepath.Walk(path, check_file)
    if err != nil { fmt.Printf("SHIT ERROR") }
    //fmt.Printf("filepath.Walk() returned %v\n", err)
    last_check = new_last_check
    time.Sleep(500 * time.Millisecond)
  }
}


func parse_options() (webhook_url string, path string) {

  // Declare the flags to be used
  // helpFlag    := getopt.Bool('h', "display help")
  webhookFlag := getopt.StringLong("webhook",   'w', "", "webhook URL")
  pathFlag    := getopt.StringLong("directory", 'd', "", "directory")

  getopt.Parse()

  return *webhookFlag, *pathFlag

}

func check_file(path string, f os.FileInfo, err error) error {
  // fmt.Println("Comparing", f.ModTime(), "to", last_check, "for", path)

  if f.ModTime().After(last_check) && f.Mode().IsRegular() {
    fmt.Println("YES", path, "is new")

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
}
