package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Plugin struct {
	Author              string   `json:"Author"`
	Name                string   `json:"Name"`
	Description         string   `json:"Description"`
	Punchline           string   `json:"Punchline"`
	Changelog           string   `json:"Changelog"`
	Tags                []string `json:"Tags"`
	InternalName        string   `json:"InternalName"`
	AssemblyVersion     string   `json:"AssemblyVersion"`
	RepoUrl             string   `json:"RepoUrl"`
	ApplicableVersion   string   `json:"ApplicableVersion"`
	DalamudApiLevel     int      `json:"DalamudApiLevel"`
	IconUrl             string   `json:"IconUrl"`
	ImageUrls           []string `json:"ImageUrls"`
	DownloadLinkInstall string   `json:"DownloadLinkInstall"`
	IsHide              bool     `json:"IsHide"`
	IsTestingExclusive  bool     `json:"IsTestingExclusive"`
	DownloadLinkTesting string   `json:"DownloadLinkTesting"`
	DownloadLinkUpdate  string   `json:"DownloadLinkUpdate"`
	DownloadCount       int      `json:"DownloadCount"`
	LastUpdated         int64    `json:"LastUpdated"`
}

type Release struct {
	Assets []struct {
		DownloadCount int    `json:"download_count"`
		UpdatedAt     string `json:"updated_at"`
	} `json:"assets"`
}

func main() {
	var plugins []Plugin

	folder, err := os.Open("Plugins")
	if err != nil {
		fmt.Println("error opening folder:", err)
		return
	}
	defer folder.Close()

	contents, err := folder.Readdir(-1)
	if err != nil {
		fmt.Println("error reading contents of folder:", err)
		return
	}

	for _, info := range contents {
		if info.IsDir() {
			file, err := os.Open(filepath.Join("Plugins", info.Name(), info.Name()+".json"))
			if err != nil {
				fmt.Printf("error opening .json in %s: %v\n", info.Name(), err)
				continue
			}
			defer file.Close()

			bytes, err := io.ReadAll(file)
			if err != nil {
				fmt.Printf("error reading .json in %s: %v\n", info.Name(), err)
				continue
			}

			var plugin Plugin
			if err := json.Unmarshal(bytes, &plugin); err != nil {
				fmt.Printf("error unmarshaling .json in %s: %v\n", info.Name(), err)
				continue
			}

			downloadLink := plugin.RepoUrl + "/releases/latest/download/latest.zip"

			plugin.DownloadLinkInstall = downloadLink
			plugin.DownloadLinkTesting = downloadLink
			plugin.DownloadLinkUpdate = downloadLink

			//region get downloads from GitHub api

			api := strings.Replace(plugin.RepoUrl, "github.com", "api.github.com/repos", -1) + "/releases"
			getApi, _ := http.Get(api)
			body, _ := io.ReadAll(getApi.Body)

			var releases []Release
			err = json.Unmarshal(body, &releases)
			if err != nil {
				fmt.Println(err)
				return
			}

			var totalDownloadCount int
			for _, release := range releases {
				for _, asset := range release.Assets {
					totalDownloadCount += asset.DownloadCount
				}
			}
			//endregion

			//region get latest update time from GitHub api

			latestApi := api + "/latest"
			getLatestApi, _ := http.Get(latestApi)
			body, _ = io.ReadAll(getLatestApi.Body)

			var latestRelease Release
			err = json.Unmarshal(body, &latestRelease)
			if err != nil {
				return
			}

			var updatedString = latestRelease.Assets[0].UpdatedAt
			var lastUpdated, _ = time.Parse(time.RFC3339, updatedString)

			//endregion

			plugin.DownloadCount = totalDownloadCount
			plugin.LastUpdated = lastUpdated.Unix()

			plugins = append(plugins, plugin)
			fmt.Println(fmt.Sprintf("added %s to plugin manifest", plugin.Name))
		}
	}

	bytes, err := json.MarshalIndent(plugins, "", "  ")
	if err != nil {
		fmt.Println("error marshaling plugins:", err)
		return
	}

	err = os.WriteFile("pluginmaster.json", bytes, 0644)
	if err != nil {
		fmt.Println("error writing to file:", err)
		return
	}

	fmt.Println(fmt.Sprintf("successfully generated pluginmaster with %d plugins", len(contents)))
}
