package main

import (
	"bufio"
	"fmt"
	"imgdl/album"
	"os"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	l := logger.Sugar()
	var job album.Album

	// read album url from stdin
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter album URL:")
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) != 0 {
			var err error
			job, err = album.New(l, line)
			if err != nil {
				continue
			}
			break
		}
		fmt.Println("Please enter a valid URL:")
	}

	// start new job
	err := job.Download()
	if err != nil {
		panic(err)
	}
}
