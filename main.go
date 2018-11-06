package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/codenickstar/imgdl/album"

	"go.uber.org/zap"
)

func main() {
	// logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	l := logger.Sugar()

	// album
	var job album.Album
	var err error

	// read album from env
	url := os.Getenv("ALBUM_URL")
	if url == "" {
		// read album url from stdin
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Println("Enter album URL:")
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) != 0 {
				job, err = album.New(l, line)
				if err != nil {
					continue
				}
				break
			}
			fmt.Println("Please enter a valid URL:")
		}
	} else {
		job, err = album.New(l, url)
		if err != nil {
			l.DPanic("error creating album", err)
		}
	}

	// start new job
	err = job.Download()
	if err != nil {
		l.DPanic(err)
	}
}
