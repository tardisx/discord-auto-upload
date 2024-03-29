// Package upload encapsulates prepping an image for sending to discord,
// and actually uploading it there.
package upload

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tardisx/discord-auto-upload/config"
	"github.com/tardisx/discord-auto-upload/image"

	daulog "github.com/tardisx/discord-auto-upload/log"
)

type State string

const (
	StatePending      State = "Pending"          // waiting for decision to upload (could be edited)
	StateQueued       State = "Queued"           // ready for upload
	StateWatermarking State = "Adding Watermark" // thumbnail generation
	StateUploading    State = "Uploading"        // uploading
	StateComplete     State = "Complete"         // finished successfully
	StateFailed       State = "Failed"           // failed
	StateSkipped      State = "Skipped"          // user did not want to upload
)

var currentId int32

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Uploader struct {
	Uploads []*Upload `json:"uploads"`
	Lock    sync.Mutex
}

type Upload struct {
	Id         int32     `json:"id"`
	UploadedAt time.Time `json:"uploaded_at"`

	Image *image.Store

	webhookURL string

	usernameOverride string

	Url string `json:"url"` // url on the discord CDN

	Width  int `json:"width"`
	Height int `json:"height"`

	State       State  `json:"state"`
	StateReason string `json:"state_reason"`

	Client HTTPClient `json:"-"`
}

func NewUploader() *Uploader {
	u := Uploader{}
	uploads := make([]*Upload, 0)
	u.Uploads = uploads
	return &u
}

func (u *Uploader) AddFile(file string, conf config.Watcher) {
	u.Lock.Lock()
	atomic.AddInt32(&currentId, 1)
	thisUpload := Upload{
		Id:               currentId,
		UploadedAt:       time.Time{},
		Image:            &image.Store{OriginalFilename: file, Watermark: !conf.NoWatermark, MaxBytes: 8_000_000},
		webhookURL:       conf.WebHookURL,
		usernameOverride: conf.Username,
		Url:              "",
		State:            StateQueued,
		Client:           nil,
	}
	// if the user wants uploads to be held for editing etc,
	// set it to Pending instead
	if conf.HoldUploads {
		thisUpload.State = StatePending
		thisUpload.StateReason = ""
	}
	u.Uploads = append(u.Uploads, &thisUpload)
	u.Lock.Unlock()

}

// Upload uploads any files that have not yet been uploaded
func (u *Uploader) Upload() {
	u.Lock.Lock()

	for _, upload := range u.Uploads {
		if upload.State == StateQueued {
			upload.processUpload()
		}
	}
	u.Lock.Unlock()

}

func (u *Uploader) UploadById(id int32) *Upload {
	u.Lock.Lock()
	defer u.Lock.Unlock()

	for _, anUpload := range u.Uploads {
		if anUpload.Id == int32(id) {
			return anUpload
		}
	}
	return nil
}

