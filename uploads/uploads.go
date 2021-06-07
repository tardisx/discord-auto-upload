// The uploads pacakge encapsulates dealing with file uploads to
// discord
package uploads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fogleman/gg"
	"github.com/tardisx/discord-auto-upload/config"
	daulog "github.com/tardisx/discord-auto-upload/log"
	"golang.org/x/image/font/inconsolata"
)

type Upload struct {
	Uploaded         bool      `json:"uploaded"` // has this file been uploaded to discord
	UploadedAt       time.Time `json:"uploaded_at"`
	originalFilename string    // path on the local disk
	mungedFilename   string    // post-watermark
	Url              string    `json:"url"` // url on the discord CDN
	Width            int       `json:"width"`
	Height           int       `json:"height"`
}

var Uploads []*Upload

func AddFile(file string) {
	thisUpload := Upload{
		Uploaded:         false,
		originalFilename: file,
	}
	Uploads = append(Uploads, &thisUpload)

	ProcessUpload(&thisUpload)
}

func ProcessUpload(up *Upload) {

	file := up.originalFilename

	if !config.Config.NoWatermark {
		daulog.SendLog("Copying to temp location and watermarking ", daulog.LogTypeInfo)
		file = mungeFile(file)
		up.mungedFilename = file
	}

	if config.Config.WebHookURL == "" {
		daulog.SendLog("WebHookURL is not configured - cannot upload!", daulog.LogTypeError)
		return
	}

	extraParams := map[string]string{}

	if config.Config.Username != "" {
		daulog.SendLog("Overriding username with "+config.Config.Username, daulog.LogTypeInfo)
		extraParams["username"] = config.Config.Username
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

		request, err := newfileUploadRequest(config.Config.WebHookURL, extraParams, "file", file)
		if err != nil {
			log.Fatal(err)
		}
		start := time.Now()
		client := &http.Client{Timeout: time.Second * 30}
		resp, err := client.Do(request)
		if err != nil {
			log.Print("Error performing request:", err)
			retriesRemaining--
			sleepForRetries(retriesRemaining)
			continue
		} else {

			if resp.StatusCode != 200 {
				// {"message": "Request entity too large", "code": 40005}
				log.Print("Bad response from server:", resp.StatusCode)
				if b, err := ioutil.ReadAll(resp.Body); err == nil {
					log.Print("Body:", string(b))
					daulog.SendLog(fmt.Sprintf("Bad response: %s", string(b)), daulog.LogTypeError)
				}
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}

			resBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Print("could not deal with body: ", err)
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			resp.Body.Close()

			var res DiscordAPIResponse
			err = json.Unmarshal(resBody, &res)

			//  {"id": "851092588608880670", "type": 0, "content": "", "channel_id": "849615269706203171", "author": {"bot": true, "id": "849615314274484224", "username": "abcdedf", "avatar": null, "discriminator": "0000"}, "attachments": [{"id": "851092588332449812", "filename": "dau480457962.png", "size": 859505, "url": "https://cdn.discordapp.com/attachments/849615269706203171/851092588332449812/dau480457962.png", "proxy_url": "https://media.discordapp.net/attachments/849615269706203171/851092588332449812/dau480457962.png", "width": 640, "height": 640, "content_type": "image/png"}], "embeds": [], "mentions": [], "mention_roles": [], "pinned": false, "mention_everyone": false, "tts": false, "timestamp": "2021-06-06T13:38:05.660000+00:00", "edited_timestamp": null, "flags": 0, "components": [], "webhook_id": "849615314274484224"}

			daulog.SendLog(fmt.Sprintf("Response: %s", string(resBody[:])), daulog.LogTypeDebug)

			if err != nil {
				log.Print("could not parse JSON: ", err)
				fmt.Println("Response was:", string(resBody[:]))
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			if len(res.Attachments) < 1 {
				log.Print("bad response - no attachments?")
				retriesRemaining--
				sleepForRetries(retriesRemaining)
				continue
			}
			var a = res.Attachments[0]
			elapsed := time.Since(start)
			rate := float64(a.Size) / elapsed.Seconds() / 1024.0

			daulog.SendLog(fmt.Sprintf("Uploaded to %s %dx%d", a.URL, a.Width, a.Height), daulog.LogTypeInfo)
			daulog.SendLog(fmt.Sprintf("id: %d, %d bytes transferred in %.2f seconds (%.2f KiB/s)", res.ID, a.Size, elapsed.Seconds(), rate), daulog.LogTypeInfo)

			up.Url = a.URL
			up.Uploaded = true
			up.Width = a.Width
			up.Height = a.Height
			up.UploadedAt = time.Now()

			break
		}
	}

	if !config.Config.NoWatermark {
		daulog.SendLog(fmt.Sprintf("Removing temporary file: %s", file), daulog.LogTypeDebug)
		os.Remove(file)
	}

	if retriesRemaining == 0 {
		log.Fatal("Failed to upload, even after retries")
	}
}

func newfileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
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

func mungeFile(path string) string {

	reader, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	im, _, err := image.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	bounds := im.Bounds()
	// var S float64 = float64(bounds.Max.X)

	dc := gg.NewContext(bounds.Max.X, bounds.Max.Y)
	dc.Clear()
	dc.SetRGB(0, 0, 0)

	dc.SetFontFace(inconsolata.Regular8x16)

	dc.DrawImage(im, 0, 0)

	dc.DrawRoundedRectangle(0, float64(bounds.Max.Y-18.0), 320, float64(bounds.Max.Y), 0)
	dc.SetRGB(0, 0, 0)
	dc.Fill()

	dc.SetRGB(1, 1, 1)

	dc.DrawString("github.com/tardisx/discord-auto-upload", 5.0, float64(bounds.Max.Y)-5.0)

	tempfile, err := ioutil.TempFile("", "dau")
	if err != nil {
		log.Fatal(err)
	}
	tempfile.Close()
	os.Remove(tempfile.Name())
	actualName := tempfile.Name() + ".png"

	dc.SavePNG(actualName)
	return actualName
}

func sleepForRetries(retry int) {
	if retry == 0 {
		return
	}
	retryTime := (6-retry)*(6-retry) + 6
	daulog.SendLog(fmt.Sprintf("Will retry in %d seconds (%d remaining attempts)", retryTime, retry), daulog.LogTypeError)
	time.Sleep(time.Duration(retryTime) * time.Second)
}
