# Automatically upload screenshots into a discord channel

[![Go](https://github.com/tardisx/discord-auto-upload/actions/workflows/go.yml/badge.svg)](https://github.com/tardisx/discord-auto-upload/actions/workflows/go.yml)

This program automatically uploads new screenshots that appear in a folder on your computer to Discord and posts them in a channel:

![Screenshot](http://i.imgur.com/QPS9V6f.jpg)

Point it at your Steam screenshot folder, or similar, and shortly after you hit your screenshot hotkey the screenshot will appear in your discord chat.

Need help? Join our discord: https://discord.gg/eErG9sntbZ

## What you'll need

* A folder where screenshots are stored
* A [discord webhook](https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
* This program

## Getting started

### Binaries

Binaries are available for Mac, Linux and Windows [here](https://github.com/tardisx/discord-auto-upload/releases/latest).

#### From source

You'll need to [download Go](https://golang.org/dl/), check the code out somewhere, run 'go generate' and then 'go build'.

## Using it

`dau` configuration is managed via its internal web interface. When the executable is run, you can visit
`http://localhost:9090` in your web browser to configure the it. Configuration persists across runs, it is
saved in a file called '.dau.json' in your home directory.

The first time you run it, you will need to configure at least the discord web hook and the watch path for
`dau` to be useful.

While running, `dau` will continually scan a directory for new images, and each time it finds one it will upload it to discord, via the discord web hook.

`dau` will only upload "new" screenshots, where "new" means a file that appears in a directory that it is watching, if it appears *after* it has started executing.

Thus, you do not have to worry about pointing `dau` at a directory full of images, it will only upload new ones.

## Configuration options

See the web interface at http://localhost:9090 to configure `dau`. The configuration is a single page of options,
no changes will take effect until the "Save All Configuration" button has been pressed.

### Global options

* Server port - the port number the web server listens on. Requires restart
* Watch interval - how often each watcher will check the directory for new files

### Watcher configuration

There can be one or more watchers configured. Each watcher looks in a particular directory,
and uploads new files to a different discord channel.

Each watcher has the following configuration options:

* Directory to watch - This is the path that `dau` will periodically inspect, looking for new images.
Note that subdirectories are also scanned. You need to enter the full filesystem path here.
* Discord WebHook URL - The webhook URL from Discord. See https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks for more information on setting one up.
* Username - This is completely optional and can be any arbitrary string. It makes the upload
appear to come from a different user (though this is visual only, and does not
actually hide the bot identity in any way). You might like to set it to your own
discord name.
* Watermark - Disabling the watermark will prevent `dau` from putting a link to the projects
on the bottom left hand corner of your uploaded images. I really appreciate it when you leave this enabled :-)
* Hold Uploads - See "Holding uploads" below
* Exclusions - You can set one or more arbitrary strings to exclude files from being matched by this watcher.
This is most commonly used to prevent thumbnail images from being uploads.

## Holding uploads

If the "Hold Uploads" option is selected, newly found files will not immediately be uploaded. They will be available
in the "uploads" tab of the web interface. This has two purposes:

* It gives you a chance to vet your screenshot selection before uploading
* It allows you to edit the images before uploading.

In the list of uploads there are three actions you can take on each file:

* Press "upload" to upload the image
* Press "reject" to reject the image
* Click on the image thumbnail to edit the image

If you click on the image thumbnail, an image editor will open, and allow you to add text captions to your image.
More functionality is coming soon. When you are finished editing, choose "Apply" and you will return to the uploads
list. Click "upload" to upload your edited image.

## Limitations/bugs

* Only files ending jpg, gif or png are uploaded.
* If multiple screenshots occur quickly (<1 second apart) not all may be uploaded.
* Files to upload are determined by the file modification time. If you drag and drop existing files they will
  not be detected and uploaded. Only newly created files will be detected.

## Troubleshooting

Please check the "log" page on the web interface for information when things are
not working as you expect.

## TODO

Open an [issue](https://github.com/tardisx/discord-auto-upload/issues/new) and let me know what you'd like to see.

Please include any relevant logs from the console when reporting bugs.
