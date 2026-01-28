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
	"sync"
	"unsafe"
)

type Engine struct {
	table *C.HashMap
	mu    sync.RWMutex
}

func New(size int) (*Engine, error) {
	table := C.create_table(C.size_t(size))
	if table == nil {
		return nil, errors.New("failed to allocate C hash map")
	}
	return &Engine{table: table}, nil
}

func (e *Engine) Set(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cKey := C.CString(key)
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cKey))
	defer C.free(unsafe.Pointer(cValue))

	C.set_item(e.table, cKey, cValue)
}

func (e *Engine) Get(key string) (string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	res := C.get_item(e.table, cKey)
	if res == nil {
		return "", false
	}
	return C.GoString(res), true
}

func (e *Engine) Close() {
	e.mu.Lock()
	if e.table != nil {
		C.free_table(e.table)
		e.table = nil
	}
}

func (e *Engine) Remove(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cKey := C.CString(key)
	defer C.free(unsafe.Pointer(cKey))

	C.remove_item(e.table, cKey)
}
