# Automatically upload screenshots from your computer into a discord channel

This script automaticall uploads new screenshots that appear in a folder on your computer to Discord and posts them in a channel:

![Screenshot](http://i.imgur.com/QPS9V6f.jpg)

Point it at your Steam screenshot folder, or similar, and shortly after you hit your screenshot hotkey the screenshot will appear in your discord chat.

## What you'll need

* A folder where screenshots are stored
* A [discord webhook](https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
* This script
* perl installed, or one of the provided binaries

## Getting started

### Linux

#### Standalone binary

* Grab the latest binary from `https://github.com/tardisx/discord-auto-upload/releases/latest` called 'dau-linux.gz'.
* `gunzip dau-linux.gz`
* `mv dau-linux dau`
* `chmod +x dau`
* put `dau` somewhere on your path

#### From Source

* Download the script:

`curl -O https://raw.githubusercontent.com/tardisx/discord-auto-upload/master/dau`

* Put it somewhere on your path (if you want to be able to run it from anywhere)
* chmod +x it
* Install the dependencies:

CPAN: `cpan install Mojolicious IO::Socket::SSL`

CPANM: `cpanm Mojolicious IO::Socket::SSL`

Ubuntu/Debian: `sudo apt-get install libmojolicious-perl libio-socket-ssl-perl`

* test it:

`dau --help`

### Mac

#### Standalone binary

* Grab the latest binary from `https://github.com/tardisx/discord-auto-upload/releases/latest` called 'dau-mac.gz'.
* `gunzip dau-mac.gz`
* `mv dau-mac dau`
* `chmod +x dau`
* put `dau` somewhere on your path

#### From source

Basically the same as Linux above. [Perlbrew](https://perlbrew.pl) is highly recommended so as not to disturb the system perl. No need for superuser access then either.

### Windows

* Grab the latest windows exe

`https://github.com/tardisx/discord-auto-upload/releases/latest`

* Optional, put it somewhere on your path
* Open a command prompt
* Test it

`\some\path\dau --help`

If you want to hack it, audit it, or don't trust my exe, you can install
[Strawberry Perl](http://strawberryperl.com) and run it using that directly.
You'll need the same dependencies mentioned above in the Linux setup.

If you want to build your own .exe, see BINARIES.md

## Using it

`dau` is a command line driven program. When executed, it will continually scan a directory for new images, and each time it finds one it will upload it to discord, via the discord web hook.

`dau` will only upload "new" screenshots, where "new" means a file that appears in a directory that it is watching, if it appears *after* it has started executing.

Thus, you do not have to worry about pointing `dau` at a directory full of images, it will only upload new ones.

If `dau` is on your path, you can run it from your screenshot folder and there is then no need to specify the path to your images.

Note that currently `dau` does not look in subdirectories. Please submit an issue if this is a use case for you.

The only mandatory command line parameter is the discord webhook URL:

`--webhook URL` - the webhook URL (see [here](https://support.discordapp.com/hc/en-us/articles/228383668-Intro-to-Webhooks) for details).

Other parameters are:

`--watch xx` - specify how many seconds to wait between scanning the directory. The default is 10 seconds.

`--directory <somedir>` - the directory to watch for images to appear in. If this option is not supplied, will look in the current directory.

You will have to quote the path on windows, or anywhere where the directory path contains spaces.

`--username` - supply a 'username' with the webhook submission. Slightly misleading, it basically provides some extra text next to the "Bot" display on the upload to the channel.

In the example screenshot, this was set to "tardisx uploaded from EDD".

`--debug` - provide extra debugging.

## Limitations/bugs

* Only files ending jpg, gif or png are uploaded.
* Subdirectories are not scanned.
* If multiple screenshots occur quickly (<1 second apart) not all may be uploaded.

## TODO
This is just a quick hack. Open to suggestions on new features and improvements.

Open an [issue](https://github.com/tardisx/discord-auto-upload/issues/new) and let me know.
