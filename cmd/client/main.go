package main

import (
	"fmt"
	"os"

	"github.com/valhalla/go-torrent/pkg/client"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: client <torrent_file> <output_dir>")
		os.Exit(1)
	}

	torrentPath := os.Args[1]
	outputDir := os.Args[2]

	client, err := client.New(torrentPath, outputDir)
	if err != nil {
		fmt.Printf("Error checking args: %v\n", err)
		return
	}
	
	if err := client.Run(); err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		os.Exit(1)
	}
}
