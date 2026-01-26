#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include "hash_map.h"

// Helper
size_t hash(HashMap *table, const char *key) {
    size_t hash = 0;

    for (size_t i = 0; key[i] != '\0'; i++) {
        hash += (unsigned char)key[i];
    }
    return hash % table->size;
}

// Funcions

HashMap *create_table(size_t size) {
    HashMap *map = malloc(sizeof(HashMap));
    if (!map) return NULL;

    map->size = size;
    map->count = 0;

    map->buckets = calloc(size, sizeof(Entry *));
    if (!map->buckets) {
        free(map);
        return NULL;
    }

    return map;
}

void set_item(HashMap *table, const char *key, const char *value) {
    size_t index = hash(table, key);

    Entry *current = table->buckets[index];

    // TODO: function to double the array size and realloc all the itens
    while (current) {
        if (strcmp(current->key, key) == 0) {
            free(current->value);
            current->value = strdup(value);
            return;
        }
        current = current->next;
    }

    Entry *new_entry = malloc(sizeof(Entry));
    new_entry->key = strdup(key);
    new_entry->value = strdup(value);
    new_entry->next = table->buckets[index];
    
    table->buckets[index] = new_entry;
    table->count++;
}