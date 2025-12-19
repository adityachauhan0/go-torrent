# Go-Torrent

A modular BitTorrent client implementation in Go, following the BEP 3 specification.

## Features

- **Bencode Parsing**: Robust encoder/decoder for bencoded data.
- **Torrent Parsing**: Support for single and multi-file torrents.
- **Tracker Communication**: HTTP tracker support with compact peer lists.
- **Peer Wire Protocol**: Full implementation of the peer protocol (Handshake, Bitfield, Choke/Unchoke, Interested/Not Interested, Request, Piece).
- **Pipelining**: Efficient piece downloading with request pipelining.
- **Modular Design**: Clean separation of concerns (Peer, Client, Tracker, Torrent packages).

## Prerequisites

- Go 1.25 or later

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/valhalla/go-torrent.git
   cd go-torrent
   ```

2. Build the client:
   ```bash
   go mod tidy
   go build -o torrent-client ./cmd/client
   ```

## Usage

Run the client by providing the path to a `.torrent` file and an output directory:

```bash
./torrent-client <path_to_torrent_file> <output_directory>
```

### Example

```bash
./torrent-client ubuntu.iso.torrent ./downloads
```

## Project Structure

- `cmd/client/`: Application entry point.
- `pkg/bencode/`: Bencode encoding and decoding.
- `pkg/torrent/`: .torrent file parsing and structures.
- `pkg/tracker/`: Tracker communication logic.
- `pkg/peer/`: Peer wire protocol implementation.
- `pkg/client/`: Download orchestration and file writing.
- `pkg/p2p/`: Network connection helpers.
- `pkg/bitfield/`: Bitfield manipulation utilities.

## Documentation
For a detailed guide on the codebase, implementation timeline, and key concepts, see [documentation.md](documentation.md).

## Status

This is a functional implementation of the core BitTorrent protocol. It currently supports:
- Leeching (downloading)
- Seeding (uploading)
- Standard HTTP Trackers
- UDP Trackers (BEP 15)
- Bitfield/Have Broadcasting

Future improvements:
- DHT implementation
- Magnet links
