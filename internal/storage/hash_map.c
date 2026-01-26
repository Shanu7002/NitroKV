#include <stdlib.h>
#include <errno.h>
#include "hash_map.h"

// Helper
int is_int(const char *str, int *out) {
    char *end;
    errno = 0;

    long val = strtol(str, &end, 10);

    if (errno != 0 || *end != '\0') {
        return 0;
    }

    *out = (int) val;
    return 1;
}

int hash(HashMap *table, const char *key) {
    int value;
    int hash;

    if (is_int(key, &value)) {
        return value % table->size;
    }
    for (int i = 0; key[i] != "\0"; i++) {
        hash += key[i];
    }
    return hash % table->size;
}