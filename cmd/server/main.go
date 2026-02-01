package main

import (
	"fmt"
	"log"
	"nitrokv/internal/network"
	"os"
	"strconv"
)

func setTxt(key, value int) {
	f, err := os.OpenFile("test.txt", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	_, err = f.WriteString("SET " + strconv.Itoa(key) + " " + strconv.Itoa(value) + "\n")
}

func main() {
	// f, err := os.OpenFile("test.txt", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer f.Close()

	// var key int
	// var value int
	// for count := 0; count <= 10; count++ {
	// 	setTxt(key, value)
	// 	key++
	// 	value++
	// }

	// _, err = f.Seek(0, 0)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// r := bufio.NewReader(f)

	// for {
	// 	line, err := r.ReadString('\n')
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		fmt.Println(err)
	// 		return
	// 	}

	// 	line = strings.TrimSpace(line)

	// 	if line == "" {
	// 		continue
	// 	}
	// 	if line == "end" {
	// 		break
	// 	}

	// 	fmt.Println(line)
	// }
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
