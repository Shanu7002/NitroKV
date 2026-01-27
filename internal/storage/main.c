#include <stdio.h>
#include <assert.h>
#include "hash_map.h"

int main() {
    HashMap *map = create_table(10); 

    printf("Inserting items...\n");
    set_item(map, "key1", "value1");
    set_item(map, "key2", "value2");
    
    set_item(map, "key1", "new_value1"); 
    
    for (int i = 0; i < 100; i++) {
        char k[20], v[20];
        sprintf(k, "stress_%d", i);
        sprintf(v, "val_%d", i);
        set_item(map, k, v);
    }

    const char *res = get_item(map, "key1");
    const char *res2 = get_item(map, "key2");
    const char *res3 = get_item(map, "key3");
    printf("Value for key1: %s\n", res);
    printf("Value for key2: %s\n", res2);
    printf("Value for invalid key: %s\n", res3);

    remove_item(map, "key2");
    if (get_item(map, "key2") == NULL) {
        printf("Successfully removed key2\n");
    }

    free_table(map);
    printf("Table free\n");

    return 0;
}