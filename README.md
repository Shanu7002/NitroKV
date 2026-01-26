# NitroKV | High-Performance Hybrid Storage Engine

A high-performance, in-memory key-value store featuring a core engine written in **C** for manual memory management, a **Go**-based TCP server for high concurrency, and a **Write-Ahead Log (WAL)** for $O(1)$ disk persistence.

## 🚀 Overview
NitroKV is a hybrid-language storage engine designed to explore the boundaries of memory efficiency and network throughput. By offloading the storage logic to **C** and the networking layer to **Go**, this project demonstrates a sophisticated understanding of language interoperability and systems architecture.



## 🛠️ Technical Architecture

### 1. Storage Engine (C)
* **Custom Hash Table:** Built from scratch using the **FNV-1a** hashing algorithm.
* **Collision Resolution:** Implemented via linked-list chaining.
* **Memory Management:** Manual heap control using `malloc` and `free`, featuring a dynamic resizing (rehashing) mechanism to maintain $O(1)$ lookup complexity.

### 2. Concurrency Layer (Go)
* **Networking:** Multi-threaded TCP server utilizing **Goroutines** to handle thousands of simultaneous client connections.
* **Binary Protocol:** Custom-designed byte-level protocol for reduced payload size and faster parsing compared to JSON/Text.
* **Synchronization:** Utilization of `sync.RWMutex` to manage thread-safe access to the underlying C pointers.

### 3. Persistence Layer (WAL)
* **Durability:** Implements a **Write-Ahead Log**. All mutations are appended to a sequential log file before being committed to memory.
* **Performance:** Optimized for $O(1)$ sequential disk I/O to prevent disk bottlenecks during high-frequency writes.

## 📈 Engineering Standards & Analysis
* **Time Complexity:** * **Read/Write:** Average $O(1)$
    * **Recovery:** $O(N)$ where $N$ is the number of log entries.
* **Memory Management:** Zero-leak policy. Profiled using Valgrind to ensure integrity during long-running processes.
* **Interoperability:** Uses **CGO** to bridge the high-level concurrency of Go with the low-level efficiency of C.
