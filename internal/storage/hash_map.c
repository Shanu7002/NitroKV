#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include "hash_map.h"

// Helpers
static size_t get_hash_raw(const char *key) {
    const size_t FNV_offset_basis = 14695981039346656037ULL;
    const size_t FNV_prime = 1099511628211ULL;
    size_t hash_val = FNV_offset_basis;

    for (size_t i = 0; key[i] != '\0'; i++) {
        hash_val ^= (unsigned char)key[i];
        hash_val *= FNV_prime;
    }

    return hash_val;
}

size_t hash(HashMap *table, size_t hash_val) {
    return hash_val % table->size;
}

static void resize_table(HashMap *table) {
    size_t old_size = table->size;
    size_t new_size = old_size * 2;
    
    Entry **new_buckets = calloc(new_size, sizeof(Entry *));
    if (!new_buckets) return;

    for (size_t i = 0; i < old_size; i++) {
        Entry *entry = table->buckets[i];
        while (entry) {
            Entry *next = entry->next;

            size_t h = get_hash_raw(entry->key);
            size_t inx = h % new_size;

            entry->next = new_buckets[inx];
            new_buckets[inx] = entry;
            entry = next;
        }
    }

    free(table->buckets);
    table->buckets = new_buckets;
    table->size = new_size;
}

// Functions
HashMap *create_table(size_t size) {
    HashMap *table = malloc(sizeof(HashMap));
    if (!table) return NULL;

    table->size = size;
    table->count = 0;
    table->buckets = calloc(size, sizeof(Entry *));
    if (!table->buckets) {
        free(table);
        return NULL;
    }
    return table;
}

void set_item(HashMap *table, const char *key, const char *value) {
    if ((double)table->count / table->size > 0.75f) {
        resize_table(table);
    }

    if (!table || !key || !value) return;

    size_t raw_hash = get_hash_raw(key);
    size_t index = hash(table, raw_hash);
    Entry *currently = table->buckets[index];

    while (currently) {
        if (strcmp(currently->key, key) == 0) {
            char *new_val = strdup(value);
            if (!new_val) return;
            free(currently->value);
            currently->value = new_val;
            return;
        }
        currently = currently->next;
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

const char *get_item(HashMap *table, const char *key) {
    if (!table || !key) return NULL;

    size_t raw_hash = get_hash_raw(key);
    size_t index = hash(table, raw_hash);
    Entry *currently = table->buckets[index];

    while (currently) {
        if (strcmp(currently->key, key) == 0) {
            // printf("%s ==> %s\n", currently->key, currently->value);
            return currently->value; 
        }

        currently = currently->next;
    }

    return NULL;
}

void remove_item(HashMap *table, const char *key) {
    if (!table || !key) return;

    size_t raw_hash = get_hash_raw(key);
    size_t index = hash(table, raw_hash);
    Entry *currently = table->buckets[index];
    Entry *prev = NULL;

    while (currently) {
        if (strcmp(currently->key, key) == 0) {
            if (prev) {
                prev->next = currently->next;
            } else {
                table->buckets[index] = currently->next;
            }

            free(currently->key);
            free(currently->value);
            free(currently);

            table->count--;
            return;
        }

        prev = currently;
        currently = currently->next;
    }
}

void free_table(HashMap *table) {
    if (!table) return;

    for (size_t i = 0; i < table->size; i++) {
        Entry *currently = table->buckets[i];
        
        while (currently) {
            Entry *next = currently->next;

            free(currently->key);
            free(currently->value);
            free(currently);

            currently = next;
        }
    }

    free(table->buckets);
    free(table);
}
