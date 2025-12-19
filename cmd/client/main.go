package main

import (
	"fmt"
	"os"

	"github.com/valhalla/go-torrent/pkg/client"
)

const banner = `
   ____       _____                                  _ 
  / ___| ___ |_   _|__  _ __ _ __ ___ _ __   | |_ 
 | |  _ / _ \  | |/ _ \| '__| '__/ _ \ '_ \  | __|
 | |_| | (_) | | | (_) | |  | | |  __/ | | | | |_ 
  \____|\___/  |_|\___/|_|  |_|  \___|_| |_|  \__|
                                                    
`

func main() {
	fmt.Println(banner)
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
