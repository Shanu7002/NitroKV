# NitroKV | High-Performance Hybrid Storage Engine

A high-performance, in-memory key-value store featuring a core engine written in **C** for manual memory management, a **Go**-based TCP server for high concurrency, and a **Write-Ahead Log (WAL)** for O(1) disk persistence.

---

## 🚀 Overview
NitroKV is a hybrid-language storage engine designed to explore the boundaries of memory efficiency and network throughput. By offloading the storage logic to **C** and the networking layer to **Go**, this project demonstrates a sophisticated understanding of language interoperability, systems architecture, and distributed networking.

---

## 🛠️ Technical Architecture

### 1. Storage Engine (C)
* **Custom Hash Table:** Built from scratch using the **FNV-1a** hashing algorithm.
* **Collision Resolution:** Implemented via linked-list chaining.
* **Memory Management:** Manual heap control using `malloc` and `free`, featuring a dynamic resizing mechanism to maintain O(1) lookup complexity.

### 2. Concurrency & Networking (Go)
* **High Concurrency:** Multi-threaded TCP server utilizing **Goroutines** to handle thousands of simultaneous client connections.
* **Network Binding:** Configured to bind to `0.0.0.0`, allowing for remote access from external devices and cloud environments (Oracle Cloud/AWS).
* **Synchronization:** Utilization of `sync.RWMutex` to manage thread-safe access to the underlying C pointers across concurrent requests.

### 3. Resilience & Security
* **Authentication Layer:** Secure session management via `AUTH <password>`. 
* **Guest Rate Limiting:** Built-in protection against brute-force and DoS attacks. Unauthenticated users (Guests) are limited to **1 request/second**.
* **Environment Aware:** Supports `NITRO_ENV=test` to bypass limits during high-frequency stress testing and internal benchmarks.

### 4. Persistence Layer (WAL)
* **Durability:** Implements a **Write-Ahead Log**. All mutations are appended to a sequential log file before being committed to memory.
* **Auto-Maintenance:** Scheduled log compaction to prevent disk bloat while maintaining O(1) sequential disk I/O.

---

## 📈 Performance Analysis
Based on benchmarks performed on a 12th Gen Intel(R) Core(TM) i5-12450H:

| Operation | Latency (ns/op) | Status |
| :--- | :--- | :--- |
| **In-Memory GET** | ~2,300 ns | Ultra-fast (C Engine) |
| **Parallel SET** | ~380,000 ns | High Scalability |
| **Disk-Synced SET** | ~1,600,000 ns | Bottlenecked by SSD Sync |

> **Note:** The memory engine is ~680x faster than disk persistence, highlighting the efficiency of the C implementation.

---

## 📡 Usage & Commands

### Prerequisites
* **Go** 1.20+
* **GCC** (for CGO compilation)
* **Make**

### Quick Start
1. **Configure Environment:**
   ```bash
   export NITRO_PWD="your_password"
   export NITRO_ENV="production"
   ```
2. **Build and Run:**
   ```bash
   make run
   ```
3. **Connect from any device:**
   ```bash
   nc <server-ip> 6379
   ```
### ⌨️ Available Commands

| Command | Description | Example |
| :--- | :--- | :--- |
| **HELP** | Displays the help menu with all available options. | `HELP` |
| **AUTH <pwd>** | Authenticates for Admin access (No Rate Limit). | `AUTH my_secret` |
| **REGISTER <db>** | Creates a new isolated database instance. | `REGISTER prod_db` |
| **LOGIN <db>** | Enters a specific database session. | `LOGIN prod_db` |
| **SET "key" val** | Stores a value (O(1)). | `SET "User" "Eduardo"` |
| **GET "key"** | Retrieves the value associated with a key (O(1)). | `GET "User"` |
| **REMOVE "key"** | Deletes a key-value pair from the active database. | `REMOVE "User"` |
| **RESTORE** | Restores all databases from disk. | `RESTORE` |
| **RESTORE <db>** | Restores specific database from disk. | `RESTORE prod_db` |
| **CLOSE** | Clears memory and logs out the current session. | `CLOSE` |