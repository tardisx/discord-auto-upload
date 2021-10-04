package web

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHome(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	getStatic(w, req)
	res := w.Result()
	defer res.Body.Close()
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

	notFounds := []string{
		"/abc.html", "/foo.html", "/foo.html", "/../foo.html",
		"/foo.gif",
	}

	for _, nf := range notFounds {

		req := httptest.NewRequest(http.MethodGet, nf, nil)
		w := httptest.NewRecorder()
		getStatic(w, req)
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
