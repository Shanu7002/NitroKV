#include <stdio.h>
#include "arena.h"

#define ARENA_SIZE 1024

unsigned char buffer[ARENA_SIZE];

int main() {
    arena_init(buffer, ARENA_SIZE);

    int *a = arena_alloc(sizeof(int));
    int *b = arena_alloc(sizeof(int));

    *a = 10;
    *b = 20;

    printf("%d %d\n", *a, *b);

    arena_reset();  // frees everything at once

    return 0;
}
