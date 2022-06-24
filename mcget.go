package main

import (
	_ "embed"
	"fmt"
	"github.com/buger/jsonparser"
	"io"
	"net/http"
	"os"
)

const metaUrl = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

//go:embed help.txt
var helpText string

func main() {
	os.Exit(main1())
}

func main1() int {
	var err error
	var out string
	var resp *http.Response
	var launcherMeta, versionMeta []byte

	if len(os.Args) < 2 {
		fmt.Println(helpText)
		return 0
	}

	//check that command is valid
	switch os.Args[1] {
	case "lver", "url", "sha1", "jver":
	default:
		fmt.Println(helpText)
		return 1
	}

	resp, err = http.Get(metaUrl)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error getting launcher metadata: %s\n", err.Error())
		return 1
	}
	launcherMeta, err = io.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error reading launcher metadata: %s\n", err.Error())
		return 1
	}
	_ = resp.Body.Close()

	if os.Args[1] == "lver" { //print latest version of channel
		if len(os.Args) != 3 {
			goto argerr
		}
		switch os.Args[2] { //validate channel input
		case "release", "snapshot":
		default:
			_, _ = fmt.Fprintln(os.Stderr, "Invalid channel specified")
			return 1
		}

		out, err = jsonparser.GetString(launcherMeta, "latest", os.Args[2])
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "Error getting latest version for channel (%s): %s", os.Args[2], err.Error())
			return 1
		}

		fmt.Println(out)
		return 0
	} else {
		if len(os.Args) != 3 {
			goto argerr
		}

		var verUrl string
		i := 0
		for { //loop through list elements until version is found
			out, err = jsonparser.GetString(launcherMeta, "versions", fmt.Sprintf("[%d]", i), "id")
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, "Version not found ")
				return 1
			}
			if out == os.Args[2] {
				break
			}
			i++
		}

		verUrl, err = jsonparser.GetString(launcherMeta, "versions", fmt.Sprintf("[%d]", i), "url")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error getting version url: %s\n", err.Error())
			return 1
		}
		resp, err = http.Get(verUrl)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error getting version meta: %s\n", err.Error())
			return 1
		}
		versionMeta, err = io.ReadAll(resp.Body)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error reading version meta: %s\n", err.Error())
			return 1
		}

		switch os.Args[1] {
		case "sha1", "url":
			out, err = jsonparser.GetString(versionMeta, "downloads", "server", os.Args[1])
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error getting SHA1/URL: %s\n", err.Error())
				return 1
			}
			fmt.Println(out)
			return 0
		case "jver":
			var ver int64
			ver, err = jsonparser.GetInt(versionMeta, "javaVersion", "majorVersion")
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error getting JVM version: %s\n", err.Error())
				return 1
			}
			fmt.Println(ver)
			return 0
		}
	}

argerr:
	_, _ = fmt.Fprintln(os.Stderr, "Invalid arguments passed")
	return 1
}
