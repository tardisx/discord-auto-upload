package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/tardisx/discord-auto-upload/assets"
//	"strings"
	"regexp"
)

// DAUWebServer - stuff for the web server
type DAUWebServer struct {
	ConfigChange chan int
}

type response struct {
	Success int
}

// I am too noob to work out how to pass context around
var wsConfig DAUWebServer

// r.ParseForm()       // parse arguments, you have to call this by yourself
// fmt.Println(r.Form) // print form information in server side
// fmt.Println("path", r.URL.Path)
// fmt.Println("scheme", r.URL.Scheme)
// fmt.Println(r.Form["url_long"])
// for k, v := range r.Form {
// 	fmt.Println("key:", k)
// 	fmt.Println("val:", strings.Join(v, ""))
// }

func getStatic(w http.ResponseWriter, r *http.Request) {
	// haha this is dumb and I should change it
	fmt.Println(r.URL)
	re := regexp.MustCompile(`[^a-zA-Z0-9\.]`)
	path := r.URL.Path
	sanitized_path := re.ReplaceAll([]byte(path), []byte("_"))

	fmt.Println(sanitized_path)

	data, err := assets.Asset(string(sanitized_path))
	if err != nil {
		// Asset was not found.
		fmt.Fprintln(w, err)
	}
	w.Write(data)
//
}

func getSetWebhook(w http.ResponseWriter, r *http.Request) {
	wsConfig.ConfigChange <- 1
	goodResponse := response{1}
	js, err := json.Marshal(goodResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func getSetDirectory(w http.ResponseWriter, r *http.Request) {
}

// Init is great
func Init() DAUWebServer {
	wsConfig.ConfigChange = make(chan int)
	go startWebServer()
	return wsConfig
}

func startWebServer() {
	http.HandleFunc("/", getStatic)
	http.HandleFunc("/rest/config/webhook", getSetWebhook)
	http.HandleFunc("/rest/config/directory", getSetDirectory)

	log.Print("Starting web server on http://localhost:9090")
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
