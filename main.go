package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Embed struct {
	Description string `json:"description"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <path/to/logfile> <webhookUrl>")
		return
	}

	filePath := os.Args[1]
	webhookUrl := os.Args[2]
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalln("Can't read log file")
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return
	}

	// Read from last line
	fileSize := fileInfo.Size()
	file.Seek(fileSize, io.SeekStart)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(500 * time.Millisecond)

				truncated, err := isTruncated(file)
				if err != nil {
					break
				}

				if truncated {
					_, err := file.Seek(0, io.SeekStart)
					if err != nil {
						break
					}
				}
				continue
			}
			break
		}

		go postWebhook(webhookUrl, line)
		fmt.Printf("%s", string(line))
	}
}

func isTruncated(file *os.File) (bool, error) {
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}

	return currentPos > fileInfo.Size(), nil
}

func postWebhook(url string, content string) {
	values := map[string]interface{}{"content": "", "embeds": []Embed{
		{
			Description: fmt.Sprintf("```%v```", content),
		},
	}}
	jsonData, err := json.Marshal(values)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Print("Error can't post to webhook")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Can't read response body")
	}

	fmt.Println(string(body))
}
