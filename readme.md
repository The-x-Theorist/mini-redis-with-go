# mini-redis-with-go

A lightweight Redis-like key-value store server built with Go. Implements a TCP server with thread-safe in-memory storage supporting basic commands (PING, SET, GET, DEL, EXISTS) with automatic expiration. Perfect for learning TCP networking, concurrent programming, and building custom protocols in Go.

## Features

- **TCP-based Server**: Raw TCP socket communication for low-level protocol control
- **Thread-Safe Storage**: Concurrent access using Go's `sync.RWMutex` for safe read/write operations
- **Automatic Expiration**: Built-in TTL (Time To Live) - keys expire after 5 seconds
- **Concurrent Connections**: Handles multiple clients simultaneously using goroutines
- **Simple Protocol**: Line-based text protocol (similar to Redis RESP protocol basics)
- **In-Memory Storage**: Fast key-value operations with Go maps

## Usage/Quick Start

### Prerequisites

- Go 1.19 or higher
- `nc` (netcat) for command-line testing (usually pre-installed on macOS/Linux)

### Installation

```bash
# Clone the repository
git clone https://github.com/The-x-Theorist/mini-redis-with-go.git
cd mini-redis-with-go

# Run the server
go run main.go
```

The server will start listening on `localhost:8000`.

### Quick Test

In a new terminal, connect using `nc`:

```bash
nc localhost 8000
```

Then try some commands:

```
PING
SET name John
GET name
EXISTS name
DEL name
```

## Commands Reference

The server supports the following commands:

| Command | Syntax | Description | Response |
|---------|--------|-------------|----------|
| `PING` | `PING` | Check if server is responsive | `PONG` |
| `SET` | `SET <key> <value>` | Store a key-value pair | `OK` or error message |
| `GET` | `GET <key>` | Retrieve value for a key | Value or error message |
| `DEL` | `DEL <key>` | Delete a key-value pair | `OK` or error message |
| `EXISTS` | `EXISTS <key>` | Check if key exists | `Yes` or `No` |

### Error Responses

- `ERR wrong number of arguments for '<command>' command` - Invalid argument count
- `ERR data doesn't exist` - Key not found
- `ERR data expired` - Key expired (TTL exceeded)
- `ERR property doesn't exist in store` - Key doesn't exist (legacy error)
- `ERR unknown command` - Unrecognized command

## Examples

### Command-Line Examples (using `nc`)

```bash
# Connect to server
nc localhost 8000

# Test connection
PING
# Response: PONG

# Set a value
SET username alice
# Response: OK

# Get the value
GET username
# Response: alice

# Check if key exists
EXISTS username
# Response: Yes

# Wait 5+ seconds, then try to get again (will expire)
GET username
# Response: ERR data expired

# Delete a key
DEL username
# Response: OK

# Try to get deleted key
GET username
# Response: ERR data doesn't exist
```

### Go Client Example

```go
package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	// Connect to server
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Send commands
	commands := []string{
		"PING",
		"SET name GoClient",
		"GET name",
		"EXISTS name",
		"DEL name",
	}

	reader := bufio.NewReader(conn)
	
	for _, cmd := range commands {
		// Send command
		fmt.Fprintf(conn, "%s\n", cmd)
		
		// Read response
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)
		
		fmt.Printf("Command: %s\nResponse: %s\n\n", cmd, response)
	}
}
```

## Architecture/How It Works

### System Architecture

```
┌─────────────┐         TCP Connection          ┌──────────────┐
│   Client 1  │─────────────────────────────────▶│              │
└─────────────┘                                   │              │
                                                  │              │
┌─────────────┐         TCP Connection          │   TCP        │
│   Client 2  │─────────────────────────────────▶│   Server     │
└─────────────┘                                   │   (Port 8000)│
                                                  │              │
┌─────────────┐         TCP Connection          │              │
│   Client N  │─────────────────────────────────▶│              │
└─────────────┘                                   └──────┬───────┘
                                                          │
                                                          │ Shared
                                                          │ Store
                                                          ▼
                                                  ┌──────────────┐
                                                  │              │
                                                  │  In-Memory   │
                                                  │  Key-Value   │
                                                  │    Store     │
                                                  │              │
                                                  │  + RWMutex   │
                                                  │  + TTL       │
                                                  └──────────────┘
```

### Request Flow

```
1. Client connects via TCP
   │
   ▼
2. Server accepts connection
   │
   ▼
3. New goroutine spawned for client
   │
   ▼
4. Read command line-by-line
   │
   ▼
5. Parse command and arguments
   │
   ▼
6. Execute command on shared Store
   │   (with appropriate locking)
   │
   ▼
7. Return response to client
   │
   ▼
8. Continue loop or close connection
```

