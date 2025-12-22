# Go-Torrent

A modular BitTorrent client implementation in Go, following the BEP 3 specification.
<img width="983" height="442" alt="image" src="https://github.com/user-attachments/assets/e065e5a1-b81c-4ff6-8dbb-a66465e98950" />

## Features

- **Magnet Link Support**: Full implementation of BEP 9 (Metadata Exchange) and BEP 5 (DHT) for magnet downloads.
- **Modern Web UI**: A futuristic, interactive dashboard built with React, Vite, and Three.js.
- **Real-Time Stats**: Live monitoring of download/upload speeds, active peers, and progress.
- **File Upload**: Support for uploading `.torrent` files directly via the UI.
- **Bencode Parsing**: Robust encoder/decoder for bencoded data.
- **Torrent Parsing**: Support for single and multi-file torrents.
- **Tracker Communication**: HTTP and UDP tracker support.
- **Peer Wire Protocol**: Full implementation (Handshake, Bitfield, Choke/Unchoke, Interested, Request, Piece).

## Prerequisites

- Go 1.25 or later
- Node.js & npm (for Web UI)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/valhalla/go-torrent.git
   cd go-torrent
   ```

2. Build the backend:
   ```bash
   go mod tidy
   go build -o torrent-client ./cmd/client
   ```

3. Setup the Frontend:
   ```bash
   cd web
   npm install
   npm run build
   cd ..
   ```

## Usage

### Web Interface (Recommended)

Start the server:
```bash
./torrent-client serve --port 8080
```
Then open `http://localhost:5173` (if running dev server) or serve the static files (if configured).
*Note: The current dev setup requires running `npm run dev` in the `web/` directory parallel to the backend.*

### CLI Mode

Run the client by providing the path to a `.torrent` file:

```bash
./torrent-client <path_to_torrent_file> <output_directory>
```
Or for magnet links:
```bash
./torrent-client magnet "<magnet_link>" <output_directory>
```

## Status

This project has evolved into a fully functional BitTorrent client with:
- [x] Standard HTTP & UDP Trackers
- [x] DHT (Distributed Hash Table) for peer discovery
- [x] Magnet Link Support (Metadata Exchange)
- [x] Modern Web Dashboard
- [x] File Uploads

