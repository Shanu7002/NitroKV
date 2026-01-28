package main

import (
	"fmt"
	"nitrokv/internal/engine"
)

func main() {
	kv, err := engine.New(1024)
	if err != nil {
		panic(err)
	}
	defer kv.Close()

	kv.Set("status", "connected")
	if val, ok := kv.Get("status"); ok {
		fmt.Println("Value from C engine:", val)
	}
}
