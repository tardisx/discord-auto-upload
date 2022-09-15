# Changelog
All notable changes to this project will be documented in this file.

## [Unreleased]

## [v0.12.4] - 2022-09-15

- Document that watcher intervals are in seconds

## [v0.12.3] - 2022-05-09

- Fix a race condition occasionally causing multiple duplicate uploads

## [v0.12.2] - 2022-05-01

- Automatically open your web browser to the `dau` web interface
  (can be disabled in configuration)
- Add system tray/menubar icon with menus to open web interface, quit and
  other links
- Superfluous text console removed on windows

## [v0.12.1] - 2022-05-01

- Show if a new version is available in the web interface
- Rework logging and fix the log display in the web interface

## [v0.12.0] - 2022-04-03

- Break upload page into pending/current/complete sections
- Add preview thumbnails for each upload
- Add feature to hold an image for upload, so the user can
  choose to upload it or not
- Add simple image editor to add text captions
- Discord server created: https://discord.gg/eErG9sntbZ

## [v0.11.2] - 2021-10-19

- Really fix the bug where too large attachments keep retrying
- Fix tests on Windows

## [v0.11.1] - 2021-10-11

- Improve logging and error handling
- Improve tests
- Fix problem where attachments too large for discord fail immediately and do not retry
- Fix problem with version checking

## [v0.11.0] - 2021-10-10

- Switched to semantic versioning
- Now supports multiple watchers - multiple directories can be monitored for new images
- Complete UI rework to support new features and decrease ugliness
- Add many tests

## [0.10.0] - 2021-06-08

This version adds a page showing recent uploads, with thumbnails.

This is not much use except as a log at this stage, but is the basis for future versions which will allow you to hold files before uploading, and edit them (crop, add text, etc) as well.

## [0.9.0] - 2021-06-04

Fix the version update check so that users are actually informed about new releases.

## [0.8.0] - 2021-06-03

This version makes the logs available in the web interface.

## [0.7.0] - 2021-02-09

The long awaited (!) web interface launches with this version. No more messing with command line arguments and .bat files.

Just run the exe and hit http://localhost:9090 to configure the app. See the updated README.md for more information on the configuration.

## [0.6.0] - 2017-02-28

Add --exclude option to avoid uploading files in thumbnail directories

## [0.5.0] - 2017-02-28

* Automatic watermarking of images to perform shameless self-promotion of this tool (disable with --no-watermark)
* Automatically retry failed uploads
* Internal cleanups

## [0.4.0] - 2017-02-28

* Fix crash if the specified directory did not exist
* Better output for showing new version info
* Show speed of upload

## [0.3.0] - 2017-02-21

* Support 'username' sending
* Timeout on all HTTP connections
* Default to current directory if --directory not specified

## [0.2.0] - 2017-02-21

* First golang version, improved output and parsing of responses.
* Built in update checks.

## [0.1.0] - 2017-02-16

Initial release
