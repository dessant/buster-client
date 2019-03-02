package main

import (
	"io"
	"net/http"
	"os"
	"time"
)

func downloadFile(filePath string, url string) error {
	httpClient := &http.Client{Timeout: 120 * time.Second}

	resp, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}
