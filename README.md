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

Put them somewhere on your path and run from the command line.

The windows version comes with a .bat file to make this a little easier - edit the `dau.bat` file to include your webhook URL and
other parameters, then you can simply double click `dau.bat` to start `dau` running.

#### From source

You'll need to [download Go](https://golang.org/dl/) check the code out somewhere, and 'go build'.

## Using it

`dau` is a command line driven program. When executed, it will continually scan a directory for new images, and each time it finds one it will upload it to discord, via the discord web hook.

`dau` will only upload "new" screenshots, where "new" means a file that appears in a directory that it is watching, if it appears *after* it has started executing.

Thus, you do not have to worry about pointing `dau` at a directory full of images, it will only upload new ones.

If `dau` is on your path, you can run it from your screenshot folder and there is then no need to specify the path to your images.

The only two mandatory command line parameters are the discord webhook URL:

`--webhook URL` - the webhook URL (see [here](https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks) for details).

and the directory to watch:

`--directory /some/path/here` - the directory that screenshots will appear in.

You will have to quote the path on windows, or anywhere where the directory path contains spaces. Note that
subdirectories will also be scanned.

Other parameters are:

`--exclude <string>` - exclude any files that contain this string (commonly used to avoid uploading thumbnails).

`--watch xx` - specify how many seconds to wait between scanning the directory. The default is 10 seconds.

`--username <username>` - an arbitrary string to show as the bot's username in the channel.

`--no-watermark` - don't watermark images with a reference to this tool.

`--help` - show command line help.

`--version` - show the version.

## Limitations/bugs

* Only files ending jpg, gif or png are uploaded.
* If multiple screenshots occur quickly (<1 second apart) not all may be uploaded.
* Files to upload are determined by the file modification time. If you drag and drop existing files they will not be detected and uploaded. Only newly created files will be detected.

## TODO
This is just a quick hack. Open to suggestions on new features and improvements.

Open an [issue](https://github.com/tardisx/discord-auto-upload/issues/new) and let me know.
