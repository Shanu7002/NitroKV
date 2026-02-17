# Variables
CC = gcc
CFLAGS = -Wall -Wextra -g -I./internal/storage
TARGET_C = nitro_test
SRCS_C = internal/storage/hash_map.c tests/unit/test_hash_map.c

# Go Variables
GO_BINARY = nitrokv
GO_MAIN = cmd/server/main.go

.PHONY: all compile run valgrind clean build-go

all: compile build-go

# Compile and run C
$(TARGET_C): $(SRCS_C)
	$(CC) $(CFLAGS) $^ -o $@

compile: $(TARGET_C)

run-c: $(TARGET_C)
	./$(TARGET_C)

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