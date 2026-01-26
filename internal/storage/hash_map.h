#ifndef NITRO_HASH_MAP_H
#define NITRO_HASH_MAP_H

#include <stddef.h>

typedef struct Entry{
    char *key;
    char *value;
    struct Entry *next;
} Entry;

typedef struct {
    Entry **buckets;
    size_t size;
    size_t count;
} HashMap;

HashMap *create_table(size_t size);
void set_item(HashMap *table, const char *key, const char *value);
char *get_item(HashMap *table, const char *key);
void remove_item(HashMap *table, const char *key);
void free_table(HashMap *table);
size_t hash(HashMap *table, const char *key);
#endif