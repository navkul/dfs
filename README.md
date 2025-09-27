# Distributed File System (DFS)

A decentralized, peer-to-peer distributed file system implementation in Go

## Overview

While the current implementation uses **Content-Addressable Storage (CAS)** and **TCP transport**, the system is designed with modular interface-driven architecture that allows users to implment their desired storage backends, transport protocols, and networking layers.

The core architecture separates concerns through well-defined interfaces:
- **Transport Interface**: Currently implemented with TCP, but can be extended to UDP, WebSockets, or any custom protocol
- **Storage Interface**: Currently uses CAS with SHA-1 hashing, but supports any key-value storage mechanism
- **Encoding Interface**: Currently uses GOB serialization, but can be swapped for JSON, Protocol Buffers, or custom formats


## Architecture & Key Features

### Distributed Systems Principles
- **Decentralization**: No central authority or single point of failure - all nodes are equal peers
- **Fault Tolerance**: System continues to operate even if individual nodes fail
- **Scalability**: New peers can join the network dynamically
- **Network Transparency**: Files can be retrieved from any peer without knowing their location

### Current Implementation
- **Content-Addressable Storage (CAS)**: Files identified and stored using SHA-1 cryptographic hashes
- **TCP Transport Layer**: Reliable, connection-oriented communication between peers
- **AES Encryption**: All file transfers and storage encrypted for security
- **Automatic Replication**: Files automatically distributed across all connected peers
- **Message-based P2P Protocol**: Structured communication for file operations and peer coordination

## System Components

### Core Components

1. **FileServer**: Main server component managing peers and file operations
2. **Store**: Local storage engine with encryption/decryption capabilities
3. **Transport Layer**: TCP-based peer-to-peer communication
4. **Crypto Module**: Encryption, decryption, and key generation utilities

### Message Types

- **MessageStoreFile**: Notifies peers about new file availability
- **MessageGetFile**: Requests file from network peers
- **Stream Messages**: Binary file data transfer

## How It Works

1. **File Storage**: When a file is stored on one node:
   - File is encrypted and stored locally using its hash as the key
   - A broadcast message notifies all connected peers
   - Peers receive and store encrypted copies of the file

2. **File Retrieval**: When a file is requested:
   - System first checks local storage
   - If not found locally, broadcasts a request to all peers
   - First available peer serves the file
   - File is decrypted and returned to the requester

3. **Peer Discovery**: Nodes connect to bootstrap peers and maintain persistent connections

## Future Enhancements

- Implement intelligent caching mechanisms to improve file retrieval performance and reduce network overhead
- Add peer discovery protocol to automatically find and connect to peers without manual bootstrap configuration
- Implement distributed delete functionality to remove files across all peers in the network

## Installation & Setup

### Prerequisites

- Go 1.18 or higher
- Available network ports (default: 3000, 4000, 5001)

### Building the Project

```bash
# Clone the repository
git clone <repository-url>
cd dfs

# Build the project
make build

# Or build manually
go build -o bin/dfs
```

## Running the System

### Basic Usage

```bash
# Run the distributed file system with 3 nodes
make run

# Or run manually
./bin/dfs
```

### Default Configuration

The system starts with three nodes:
- **Node 1**: Port 3000 (Bootstrap node)
- **Node 2**: Port 4000 (Bootstrap node)  
- **Node 3**: Port 5001 (Client node, connects to nodes 1 & 2)

Node 3 performs the following operations:
1. Stores 20 test files (`picture_0.png` through `picture_19.png`)
2. Deletes each file locally after storage
3. Retrieves each file from the network peers
4. Verifies file integrity

## Testing

### Running Tests

```bash
# Run all tests
make test

# Or run manually
go test ./... -v
```

### Manual Testing

You can modify the `main.go` file to test different scenarios:

```go
// Store a custom file
data := bytes.NewReader([]byte("Hello, distributed world!"))
s3.Store("my-file.txt", data)

// Retrieve the file
reader, err := s3.Get("my-file.txt")
if err != nil {
    log.Fatal(err)
}

content, _ := ioutil.ReadAll(reader)
fmt.Println(string(content))
```

## Configuration

### Custom Node Setup

```go
// Create a custom node configuration
server := makeServer(":8000", ":3000", ":4000") // Listen on 8000, connect to 3000 and 4000
go server.Start()
```

## Project Structure

```
dfs/
├── main.go              # Application entry point and demo
├── server.go            # Core file server implementation
├── store.go             # Local storage engine
├── crypto.go            # Encryption/decryption utilities
├── p2p/                 # Peer-to-peer networking layer
│   ├── transport.go     # Transport interface
│   ├── tcp_transport.go # TCP implementation
│   ├── encoding.go      # Message encoding/decoding
│   ├── handshake.go     # Peer handshake protocol
│   └── message.go       # Message type definitions
├── *_test.go           # Unit tests
├── Makefile            # Build automation
└── README.md           # This file
```

**Important Note**: Some systems may have processes running on the default ports. If you encounter port conflicts:

1. **Check for conflicts**:
   ```bash
   lsof -i:3000,4000,5001
   ```

2. **Modify ports in main.go**:
   ```go
   s1 := makeServer(":3001", "")           // Change from :3000
   s2 := makeServer(":4001", "")           // Change from :4000  
   s3 := makeServer(":5002", ":3001", ":4001") // Change all ports
   ```

3. **Common conflicting services**:
   - Port 3000: Development servers (React, Node.js)
   - Port 4000: Development servers, some system services
   - Port 5000: macOS Control Center (AirPlay Receiver)



