# Automatically upload screenshots into a discord channel

This program automatically uploads new screenshots that appear in a folder on your computer to Discord and posts them in a channel:

![Screenshot](http://i.imgur.com/QPS9V6f.jpg)

Point it at your Steam screenshot folder, or similar, and shortly after you hit your screenshot hotkey the screenshot will appear in your discord chat.

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

See the web interface at http://localhost:9090 to configure `dau`.

### 'Discord WebHook URL'

The webhook URL from Discord. See https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks
for more information on setting one up.

### 'Bot Username'

This is completely optional and can be any arbitrary string. It makes the upload
appear to come from a different user (though this is visual only, and does not
actually hide the bot identity in any way). You might like to set it to your own
discord name.

### 'Directory to watch'

This is the path that `dau` will periodically inspect, looking for new images.
Note that subdirectories are also scanned. You need to enter the full filesystem
path here.

### 'Period between filesystem checks'

This is the number of seconds between which `dau` will look for new images.

### 'Do not watermark images'

This will disable the watermarking of images. I like it when you don't set this :-)

### 'Files to exclude'

This is a string to match against the filename to check for exclusions. The common
use case is to use 'thumbnail' or similar if your image directory contains additional
thumbnail files.

## Limitations/bugs

* Only files ending jpg, gif or png are uploaded.
* If multiple screenshots occur quickly (<1 second apart) not all may be uploaded.
* Files to upload are determined by the file modification time. If you drag and drop existing files they will
  not be detected and uploaded. Only newly created files will be detected.

## Troubleshooting

Please check the "log" page on the web interface for information when things are
not working as you expect.

## TODO
This is just a relatively quick hack. Open to suggestions on new features and improvements.

Open an [issue](https://github.com/tardisx/discord-auto-upload/issues/new) and let me know.
Please include any relevant logs from the console when reporting bugs.
