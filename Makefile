# Variables
CC = GCC
CFLAGS = -Wall -Wextra -g -I./internal/storage
TARGET = nitro_test
SRCS = internal/storage/hash_map.c internal/storage/main.c

.PHONY: all compile run valgrind clean run_dev

$(TARGET): $(SRCS)
		$(CC) $(CFLAGS) $^ -o $@

compile: $(TARGET)

run: $(TARGET)
		./$@

valgrind: $(TARGET)
	valgrind --leak-check=full --show-leak-kinds=all ./$@

run_dev: valgrind

clean:	
		rm -r $(TARGET)
