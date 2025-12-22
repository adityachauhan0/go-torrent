package main

import (
	"fmt"
	"os"

	"github.com/valhalla/go-torrent/pkg/client"
	"github.com/valhalla/go-torrent/pkg/server"
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
	
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		startServer()
	case "download":
		startDownload()
	default:
		// Fallback for backward compatibility or simple usage
		// Check if it looks like a file path
		if _, err := os.Stat(command); err == nil {
			// It's a file, assume download mode
			startDownloadWithArgs(command, os.Args[2:])
		} else {
			fmt.Printf("Unknown command: %s\n", command)
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client serve [--port <port>]          Start the web GUI server")
	fmt.Println("  client download <file> <out_dir>      Download a torrent")
}

func startServer() {
	port := "8080"
	if len(os.Args) > 3 && os.Args[2] == "--port" {
		port = os.Args[3]
	}
    
    // Initialize Manager
    manager := client.NewManager()
    

	
	fmt.Printf("Starting Web GUI on http://localhost:%s\n", port)
	srv := server.NewServer(port, manager)
	if err := srv.Start(); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func startDownload() {
	if len(os.Args) < 4 {
		printUsage()
		os.Exit(1)
	}
	startDownloadWithArgs(os.Args[2], os.Args[3:])
}

func startDownloadWithArgs(torrentPath string, args []string) {
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}
	outputDir := args[0]

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
