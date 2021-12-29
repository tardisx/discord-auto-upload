package web

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/tardisx/discord-auto-upload/config"
	"github.com/tardisx/discord-auto-upload/imageprocess"
	daulog "github.com/tardisx/discord-auto-upload/log"
	"github.com/tardisx/discord-auto-upload/upload"
	"github.com/tardisx/discord-auto-upload/version"
)

type WebService struct {
	Config   *config.ConfigService
	Uploader *upload.Uploader
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StartUploadRequest struct {
	Id int32 `json:"id"`
}

type StartUploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
			daulog.SendLog(fmt.Sprintf("when fetching: %s got: %s", path, err), daulog.LogTypeError)
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
			return
		}

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
			daulog.SendLog(fmt.Sprintf("when fetching: %s got: %s", path, err), daulog.LogTypeError)
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
			returnJSONError(w, "could not read body?")
			return
		}
		err = json.Unmarshal(b, &newConfig)
		if err != nil {
			returnJSONError(w, "badly formed JSON")
			return
		}
		ws.Config.Config = &newConfig
		err = ws.Config.Save()
		if err != nil {
			returnJSONError(w, err.Error())
			return
		}
		// config has changed, so tell the world
		if ws.Config.Changed != nil {
			ws.Config.Changed <- true
		}

	}

	b, _ := json.Marshal(ws.Config.Config)
	w.Write(b)
}

func (ws *WebService) getUploads(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ups := ws.Uploader.Uploads

	text, err := json.Marshal(ups)
	if err != nil {
		// not sure how this would happen, so we probably want to find out the hard way
		panic(err)
	}
	w.Write([]byte(text))
}

func (ws *WebService) imageThumb(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "image/png")
	processor := imageprocess.Processor{}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		returnJSONError(w, "bad id")
		return
	}

	ul := ws.Uploader.UploadById(int32(id))
	if ul == nil {
		returnJSONError(w, "bad id")
		return
	}
	err = processor.ThumbPNG(ul, "orig", w)
	if err != nil {
		returnJSONError(w, "could not create thumb")
		return
	}
}

func (ws *WebService) imageMarkedupThumb(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "image/png")
	processor := imageprocess.Processor{}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		returnJSONError(w, "bad id")
		return
	}

	ul := ws.Uploader.UploadById(int32(id))
	if ul == nil {
		returnJSONError(w, "bad id")
		return
	}
	err = processor.ThumbPNG(ul, "markedup", w)
	if err != nil {
		returnJSONError(w, "could not create thumb")
		return
	}
}

func (ws *WebService) image(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		returnJSONError(w, "bad id")
		return
	}

	ul := ws.Uploader.UploadById(int32(id))
	if ul == nil {
		returnJSONError(w, "bad id")
		return
	}

	img, err := os.Open(ul.OriginalFilename)
	if err != nil {
		returnJSONError(w, "could not open image file")
		return
	}
	defer img.Close()
	io.Copy(w, img)
}

func (ws *WebService) modifyUpload(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.Method == "POST" {

		vars := mux.Vars(r)
		change := vars["change"]
		id, err := strconv.ParseInt(vars["id"], 10, 32)
		if err != nil {
			returnJSONError(w, "bad id")
			return
		}

		anUpload := ws.Uploader.UploadById(int32(id))
		if anUpload == nil {
			returnJSONError(w, "bad id")
			return
		}

		if anUpload.State == upload.StatePending {
			if change == "start" {
				anUpload.State = upload.StateQueued
				res := StartUploadResponse{Success: true, Message: "upload queued"}
				resString, _ := json.Marshal(res)
				w.Write(resString)
				return
			} else if change == "skip" {
				anUpload.State = upload.StateSkipped
				res := StartUploadResponse{Success: true, Message: "upload skipped"}
				resString, _ := json.Marshal(res)
				w.Write(resString)
				return
			} else if change == "markup" {
				newImageData := r.FormValue("image")
				//data:image/png;base64,xxxx
				// I know this is dumb, we should just send binary image data, but I can't
				// see that Fabric makes that possible.
				if strings.Index(newImageData, "data:image/png;base64,") != 0 {
					returnJSONError(w, "bad image data")
					return
				}
				imageDataBase64 := newImageData[22:]
				b, err := base64.StdEncoding.DecodeString(imageDataBase64)
				if err != nil {
					returnJSONError(w, err.Error())
					return
				}

				// write to a temporary file
				tempfile, err := ioutil.TempFile("", "dau_markup")
				if err != nil {
					log.Fatal(err)
				}
				n, err := tempfile.Write(b)
				if n != len(b) {
					log.Fatalf("only wrote %d bytes??", n)
				}
				if err != nil {
					log.Fatalf("Could not write temp file: %v", err)
				}

				tempfile.Close()
				anUpload.MarkedUpFilename = tempfile.Name()

			} else {
				returnJSONError(w, "bad change type")
				return
			}
		}
		res := StartUploadResponse{Success: false, Message: "upload does not exist, or already queued"}
		resString, _ := json.Marshal(res)
		w.WriteHeader(400)
		w.Write(resString)
		return
	}
	returnJSONError(w, "bad request")

}

func (ws *WebService) StartWebServer() {

	r := mux.NewRouter()

	r.HandleFunc("/rest/logs", ws.getLogs)
	r.HandleFunc("/rest/uploads", ws.getUploads)
	r.HandleFunc("/rest/upload/{id:[0-9]+}/{change}", ws.modifyUpload)

	r.HandleFunc("/rest/image/{id:[0-9]+}/thumb", ws.imageThumb)
	r.HandleFunc("/rest/image/{id:[0-9]+}/markedup_thumb", ws.imageMarkedupThumb)

	r.HandleFunc("/rest/image/{id:[0-9]+}", ws.image)

	r.HandleFunc("/rest/config", ws.handleConfig)
	r.PathPrefix("/").HandlerFunc(ws.getStatic)

	go func() {
		listen := fmt.Sprintf(":%d", ws.Config.Config.Port)
		log.Printf("Starting web server on http://localhost%s", listen)

		srv := &http.Server{
			Handler: r,
			Addr:    listen,
			// Good practice: enforce timeouts for servers you create!
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}

		log.Fatal(srv.ListenAndServe())

	}()
}

func returnJSONError(w http.ResponseWriter, errMessage string) {
	w.WriteHeader(400)
	errJSON := ErrorResponse{
		Error: errMessage,
	}
	errString, _ := json.Marshal(errJSON)
	w.Write(errString)
}
