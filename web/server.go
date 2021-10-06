package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/tardisx/discord-auto-upload/config"
	daulog "github.com/tardisx/discord-auto-upload/log"
	"github.com/tardisx/discord-auto-upload/version"
)

type WebService struct {
	Config config.ConfigService
}

//go:embed data
var webFS embed.FS

// DAUWebServer - stuff for the web server
type DAUWebServer struct {
	//	ConfigChange chan int

}

func (ws *WebService) getStatic(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	path = strings.TrimLeft(path, "/")
	if path == "" {
		path = "index.html"
	}

	extension := filepath.Ext(string(path))

	if extension == ".html" { // html file
		t, err := template.ParseFS(webFS, "data/wrapper.tmpl", "data/"+path)
		if err != nil {
			log.Printf("when fetching: %s got: %s", path, err)
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return
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
	} else { // anything else
		otherStatic, err := webFS.ReadFile("data/" + path)

		if err != nil {
			log.Printf("when fetching: %s got: %s", path, err)
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return
		}
		w.Header().Set("Content-Type", mime.TypeByExtension(extension))

		w.Write(otherStatic)
		return
	}

}

func (ws *WebService) getLogs(w http.ResponseWriter, r *http.Request) {
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

func (ws *WebService) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		newConfig := config.ConfigV2{}

		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.Write([]byte("bad body"))
			return
		}
		err = json.Unmarshal(b, &newConfig)
		if err != nil {
			w.Write([]byte("bad data"))
			log.Printf("%s", err)
			return
		}
		ws.Config.Config = newConfig
		ws.Config.Save()

	}

	b, _ := json.Marshal(ws.Config.Config)
	w.Write(b)
}

// func getUploads(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	ups := uploads.Uploads
// 	text, _ := json.Marshal(ups)
// 	w.Write([]byte(text))
// }

func (ws *WebService) StartWebServer() {

	http.HandleFunc("/", ws.getStatic)

	http.HandleFunc("/rest/logs", ws.getLogs)
	// http.HandleFunc("/rest/uploads", getUploads)
	http.HandleFunc("/rest/config", ws.handleConfig)

	go func() {
		listen := fmt.Sprintf(":%d", ws.Config.Config.Port)
		log.Print("Starting web server on http://localhost%s", listen)
		err := http.ListenAndServe(listen, nil) // set listen port
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}
