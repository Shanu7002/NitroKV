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

int main() {
    test_basic_crud();

    printf("\nALL TESTS PASSED SUCESSFULLY\n");
    return 0;
}