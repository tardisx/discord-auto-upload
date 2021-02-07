package web

import (
	"encoding/json"
	"fmt"
	"github.com/tardisx/discord-auto-upload/assets"
	"github.com/tardisx/discord-auto-upload/config"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"
)

// DAUWebServer - stuff for the web server
type DAUWebServer struct {
	ConfigChange chan int
}

type valueStringResponse struct {
	Success bool   `json: 'success'`
	Value   string `json: 'value'`
}

type valueBooleanResponse struct {
	Success bool `json: 'success'`
	Value   bool `json: 'value'`
}

type errorResponse struct {
	Success bool   `json: 'success'`
	Error   string `json: 'error'`
}

func getStatic(w http.ResponseWriter, r *http.Request) {
	// haha this is dumb and I should change it
	// fmt.Println(r.URL)
	re := regexp.MustCompile(`[^a-zA-Z0-9\.]`)
	path := r.URL.Path[1:]
	sanitized_path := re.ReplaceAll([]byte(path), []byte("_"))

	if string(sanitized_path) == "" {
		sanitized_path = []byte("index.html")
	}

	data, err := assets.Asset(string(sanitized_path))
	if err != nil {
		// Asset was not found.
		fmt.Fprintln(w, err)
	}

	extension := filepath.Ext(string(sanitized_path))

	// is this a HTML file? if so wrap it in the template
	if extension == ".html" {
		wrapper, _ := assets.Asset("wrapper.tmpl")
		t := template.Must(template.New("wrapper").Parse(string(wrapper)))
		var b struct {
			Body    string
			Path    string
			Version string
		}
		b.Body = string(data)
		b.Path = string(sanitized_path)
		b.Version = config.CurrentVersion
		t.Execute(w, b)
		return
	}

	// otherwise we are a static thing
	w.Header().Set("Content-Type", mime.TypeByExtension(extension))

	w.Write(data)
	//
}

// TODO there should be locks around all these config accesses
func getSetWebhook(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		getResponse := valueStringResponse{Success: true, Value: config.Config.WebHookURL}

		// I can't see any way this will fail
		js, _ := json.Marshal(getResponse)
		w.Write(js)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}
		config.Config.WebHookURL = r.PostForm.Get("value")
		config.SaveConfig()
		postResponse := valueStringResponse{Success: true, Value: config.Config.WebHookURL}

		js, _ := json.Marshal(postResponse)
		w.Write(js)
	}
}

// TODO there should be locks around all these config accesses
func getSetUsername(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		getResponse := valueStringResponse{Success: true, Value: config.Config.Username}

		// I can't see any way this will fail
		js, _ := json.Marshal(getResponse)
		w.Write(js)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}
		config.Config.Username = r.PostForm.Get("value")
		config.SaveConfig()

		postResponse := valueStringResponse{Success: true, Value: config.Config.Username}

		js, _ := json.Marshal(postResponse)
		w.Write(js)
	}
}

func getSetWatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		getResponse := valueStringResponse{Success: true, Value: strconv.Itoa(config.Config.Watch)}

		// I can't see any way this will fail
		js, _ := json.Marshal(getResponse)
		w.Write(js)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}

		i, err := strconv.Atoi(r.PostForm.Get("value"))

		if err != nil {
			response := errorResponse{Success: false, Error: fmt.Sprintf("Bad value for watch: %v", err)}
			js, _ := json.Marshal(response)
			w.Write(js)
			return
		}

		if i < 1 {
			response := errorResponse{Success: false, Error: "must be > 0"}
			js, _ := json.Marshal(response)
			w.Write(js)
			return
		}

		config.Config.Watch = i
		config.SaveConfig()

		postResponse := valueStringResponse{Success: true, Value: strconv.Itoa(config.Config.Watch)}

		js, _ := json.Marshal(postResponse)
		w.Write(js)
	}
}

func getSetNoWatermark(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		getResponse := valueBooleanResponse{Success: true, Value: config.Config.NoWatermark}

		// I can't see any way this will fail
		js, _ := json.Marshal(getResponse)
		w.Write(js)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}

		v := r.PostForm.Get("value")

		if v != "0" && v != "1" {
			response := errorResponse{Success: false, Error: fmt.Sprintf("Bad value for nowatermark: %v", err)}
			js, _ := json.Marshal(response)
			w.Write(js)
			return
		}

		if v == "0" {
			config.Config.NoWatermark = false
		} else {
			config.Config.NoWatermark = true
		}
		config.SaveConfig()

		postResponse := valueBooleanResponse{Success: true, Value: config.Config.NoWatermark}

		js, _ := json.Marshal(postResponse)
		w.Write(js)
	}
}

func getSetDirectory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		getResponse := valueStringResponse{Success: true, Value: config.Config.Path}

		// I can't see any way this will fail
		js, _ := json.Marshal(getResponse)
		w.Write(js)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}
		newPath := r.PostForm.Get("value")

		// sanity check this path
		stat, err := os.Stat(newPath)
		if os.IsNotExist(err) {
			// not exist
			response := errorResponse{Success: false, Error: fmt.Sprintf("Path: %s - does not exist", newPath)}
			js, _ := json.Marshal(response)
			w.Write(js)
			return
		} else if !stat.IsDir() {
			// not a directory
			response := errorResponse{Success: false, Error: fmt.Sprintf("Path: %s - is not a directory", newPath)}
			js, _ := json.Marshal(response)
			w.Write(js)
			return
		}

		config.Config.Path = newPath
		config.SaveConfig()

		postResponse := valueStringResponse{Success: true, Value: config.Config.Path}

		js, _ := json.Marshal(postResponse)
		w.Write(js)
	}
}

func StartWebServer() {
	http.HandleFunc("/", getStatic)
	http.HandleFunc("/rest/config/webhook", getSetWebhook)
	http.HandleFunc("/rest/config/username", getSetUsername)
	http.HandleFunc("/rest/config/watch", getSetWatch)
	http.HandleFunc("/rest/config/nowatermark", getSetNoWatermark)
	http.HandleFunc("/rest/config/directory", getSetDirectory)

	log.Print("Starting web server on http://localhost:9090")
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
