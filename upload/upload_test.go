package upload

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"testing"
	// 	"github.com/tardisx/discord-auto-upload/config"
)

// https://www.thegreatcodeadventure.com/mocking-http-requests-in-golang/
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func DoGoodUpload(req *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(`{"id": "123456789012345678", "type": 0, "content": "", "channel_id": "849615269706203171", "author": {"bot": true, "id": "849615314274484224", "username": "abcdedf", "avatar": null, "discriminator": "0000"}, "attachments": [{"id": "851092588332449812", "filename": "dau480457962.png", "size": 859505, "url": "https://cdn.discordapp.com/attachments/849615269706203171/851092588332449812/dau480457962.png", "proxy_url": "https://media.discordapp.net/attachments/849615269706203171/851092588332449812/dau480457962.png", "width": 640, "height": 640, "content_type": "image/png"}], "embeds": [], "mentions": [], "mention_roles": [], "pinned": false, "mention_everyone": false, "tts": false, "timestamp": "2021-06-06T13:38:05.660000+00:00", "edited_timestamp": null, "flags": 0, "components": [], "webhook_id": "123456789012345678"}`)))
	return &http.Response{
		StatusCode: 200,
		Body:       r,
	}, nil
}

func DoTooBigUpload(req *http.Request) (*http.Response, error) {
	r := ioutil.NopCloser(bytes.NewReader([]byte(`{"message": "Request entity too large", "code": 40005}`)))
	return &http.Response{
		StatusCode: 413,
		Body:       r,
	}, nil
}

func TestSuccessfulUpload(t *testing.T) {
	// create temporary file, processUpload requires that it exists, even though
	// we will not really be uploading it here
	f, _ := os.CreateTemp("", "dautest-upload-*")
	defer os.Remove(f.Name())
	u := Upload{webhookURL: "https://127.0.0.1/", originalFilename: f.Name()}
	u.Client = &MockClient{DoFunc: DoGoodUpload}
	err := u.processUpload()
	if err != nil {
		t.Errorf("error occured: %s", err)
	}
	if u.Width != 640 || u.Height != 640 {
		t.Error("dimensions wrong")
	}
	if u.Url != "https://cdn.discordapp.com/attachments/849615269706203171/851092588332449812/dau480457962.png" {
		t.Error("URL wrong")
	}
}

func TestTooBigUpload(t *testing.T) {
	// create temporary file, processUpload requires that it exists, even though
	// we will not really be uploading it here
	f, _ := os.CreateTemp("", "dautest-upload-*")
	defer os.Remove(f.Name())
	u := Upload{webhookURL: "https://127.0.0.1/", originalFilename: f.Name()}
	u.Client = &MockClient{DoFunc: DoTooBigUpload}
	err := u.processUpload()
	if err == nil {
		t.Error("error did not occur?")
	} else if err.Error() != "received 413 - file too large" {
		t.Errorf("wrong error occurred: %s", err.Error())
	}
	if u.State != StateFailed {
		t.Error("upload should have been marked failed")
	}
}

func tempImageGt8Mb() {
	// about 12Mb
	width := 2000
	height := 2000

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			color := color.RGBA{uint8(rand.Int31n(256)), uint8(rand.Int31n(256)), uint8(rand.Int31n(256)), 0xff}
			img.Set(x, y, color)
		}
	}

	// Encode as PNG.
	f, _ := os.Create("image.png")
	png.Encode(f, img)
}
