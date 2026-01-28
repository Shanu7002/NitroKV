#include <stdio.h>
#include <string.h>
#include <assert.h>
#include "../../internal/storage/hash_map.h"

void test_basic_crud() {
    printf("Running test_basic_crud\n");
    HashMap *map = create_table(16);

    set_item(map, "key1", "value1");
    set_item(map, "key2", "value2");

    assert(strcmp(get_item(map, "key1"), "value1") == 0);
    assert(strcmp(get_item(map, "key2"), "value2") == 0);
    assert(map->count == 2);

    set_item(map, "key1", "value3");
    assert(strcmp(get_item(map, "key1"), "value3") == 0);
    assert(map->count == 2);

    free_table(map);
    printf("test_basic_crud passed!\n");
}

void test_removal() {
    printf("Running test_removal\n");
    HashMap *map = create_table(16);

    set_item(map, "key1", "value1");
    set_item(map, "key2", "value2");

    assert(strcmp(get_item(map, "key1"), "value1") == 0);
    assert(strcmp(get_item(map, "key2"), "value2") == 0);

    remove_item(map, "key1");
    remove_item(map, "key2");

    assert(get_item(map, "key1") == NULL);
    assert(get_item(map, "key2") == NULL);

    assert(map->count == 0);

    free_table(map);
    printf("test_removal passed!\n");
}

void test_resize_integrity() {
    printf("Running test_resize\n");
    HashMap *map = create_table(4);

    for (int i = 0; i < 100; i++) {
        char key[20], val[20];
        sprintf(key, "k%d", i);
        sprintf(val, "v%d", i);
        set_item(map, key, val);
    }

    assert(map->size > 4);
    assert(map->count == 100);

    for (int i = 0; i < 100; i++) {
        char key[20], expected_val[20];
        sprintf(key, "k%d", i);
        sprintf(expected_val, "v%d", i);
        assert(strcmp(get_item(map, key), expected_val) == 0);
    }

    free_table(map);
    printf("test_resize_integrity passed!\n");
}

void test_collisions() {
    printf("Running test_collisions\n");
    HashMap *map = create_table(1);

    set_item(map, "key_a", "value_a");
    set_item(map, "key_b", "value_b");
    set_item(map, "key_c", "value_c");

    assert(map->count == 3);
    assert(strcmp(get_item(map, "key_a"), "value_a") == 0);
    assert(strcmp(get_item(map, "key_b"), "value_b") == 0);
    assert(strcmp(get_item(map, "key_c"), "value_c") == 0);

    remove_item(map, "key_b");
    assert(get_item(map, "key_b") == NULL);
    assert(strcmp(get_item(map, "key_a"), "value_a") == 0);
    assert(strcmp(get_item(map, "key_c"), "value_c") == 0);

    free_table(map);
    printf("test_collisions passed!\n");
}

int main() {
    printf("-------------------------\n");
    test_basic_crud();
    printf("-------------------------\n");
    test_removal();
    printf("-------------------------\n");
    test_resize_integrity();
    printf("-------------------------\n");
    test_collisions();

    printf("\nALL TESTS PASSED SUCESSFULLY\n");
    return 0;
}