// Package main implements updating current version in index.html of an
// overview site hosted in Apache Tomcat.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	"github.com/gocolly/colly"
)

// configuration represents the configuration
type configuration struct {
	RemoteVersionURL  string
	CurrentVersionURL string
	IndexHTMLFile     string
	SearchString      string
	UseLocal          bool
}

// Constants for the configuration file name, search string in the overview
// site and search string for the comand site
const (
	configurationFileName = "config"
	overviewSearchElement = "p[class=\"small text-center\"]"
	comandSearchElement   = "span[class=\"releaseInformation\"]"
	v                     = "v"
)

var version *bool
var versionNumber string
var gitCommit string
var buildDate string
var execDir string

func init() {
	version = flag.Bool(v, false, "Version")
}

// setExecDir sets the directory where the executable is located
func setExecDir() {
	exec, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	execDir = filepath.Dir(exec)
}

// ReadConfig parses the configuration file and returns a configuration
// struct.
func readConfig(configurationFilename string) configuration {
	viper.SetConfigName(configurationFilename)
	viper.AddConfigPath(execDir)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	var configuration configuration
	err = viper.Unmarshal(&configuration)
	if err != nil {
		log.Fatal(err)
	}

	return configuration
}

// Set version to given pointer for the given version URL.
func getVersionFromURLWithChan(versionURL, searchElement string, version *string, c chan string) {
	collector := colly.NewCollector()
	collector.OnHTML(searchElement, func(e *colly.HTMLElement) {
		*version = (strings.TrimSpace(e.Text))
	})

	err := collector.Visit(versionURL)
	if err != nil {
		log.Fatal(err, " ", versionURL)
	}

	c <- "Done"
}

// Returns version for the given version URL.
func getVersionFromURL(versionURL, searchElement string) string {
	var version string

	collector := colly.NewCollector()
	collector.OnHTML(searchElement, func(e *colly.HTMLElement) {
		version = strings.TrimSpace(e.Text)
	})

	err := collector.Visit(versionURL)
	if err != nil {
		log.Fatal(err, " ", versionURL)
	}

	return version
}

// Set version to given pointer for the given file file name.
func getVersionFromFileWithChan(fileName, searchElement string, version *string, c chan string) {
	dir, _ := filepath.Split(fileName)

	fileTransport := &http.Transport{}
	fileTransport.RegisterProtocol("file", http.NewFileTransport(http.Dir(dir)))

	collector := colly.NewCollector()
	collector.WithTransport(fileTransport)
	collector.OnHTML(searchElement, func(e *colly.HTMLElement) {
		*version = strings.TrimSpace(e.Text)
	})

	err := collector.Visit("file://" + fileName)
	if err != nil {
		log.Fatal(err, " ", fileName)
	}

	c <- "Done"
}

// Returns version from given file name.
func getVersionFromFile(fileName, searchElement string) string {
	var version string

	dir, _ := filepath.Split(fileName)

	fileTransport := &http.Transport{}
	fileTransport.RegisterProtocol("file", http.NewFileTransport(http.Dir(dir)))

	collector := colly.NewCollector()
	collector.WithTransport(fileTransport)
	collector.OnHTML(searchElement, func(e *colly.HTMLElement) {
		version = strings.TrimSpace(e.Text)
	})

	err := collector.Visit("file://" + fileName)
	if err != nil {
		log.Fatal(err, " ", fileName)
	}

	return version
}

// Updates the version in the given file name. Reads the file, replaces the
// old version with the new one.
func updateCurrentVersion(fileName, oldVersion, newVersion string) {
	bytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}

	fileContent := string(bytes)
	updatedFileContent := strings.Replace(fileContent, oldVersion, newVersion, 1)

	info, err := os.Stat(fileName)

	err = ioutil.WriteFile(fileName, []byte(updatedFileContent), info.Mode())
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	setExecDir()

	if *version {
		fmt.Println("Version: " + versionNumber)
		fmt.Println("Git Commit: " + gitCommit)
		fmt.Println("Build Date: " + buildDate)
		os.Exit(0)
	}

	configuration := readConfig(configurationFileName)

	var remoteVersion string
	var currentVersion string
	c := make(chan string, 2)
	if configuration.UseLocal == true {
		go getVersionFromFileWithChan(configuration.IndexHTMLFile, overviewSearchElement, &currentVersion, c)
	} else {
		go getVersionFromURLWithChan(configuration.CurrentVersionURL, overviewSearchElement, &currentVersion, c)
	}
	go getVersionFromURLWithChan(configuration.RemoteVersionURL, comandSearchElement, &remoteVersion, c)
	<-c
	<-c
	if remoteVersion != "" && strings.Contains(remoteVersion, configuration.SearchString) && currentVersion != "" {
		if remoteVersion != currentVersion {
			updateCurrentVersion(configuration.IndexHTMLFile, currentVersion, remoteVersion)
		}
	}
}
