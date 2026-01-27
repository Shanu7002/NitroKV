# Variables
CC = gcc
CFLAGS = -Wall -Wextra -g -I./internal/storage
TARGET = nitro_test
SRCS = internal/storage/hash_map.c tests/C/test_hash_map.c

.PHONY: all compile run valgrind clean

all: compile

$(TARGET): $(SRCS)
	$(CC) $(CFLAGS) $^ -o $@

compile: $(TARGET)

run: $(TARGET)
	./$(TARGET)

valgrind: $(TARGET)
	valgrind --leak-check=full --show-leak-kinds=all ./$(TARGET)

clean:
	rm -f $(TARGET)
