package web

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tardisx/discord-auto-upload/config"
)

func TestHome(t *testing.T) {
	s := WebService{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	s.getStatic(w, req)
	res := w.Result()

	data, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if !strings.Contains(string(data), "DAU") {
		t.Errorf("does not look like correct homepage at /")
	}
	if res.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("wrong content type for / - %s", res.Header.Get("Content-Type"))
	}

}

func TestNotFound(t *testing.T) {
	s := WebService{}

	notFounds := []string{
		"/abc.html", "/foo.html", "/foo.html", "/../foo.html",
		"/foo.gif",
	}

	for _, nf := range notFounds {

		req := httptest.NewRequest(http.MethodGet, nf, nil)
		w := httptest.NewRecorder()
		s.getStatic(w, req)
		res := w.Result()

		defer res.Body.Close()
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("expected error to be nil got %v", err)
		}
		if string(b) != "not found" {
			t.Errorf("expected body to be not found, not '%s'", string(b))
		}
		if res.Header.Get("Content-Type") != "text/plain" {
			t.Error("Wrong content type for not found")

		}
	}
}

func TestGetConfig(t *testing.T) {
	conf := config.DefaultConfigService()
	conf.Config = config.DefaultConfig()
	s := WebService{Config: conf}

	req := httptest.NewRequest(http.MethodGet, "/rest/config", nil)
	w := httptest.NewRecorder()
	s.handleConfig(w, req)
	res := w.Result()
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)

	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	exp := `{"WatchInterval":10,"Version":2,"Port":9090,"Watchers":[{"WebHookURL":"https://webhook.url.here","Path":"/your/screenshot/dir/here","Username":"","NoWatermark":false,"HoldUploads":false,"Exclude":[]}]}`
	if string(b) != exp {
		t.Errorf("Got unexpected response\n%v\n%v", string(b), exp)
	}
}
