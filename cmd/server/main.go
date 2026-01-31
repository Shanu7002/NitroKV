package main

import (
	"fmt"
	"log"
	"nitrokv/internal/engine"
	"nitrokv/internal/network"
	"strings"
)

func main() {
	server := network.NewServer(":6379")

	newMap, err := engine.New(16)
	if err != nil {
		log.Fatal()
	}

	go func() {
		for msg := range server.Message() {
			fmt.Printf("received massage from connection (%s): %s", msg.From, string(msg.Payload))
			if len(msg.Payload) == 0 {
				continue
			}

			text := strings.TrimSpace(string(msg.Payload))
			parts := strings.Fields(text)
			if len(parts) == 0 {
				continue
			}

			command := strings.ToUpper(parts[0])
			switch command {
			case "SET":
				if len(parts) < 3 {
					fmt.Fprintln(msg.Conn, "ERR: SET requires key and value.")
					continue
				}

				key, value := parts[1], parts[2]
				newMap.Set(key, value)
				fmt.Fprintf(msg.Conn, "OK: key '%s' was set with value '%s'\n", key, value)
			case "GET":
				if len(parts) < 2 {
					fmt.Fprintf(msg.Conn, "ERR: GET requires a key.\n")
					continue
				}
				key := parts[1]
				res, status := newMap.Get(key)
				if status == false {
					fmt.Fprintf(msg.Conn, "ERR: Key not found!\n")
					continue
				}
				fmt.Fprintf(msg.Conn, "OK: Value: %s\n", res)
			case "REMOVE":
				if len(parts) < 2 {
					fmt.Fprintf(msg.Conn, "ERR: REMOVE requires a key\n")
					continue
				}
				key := parts[1]

				if _, ok := newMap.Get(key); ok {
					newMap.Remove(key)
					fmt.Fprintf(msg.Conn, "OK: key '%s' was sucessfully removed!\n", key)
				} else {
					fmt.Fprintf(msg.Conn, "ERR: Key not found!\n")
					continue
				}
			case "CLOSE":
				fmt.Fprintf(msg.Conn, "OK: Server-wide map destroyed. Connection closing.\n")
				newMap.Close()
				msg.Conn.Close()
			case "QUIT":
				fmt.Fprintf(msg.Conn, "Goodbye!\n")
				msg.Conn.Close()
			default:
				fmt.Println("Sorry, this function do not exist.")
			}
		}
	}()

	log.Fatal(server.Start())
}
