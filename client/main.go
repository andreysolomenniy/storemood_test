package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const uri = "http://localhost:3210/watermark"

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Wrong arguments number.")
		fmt.Println("Usage: client image_path watermark_path result_path")
		return
	}
	imagePath := os.Args[1]
	watermarkPath := os.Args[2]
	resultPath := os.Args[3]

	if !pathExists(imagePath) {
		fmt.Println(imagePath, "does not exist.")
		return
	}
	if !pathExists(watermarkPath) {
		fmt.Println(watermarkPath, "does not exist.")
		return
	}

	fileImage, err := os.Open(imagePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileImage.Close()

	fileWatermark, err := os.Open(watermarkPath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileWatermark.Close()

	// Both files, image and watermark are sending using multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	partImage, err := writer.CreateFormFile("image", filepath.Base(imagePath))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(partImage, fileImage)
	if err != nil {
		log.Fatal(err)
	}
	partWatermark, err := writer.CreateFormFile("watermark", filepath.Base(watermarkPath))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(partWatermark, fileWatermark)
	if err != nil {
		log.Fatal(err)
	}
	err = writer.Close()
	if err != nil {
		log.Fatal(err)
	}

	// After filling request data make request to server
	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	} else {
		// Check request result and write result image to output file
		body := &bytes.Buffer{}
		_, err := body.ReadFrom(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode == http.StatusOK {
			resultFile, _ := os.Create(resultPath)
			_, err = io.Copy(resultFile, body)
			resultFile.Close()
		}
		resp.Body.Close()
	}
}
