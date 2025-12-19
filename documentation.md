# BitTorrent Client Implementation Guide

Welcome! This document is designed to help you understand how this BitTorrent client was built, the concepts behind it, and how the code maps to the features.

## 1. Key Concepts (BitTorrent Dictionary)

If you are new to BitTorrent, here are the essential terms used in this project:

- **Peer**: Another computer connected to the network that you upload to or download from.
- **Swarm**: The group of all peers sharing a specific torrent.
- **Tracker**: A server (HTTP or UDP) that introduces peers to each other. It doesn't host files, just IP addresses.
- **InfoHash**: A unique fingerprint (SHA-1 hash) of the torrent's metadata. It identifies the torrent in the network.
- **Bencode**: The encoding format used by `.torrent` files and tracker responses. It supports strings, integers, lists, and dictionaries.
- **Piece**: The file(s) are split into equal-sized chunks called pieces (usually 256KB - 4MB).
- **Block**: Pieces are further split into blocks (usually 16KB) for network transmission.
- **Bitfield**: A string of bits representing which pieces a peer has. If the n-th bit is 1, they have the n-th piece.
- **Choked/Unchoked**: A state of connection. If you are "choked", the peer will not serve you requests. "Unchoking" is enabling data transfer.
- **Interested/Not Interested**: A state indicating if a peer has pieces you want.
- **Seeding**: Uploading pieces to other peers after (or while) you have them.
- **Leeching**: Downloading pieces from others.

## 2. Implementation Timeline

Here is the step-by-step journey of how this client was built:

### Phase 1: Foundation
1.  **Project Setup**: Initialized Go module and directory structure.
2.  **Bencode Parser** (`pkg/bencode`): Implemented a custom decoder to read `.torrent` files and tracker responses.
3.  **Torrent Parser** (`pkg/torrent`): Built logic to read `.torrent` files and extract the `InfoHash`, `Announce` URL, and File/Piece structures.

### Phase 2: Finding Peers
4.  **HTTP Tracker** (`pkg/tracker/client.go`): Implemented the HTTP protocol to ask the tracker for a list of peers.
5.  **UDP Tracker** (`pkg/tracker/udp.go`): Added support for the binary UDP protocol (BEP 15) to support a wider range of trackers.

### Phase 3: Peer Communication (The Wire Protocol)
6.  **Handshake** (`pkg/peer/handshake.go`): Implemented the initial exchange of InfoHash and PeerID to establish a session.
7.  **Message Protocol** (`pkg/peer/message.go`): defined the formats for standard messages: `Choke`, `Unchoke`, `Interested`, `Have`, `Bitfield`, `Request`, `Piece`.
8.  **Connection Logic** (`pkg/peer/peer.go`): Created the `Peer` struct to manage TCP connections and state (choked/interested).

### Phase 4: Downloading (Leeching)
9.  **Bitfield** (`pkg/bitfield`): Created a helper to track which pieces we and our peers have.
10. **Downloader** (`pkg/client/downloader.go`): The core engine. It manages a work queue of pieces and spawns workers for each peer.
11. **Pipelining**: Optimized download speed by allowing multiple block requests (up to 5) to be in flight simultaneously per peer.
12. **File Writer** (`pkg/client/file_writer.go`): Handles assembling downloaded blocks and writing them to the correct offset on disk.

### Phase 5: Seeding & Completion
13. **Uploading** (`pkg/client/uploader.go`): Added logic to handle `MsgRequest` from peers. If we have the piece, we read it from disk and send it back.
14. **Broadcasting** (`pkg/client/result_processor.go`): When we finish a piece, we now send a `Have` message to all peers so they know we can share it.

## 3. Architecture & File Map

Where does the code live?

| Feature | File(s) | Description |
| :--- | :--- | :--- |
| **Main Entry** | `cmd/client/main.go` | Parses arguments and initializes the Client. |
| **Orchestrator** | `pkg/client/client.go` | detailed setup: parses torrent, queries tracker, starts Downloader. |
| **Download Engine** | `pkg/client/downloader.go` | Manages the `WorkQueue`, peer workers, and concurrency. |
| **Upload Logic** | `pkg/client/uploader.go` | Handles incoming requests and serves data. |
| **Tracker (HTTP)** | `pkg/tracker/client.go` | Communicates with HTTP trackers. |
| **Tracker (UDP)** | `pkg/tracker/udp.go` | Communicates with UDP trackers (binary protocol). |
| **Wire Protocol** | `pkg/peer/message.go` | Serialization/Deserialization of protocol messages. |
| **Peer State** | `pkg/peer/peer.go` | Manages an individual peer connection, handshake, and message loop. |
| **File I/O** | `pkg/client/file_writer.go` | Thread-safe(ish) reading and writing to the output file. |
| **Parsing** | `pkg/bencode/*.go` | Low-level decoding of the bencode format. |

## 4. How It Works (The "Lifecycle")

1.  **Parse**: You pass a `.torrent` file. The **Parser** reads it, calculates the `InfoHash` (unique ID), and finds the **Tracker URL**.
2.  **Announce**: The **Client** contacts the Tracker (HTTP or UDP). The Tracker returns a list of IP:Port for other peers.
3.  **Connect**: The **Downloader** starts a worker for each peer. It attempts to open a TCP connection and performs a **Handshake**.
4.  **Bitfield Exchange**: The peer tells us what pieces it has (`Bitfield` message). We tell them what we have.
5.  **Interested**: If the peer has a piece we need, we send `Interested`.
6.  **Unchoke**: We wait for the peer to send `Unchoke`. Once unchoked, we can start requesting data.
7.  **Request Loop**:
    *   The Downloader picks a piece index from the `WorkQueue`.
    *   It breaks the piece into 16KB blocks.
    *   It sends `Request` messages for these blocks (pipelined).
8.  **Downloading**: We receive `Piece` messages containing data. We assemble them into a buffer.
9.  **Writing**: Once a full piece is downloaded and verified, the **Result Processor** writes it to disk.
10. **Sharing**: We send a `Have` message to all connected peers, letting them know we can now upload this piece to them.
11. **Seeding**: If a peer requests a block of a piece we have, the **Uploader** reads it from disk and sends it.

## 5. Study Tips

To understand the code deeply:
1.  **Start with `pkg/peer/message.go`**: Understand the wire protocol bytes. It's the language the peers speak.
2.  **Look at `downloadWorker` in `pkg/client/downloader.go`**: This is the heart of the concurrency. Notice how it handles state changes (`MsgChoke`) and data flow (`MsgPiece`).
3.  **Check `pkg/tracker/udp.go`**: See how we manually construct binary packets for UDP. It's very different from the clean HTTP requests in `client.go`.

Happy Hacking!
