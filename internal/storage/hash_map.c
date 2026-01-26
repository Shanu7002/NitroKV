#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include "hash_map.h"

// Helper
size_t hash(HashMap *table, const char *key) {
    const size_t FNV_offset_basis = 14695981039346656037ULL;
    const size_t FNV_prime = 1099511628211ULL;

    size_t hash_val = FNV_offset_basis;

    for (size_t i = 0; key[i] != '\0'; i++) {
        hash_val ^= (unsigned char)key[i];
        hash_val *= FNV_prime;
    }

    return hash_val % table->size;
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
            char *new_val = strdup(value);
            if(!new_val) return;

            free(current->value);
            current->value = new_val;
            return;
        }
        current = current->next;
    }

    Entry *new_entry = malloc(sizeof(Entry));
    if (!new_entry) return;

    new_entry->key = strdup(key);
    new_entry->value = strdup(value);

    if (!new_entry->key || !new_entry->value) {
        free(new_entry->key);
        free(new_entry->value);
        free(new_entry);
        return;
    }

    new_entry->next = table->buckets[index];    
    table->buckets[index] = new_entry;
    table->count++;
}