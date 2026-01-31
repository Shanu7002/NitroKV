package main

import (
	"fmt"
	"log"
	"nitrokv/internal/network"
)

func main() {
	server := network.NewServer(":6379")

	proto := network.NewProtocolManager()

	go func() {
		for msg := range server.Message() {
			fmt.Printf("received massage from connection (%s): %s", msg.From, string(msg.Payload))
			if len(msg.Payload) == 0 {
				continue
			}
			proto.HandleCommand(msg)
		}
	}()

	log.Fatal(server.Start())
}
