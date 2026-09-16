# Variables
CC = gcc
CFLAGS = -Wall -Wextra -g -I./internal/storage
TARGET_C = nitroKV_test
SOURCE_C = internal/storage/hash_map.c tests/unit/test_hash_map.c

# Go Variables
GO_BINARY = nitrokv
GO_MAIN = cmd/server/main.go

.PHONY: all compile run valgrind clean build-go

all: compile build-go

# Compile C tests
$(TARGET_C): $(SOURCE_C)
	$(CC) $(CFLAGS) $^ -o $@

compile: $(TARGET_C)

# Run C tests
run-cTests: $(TARGET_C)
	./$(TARGET_C)

# Run C test while check for memory leaks
valgrind: $(TARGET_C)
	valgrind --leak-check=full --show-leak-kinds=all ./$(TARGET_C)

# Compile server (Go + CGO)
build-go:
	go build -o $(GO_BINARY) $(GO_MAIN)

# Run real server
run: build-go
	./$(GO_BINARY)

clean:
	rm -f $(TARGET_C) $(GO_BINARY)
	rm -rf data/*.log

# Run benchmarks
bench-server: build-go
	go test -bench=. -benchmem -count=5 ./internal/network/...
