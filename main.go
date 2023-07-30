package main

import (
	"bufio"
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {

	parsedURL := flag.String("url", "", "The URL to a Google Web Font (e.g.: https://fonts.googleapis.com/css2?family=Lato:ital,wght@0,400&display=swap)")
	parsedOutputDir := flag.String("output", "output", "A relative or absolute path to the output directory (e.g.: fonts)")
	parsedDestination := flag.String("destination", "", "The location of the CDN or server where the fonts will be self hosted (e.g.: https://cdn.my-server.com/fonts/)")
	flag.Parse()

	url := *parsedURL
	outputDir := *parsedOutputDir
	destination := *parsedDestination

	if len(url) == 0 {
		log.Fatalln("URL is required")
		return
	}
	if len(outputDir) == 0 {
		outputDir = "output"
	}
	if len(destination) > 0 && !strings.HasSuffix(destination, "/") {
		destination += "/"
	}

	err := createDirIfNotExist(outputDir)
	if err != nil {
		panic(err)
	}

	// Setting a different User-Agent will return different file format (e.g. woff2, woff, ttf, etc.)
	userAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko)"
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	originalFile, err := os.Create(outputDir + "/original.css")
	if err != nil {
		panic(err)
	}
	defer originalFile.Close()

	modifiedFile, err := os.Create(outputDir + "/modified.css")
	if err != nil {
		panic(err)
	}
	defer modifiedFile.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		originalFile.WriteString(line + "\n")

		isSRC := strings.HasPrefix(strings.TrimSpace(line), "src: url(")
		if !isSRC {
			modifiedFile.WriteString(line + "\n")
			continue
		}

		url, err := retrieveURL(line)
		if err != nil {
			panic(err)
		}
		generatedFileName, err := downloadFile(url, outputDir)
		if err != nil {
			panic(err)
		}

		modifiedLine := strings.Replace(line, url, destination+generatedFileName, 1)
		modifiedFile.WriteString(modifiedLine + "\n")
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

func createDirIfNotExist(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func retrieveURL(line string) (string, error) {
	// Define the regular expression pattern to match the URL
	pattern := `url\((.*?)\)`

	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	// Find the first match of the pattern in the input line
	match := regex.FindStringSubmatch(line)
	if len(match) != 2 {
		return "", fmt.Errorf("no URL found in the input line")
	}

	// Extract and return the URL from the matched result
	url := match[1]
	return url, nil
}

func downloadFile(url string, dir string) (string, error) {
	// Send an HTTP GET request to the URL
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Check if the request was successful (status code 200)
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: %s", response.Status)
	}

	// Get the file extension from the URL
	originalExt := filepath.Ext(url)

	// Create a temporary file to save the downloaded data
	tempFile, err := os.CreateTemp(dir, "download-*"+originalExt)
	if err != nil {
		return "", err
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Calculate the MD5 hash while copying the HTTP response body to the temporary file
	hash := md5.New()
	writer := io.MultiWriter(tempFile, hash)
	_, err = io.Copy(writer, response.Body)
	if err != nil {
		return "", err
	}

	// Get the MD5 hash as a string and create the final filename using it
	md5String := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	finalFileName := md5String + originalExt
	finalFilePath := filepath.Join(dir, finalFileName)

	// Rename the temporary file to the final filename
	err = os.Rename(tempFile.Name(), finalFilePath)
	if err != nil {
		return "", err
	}

	return finalFileName, nil
}