### Key Components

#### 1. Store Structure

```go
type StoreData struct {
    value string
    ttl   time.Time  // Expiration timestamp
}

type Store struct {
    mu   sync.RWMutex      // Read-write mutex for thread safety
    data map[string]StoreData
}
```

**Why `RWMutex`?**
- Allows multiple concurrent **readers** (GET operations)
- Only one **writer** at a time (SET/DEL operations)
- Better performance than regular `Mutex` for read-heavy workloads

#### 2. Connection Handling

Each client connection runs in its own goroutine:

```go
go handleConnection(conn, store)
```

This allows the server to handle hundreds of concurrent clients without blocking.

#### 3. TTL (Time To Live) Mechanism

- Every `SET` operation stores data with a TTL of **5 seconds**
- On `GET`, the server checks if current time exceeds the TTL
- Expired keys are automatically deleted and return an error

```go
if time.Now().After(storeData.ttl) {
    delete(s.data, key)
    return "ERR data expired"
}
```

## Development

### Using Reflex for Auto-Reload

This project uses [reflex](https://github.com/cespare/reflex) for automatic server restart on code changes.

#### Installation

```bash
go install github.com/cespare/reflex@latest
```

#### Usage

```bash
# Run with reflex (auto-reloads on .go file changes)
reflex

# Or specify config explicitly
reflex -c reflex.conf
```

The `reflex.conf` file contains:
```
# Rebuild on any .go file change
-sr '\.go$' -- go run main.go
```

This watches all `.go` files and automatically restarts the server when changes are detected.

### Project Structure

```
mini-redis-with-go/
├── main.go          # Main server implementation
├── reflex.conf      # Reflex configuration
├── README.md        # This file
└── LICENSE          # MIT License
```

## Testing

### Manual Testing with `nc`

```bash
# Terminal 1: Start server
go run main.go

# Terminal 2: Test basic operations
nc localhost 8000
PING
SET test value
GET test
EXISTS test
DEL test
```

### Testing TTL Expiration

```bash
nc localhost 8000
SET temp data
# Wait 5+ seconds
GET temp
# Should return: ERR data expired
```

### Testing Concurrency

Open multiple terminal windows and connect simultaneously:

```bash
# Terminal 1
nc localhost 8000
SET key1 value1

# Terminal 2 (simultaneously)
nc localhost 8000
SET key2 value2
GET key1
```

### Writing Automated Tests

Example test structure:

```go
func TestStoreOperations(t *testing.T) {
    store := &Store{
        mu: sync.RWMutex{},
        data: make(map[string]StoreData),
    }
    
    // Test SET
    store.Set("test", "value")
    
    // Test GET
    val := store.Get("test")
    assert.Equal(t, "value", val)
    
    // Test EXISTS
    exists := store.Exists("test")
    assert.True(t, exists)
    
    // Test DEL
    store.Del("test")
    exists = store.Exists("test")
    assert.False(t, exists)
}
```

## Performance Considerations

### Strengths

- **Fast Reads**: `RWMutex` allows concurrent reads without blocking
- **In-Memory**: No disk I/O, all operations are RAM-based
- **Goroutines**: Efficient handling of concurrent connections
- **Simple Protocol**: Minimal parsing overhead

### Limitations

- **Fixed TTL**: All keys expire after exactly 5 seconds (not configurable)
- **No Persistence**: Data is lost on server restart
- **Single Server**: No replication or clustering
- **Memory Bound**: Limited by available RAM
- **No Authentication**: No security/access control
- **Simple Protocol**: No support for complex data types (only strings)
- **No Transactions**: No atomic multi-command operations

### Known Issues

1. **TTL Cleanup**: Expired keys are only checked on `GET` - no background cleanup
2. **Error Handling**: Limited error messages, some inconsistencies
3. **Connection Management**: No connection timeout or keepalive handling
4. **Protocol**: No support for binary data or multi-line values

## Concurrency & Thread Safety

### Thread Safety Mechanisms

1. **RWMutex for Reads**: Multiple goroutines can read simultaneously
   ```go
   s.mu.RLock()  // Shared lock for reads
   defer s.mu.RUnlock()
   ```

2. **Mutex for Writes**: Exclusive lock for write operations
   ```go
   s.mu.Lock()   // Exclusive lock for writes
   defer s.mu.Unlock()
   ```

3. **Goroutine-per-Connection**: Each client gets its own goroutine, preventing blocking

### Race Condition Prevention

- All store operations are protected by mutexes
- No shared state access without proper locking
- Safe for concurrent access from multiple clients

## License

MIT License

Copyright (c) 2025 Sneh Khatri

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
