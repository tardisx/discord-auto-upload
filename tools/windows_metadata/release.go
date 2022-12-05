package main

import (
	"log"
	"os"
	"regexp"
	"strings"

	"golang.org/x/mod/semver"
)

const versionInfoTemplate = `
{
    "FixedFileInfo": {
        "FileVersion": {
            "Major": MAJOR,
            "Minor": MINOR,
            "Patch": PATCH,
            "Build": 0
        },
        "ProductVersion": {
            "Major": MAJOR,
            "Minor": MINOR,
            "Patch": PATCH,
            "Build": 0
        },
        "FileFlagsMask": "3f",
        "FileFlags ": "00",
        "FileOS": "040004",
        "FileType": "01",
        "FileSubType": "00"
    },
    "StringFileInfo": {
        "Comments": "",
        "CompanyName": "tardisx@github",
        "FileDescription": "https://github.com/tardisx/discord-auto-upload",
        "FileVersion": "",
        "InternalName": "",
        "LegalCopyright": "https://github.com/tardisx/discord-auto-upload/blob/master/LICENSE",
        "LegalTrademarks": "",
        "OriginalFilename": "",
        "PrivateBuild": "",
        "ProductName": "discord-auto-upload",
        "ProductVersion": "VERSION",
        "SpecialBuild": ""
    },
    "VarFileInfo": {
        "Translation": {
            "LangID": "0409",
            "CharsetID": "04B0"
        }
    },
    "IconPath": "dau.ico",
    "ManifestPath": ""
}
`

var nonAlphanumericRegex = regexp.MustCompile(`[^0-9]+`)

func main() {
	version := os.Args[1]
	if !semver.IsValid(version) {
		panic("bad version" + version)
	}
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		log.Fatalf("bad version: %s", version)
	}

	parts[0] = nonAlphanumericRegex.ReplaceAllString(parts[0], "")
	parts[1] = nonAlphanumericRegex.ReplaceAllString(parts[1], "")
	parts[2] = nonAlphanumericRegex.ReplaceAllString(parts[2], "")

	out := versionInfoTemplate
	out = strings.Replace(out, "MAJOR", parts[0], -1)
	out = strings.Replace(out, "MINOR", parts[1], -1)
	out = strings.Replace(out, "PATCH", parts[2], -1)
	out = strings.Replace(out, "VERSION", version, -1)

	f, _ := os.Create("versioninfo.json")
	f.Write([]byte(out))
	f.Close()

}