func (u *Upload) processUpload() error {
	daulog.Infof("Uploading: %s", u.Image.OriginalFilename)

	if u.webhookURL == "" {
		daulog.Error("WebHookURL is not configured - cannot upload!")
		return errors.New("webhook url not configured")
	}

	extraParams := map[string]string{}

	if u.usernameOverride != "" {
		daulog.Infof("Overriding username with '%s'", u.usernameOverride)
		extraParams["username"] = u.usernameOverride
	}

	type DiscordAPIResponseAttachment struct {
		URL      string
		ProxyURL string
		Size     int
		Width    int
		Height   int
		Filename string
	}

	type DiscordAPIResponse struct {
		Attachments []DiscordAPIResponseAttachment
		ID          int64 `json:",string"`
	}

	var retriesRemaining = 5
	for retriesRemaining > 0 {

		// open an io.ReadCloser for the file we intend to upload
		var err error

		imageData, err := u.Image.ReadCloser()
		if err != nil {
			panic(err)
		}

		request, err := newfileUploadRequest(u.webhookURL, extraParams, "file", u.Image.UploadFilename(), imageData)
		if err != nil {
			daulog.Errorf("error creating upload request: %s", err)
			return fmt.Errorf("could not create upload request: %s", err)
		}
		start := time.Now()

		if u.Client == nil {
			// if no client was specified (a unit test) then create
			// a default one
			u.Client = &http.Client{Timeout: time.Second * 30}
		}

		resp, err := u.Client.Do(request)
		if err != nil {
			daulog.Errorf("Error performing request: %s", err)
			retriesRemaining--
			sleepForRetries(retriesRemaining)
			continue
		} else {
			if resp.StatusCode == 413 {
				// just fail immediately, we know this means the file was too big
				daulog.Error("413 received - file too large")
				u.State = StateFailed
				u.StateReason = "discord API said file too large"
				return errors.New("received 413 - file too large")
			}

			if resp.StatusCode != 200 {
				// {"message": "Request entity too large", "code": 40005}
				daulog.Errorf("Bad response code from server: %d", resp.StatusCode)
				if b, err := ioutil.ReadAll(resp.Body); err == nil {
					daulog.Errorf("Body:\n%s", string(b))
				}
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}

			resBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				daulog.Errorf("could not deal with body: %s", err)
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			resp.Body.Close()

			var res DiscordAPIResponse
			err = json.Unmarshal(resBody, &res)

			//  {"id": "851092588608880670", "type": 0, "content": "", "channel_id": "849615269706203171", "author": {"bot": true, "id": "849615314274484224", "username": "abcdedf", "avatar": null, "discriminator": "0000"}, "attachments": [{"id": "851092588332449812", "filename": "dau480457962.png", "size": 859505, "url": "https://cdn.discordapp.com/attachments/849615269706203171/851092588332449812/dau480457962.png", "proxy_url": "https://media.discordapp.net/attachments/849615269706203171/851092588332449812/dau480457962.png", "width": 640, "height": 640, "content_type": "image/png"}], "embeds": [], "mentions": [], "mention_roles": [], "pinned": false, "mention_everyone": false, "tts": false, "timestamp": "2021-06-06T13:38:05.660000+00:00", "edited_timestamp": null, "flags": 0, "components": [], "webhook_id": "849615314274484224"}

			daulog.Debugf("Response: %s", string(resBody[:]))

			if err != nil {
				daulog.Errorf("could not parse JSON: %s", err)
				daulog.Errorf("Response was: %s", string(resBody[:]))
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			if len(res.Attachments) < 1 {
				daulog.Error("bad response - no attachments?")
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			var a = res.Attachments[0]
			elapsed := time.Since(start)
			rate := float64(a.Size) / elapsed.Seconds() / 1024.0

			daulog.Infof("Uploaded to %s %dx%d", a.URL, a.Width, a.Height)
			daulog.Infof("id: %d, %d bytes transferred in %.2f seconds (%.2f KiB/s)", res.ID, a.Size, elapsed.Seconds(), rate)

			u.Url = a.URL
			u.State = StateComplete
			u.StateReason = ""
			u.Width = a.Width
			u.Height = a.Height
			u.UploadedAt = time.Now()

			break
		}
	}

	// remove any temporary files
	u.Image.Cleanup()

	if retriesRemaining == 0 {
		daulog.Error("Failed to upload, even after all retries")
		u.State = StateFailed
		u.StateReason = "could not upload after all retries"
		return errors.New("could not upload after all retries")
	}

	return nil
}

func newfileUploadRequest(uri string, params map[string]string, paramName string, filename string, filedata io.Reader) (*http.Request, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, filedata)
	if err != nil {
		log.Fatal("Could not copy: ", err)
	}

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

func sleepForRetries(retry int) {
	if retry == 0 {
		return
	}
	retryTime := (6-retry)*(6-retry) + 6
	daulog.Errorf("Will retry in %d seconds (%d remaining attempts)", retryTime, retry)
	time.Sleep(time.Duration(retryTime) * time.Second)
}
