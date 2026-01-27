nitro_test: internal/storage/hash_map.c internal/storage/main.c
	$(CC) $(CFLAGS) $^ -o $@

remove: 
	rm -r nitro_test

compile:	nitro_test

run: 
	./nitro_test

clear:	remove
