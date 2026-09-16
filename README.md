# NitroKV

In-memory key-value store featuring a core engine written in C for manual memory management, a Go-based TCP server, and a Write-Ahead Log for disk persistence.

## Overview

NitroKV is a hybrid-language storage engine designed to explore the boundaries of memory efficiency and network throughput.

## Technical Architecture

### Storage Engine

- Custom Hash Table: Built from scratch using the FNV-1a hashing algorithm.
- Collision Resolution: Implemented via linked-list chaining.
- Memory Management: Manual heap control using `malloc` and `free`, featuring a dynamic resizing mechanism.

### Concurrency & Networking

- High Concurrency: Multi-threaded TCP server utilizing Goroutines.
- Network Binding: Configured to bind to `0.0.0.0`.
- Synchronization: Utilization of `sync.RWMutex` to manage thread-safe access to the underlying C pointers across requests.

### Persistence Layer

- Durability: Implements a Write-Ahead Log. All mutations are appended to a sequential log file before being committed to memory.
- Auto-Maintenance: Scheduled log compaction to prevent disk bloat while maintaining sequential disk I/O.

## Performance Analysis

Based on benchmarks performed on a 12th Gen Intel(R) Core(TM) i5-12450H:

| Operation           | Latency (ns/op) | Status                   |
| :------------------ | :-------------- | :----------------------- |
| **In-Memory GET**   | ~2,300 ns       | Ultra-fast (C Engine)    |
| **Parallel SET**    | ~380,000 ns     | High Scalability         |
| **Disk-Synced SET** | ~1,600,000 ns   | Bottlenecked by SSD Sync |

## Usage & Commands

### Prerequisites

- Go 1.25.11+
- GCC (for CGO compilation)
- Make (optional but recommended, you can just the commands in makefile and write ngl)

### Quick Start

1. **Get the host IP:**

   ```bash
      ip addr show or ifconfig(deprecated) -> look for inet with "scope global dynamic noprefixroute wlp9s0"
   ```

2. **Build and Run:**

   ```bash
   make run
   ```

3. **Connect from any device:**

   ```bash
   nc <host/server-ip> 6379<that is the port you can change in main if want to>
   ```

### Available Commands

| Command           | Description                                        | Example                    |
| :---------------- | :------------------------------------------------- | :------------------------- |
| **HELP**          | Displays the help menu with all available options. | `HELP`                     |
| **REGISTER _db_** | Creates a new isolated database instance.          | `REGISTER prod_db`         |
| **LOGIN _db_**    | Enters a specific database session.                | `LOGIN prod_db`            |
| **SET "key" val** | Stores a value (O(1)).                             | `SET "User teste" "Teste"` |
| **GET "key"**     | Retrieves the value associated with a key (O(1)).  | `GET "User teste"`         |
| **REMOVE "key"**  | Deletes a key-value pair from the active database. | `REMOVE "User"`            |
| **RESTORE**       | Restores all databases from disk.                  | `RESTORE`                  |
| **RESTORE _db_**  | Restores specific database from disk.              | `RESTORE prod_db`          |
| **CLOSE**         | Clears memory and logs out the current session.    | `CLOSE`                    |
