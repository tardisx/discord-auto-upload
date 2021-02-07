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

You'll need to [download Go](https://golang.org/dl/) check the code out somewhere, run 'go generate' and then 'go build'.

## Using it

`dau` configuration is managed via its internal web interface. When the executable is run, you can visit
`http://localhost:9090` in your web browser to configure the it. Configuration persists across runs, it is
saved in a file called '.dau.json' in your home directory.

The first time you run it, you will need to configure at least the discord web hook and the watch path for
`dau` to be useful.

While running, `dau` will continually scan a directory for new images, and each time it finds one it will
upload it to discord, via the discord web hook.

`dau` will only upload "new" screenshots, where "new" means a file that appears in a directory that it is watching, if it appears *after* it has started executing.

Thus, you do not have to worry about pointing `dau` at a directory full of images, it will only upload new ones.

## Limitations/bugs

* Only files ending jpg, gif or png are uploaded.
* If multiple screenshots occur quickly (<1 second apart) not all may be uploaded.
* Files to upload are determined by the file modification time. If you drag and drop existing files they will
  not be detected and uploaded. Only newly created files will be detected.

## TODO
This is just a relatively quick hack. Open to suggestions on new features and improvements.

Open an [issue](https://github.com/tardisx/discord-auto-upload/issues/new) and let me know.
Please include any relevant logs from the console when reporting bugs.
