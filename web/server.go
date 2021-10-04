package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/tardisx/discord-auto-upload/config"
	daulog "github.com/tardisx/discord-auto-upload/log"
	"github.com/tardisx/discord-auto-upload/uploads"
	"github.com/tardisx/discord-auto-upload/version"
)

//go:embed data
var webFS embed.FS

// DAUWebServer - stuff for the web server
type DAUWebServer struct {
	ConfigChange chan int
}

type valueStringResponse struct {
	Success bool   `json:"success"`
	Value   string `json:"value"`
}

type valueBooleanResponse struct {
	Success bool `json:"success"`
	Value   bool `json:"value"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func getStatic(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.TrimLeft(path, "/")
	if path == "" {
		path = "index.html"
	}

	extension := filepath.Ext(string(path))

	if extension == ".html" {

		t, err := template.ParseFS(webFS, "data/wrapper.tmpl", "data/"+path)
		if err != nil {
			panic(err)
		}

		log.Printf("req: %s", r.URL.Path)

		var b struct {
			Body    string
			Path    string
			Version string
		}
		b.Path = path
		b.Version = version.CurrentVersion

		err = t.ExecuteTemplate(w, "layout", b)
		if err != nil {
			panic(err)
		}
		return
	} else {
		otherStatic, err := webFS.ReadFile("data/" + path)

		if err != nil {
			log.Fatalf("problem with '%s': %v", path, err)
		}
		w.Header().Set("Content-Type", mime.TypeByExtension(extension))

		w.Write(otherStatic)
		return
	}

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

func getSetExclude(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "GET" {
		getResponse := valueStringResponse{Success: true, Value: config.Config.Exclude}
		// I can't see any way this will fail
		js, _ := json.Marshal(getResponse)
		w.Write(js)
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}
		config.Config.Exclude = r.PostForm.Get("value")
		config.SaveConfig()

		postResponse := valueStringResponse{Success: true, Value: config.Config.Exclude}

		js, _ := json.Marshal(postResponse)
		w.Write(js)
	}
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	showDebug := false
	debug, present := r.URL.Query()["debug"]
	if present && len(debug[0]) > 0 && debug[0] != "0" {
		showDebug = true
	}

	text := ""
	for _, log := range daulog.LogEntries {
		if !showDebug && log.Type == daulog.LogTypeDebug {
			continue
		}
		text = text + fmt.Sprintf(
			"%-6s %-19s %s\n", log.Type, log.Timestamp.Format("2006-01-02 15:04:05"), log.Entry,
		)
	}

	//	js, _ := json.Marshal(daulog.LogEntries)
	w.Write([]byte(text))
}

func getUploads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ups := uploads.Uploads
	text, _ := json.Marshal(ups)
	w.Write([]byte(text))
}

func StartWebServer() {

	http.HandleFunc("/", getStatic)
	http.HandleFunc("/rest/config/webhook", getSetWebhook)
	http.HandleFunc("/rest/config/username", getSetUsername)
	http.HandleFunc("/rest/config/watch", getSetWatch)
	http.HandleFunc("/rest/config/nowatermark", getSetNoWatermark)
	http.HandleFunc("/rest/config/directory", getSetDirectory)
	http.HandleFunc("/rest/config/exclude", getSetExclude)

	http.HandleFunc("/rest/logs", getLogs)
	http.HandleFunc("/rest/uploads", getUploads)

	go func() {
		log.Print("Starting web server on http://localhost:9090")
		err := http.ListenAndServe(":9090", nil) // set listen port
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}
