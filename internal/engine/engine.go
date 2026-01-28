package engine

/*
#cgo CFLAGS: -I../storage
#include "../storage/hash_map.h"
#include "../storage/hash_map.c"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
)

type Engine struct {
	table *C.HashMap
}

func New(size int) (*Engine, error) {
	table := C.create_table(C.size_t(size))
	if table == nil {
		return nil, errors.New("failed to allocate C hash map")
	}
	return &Engine{table: table}, nil
}

func (e *Engine) Close() {
	if e.table != nil {
		C.free_table(e.table)
		e.table = nil
	}
}
