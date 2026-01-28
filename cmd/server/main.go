package main

import (
	"nitrokv/internal/engine"
)

func main() {
	kv, err := engine.New(1024)
	if err != nil {
		panic(err)
	}
	defer kv.Close()
}
