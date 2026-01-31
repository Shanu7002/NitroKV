package main

import (
	"fmt"
	"log"
	"net"
	"nitrokv/internal/engine"
	"strings"
)

type Message struct {
	from    string
	payload []byte
	conn    net.Conn
}

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()

	s.ln = ln

	go s.acceptLoop()

	<-s.quitch
	close(s.msgch)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error:", err)
			continue
		}

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
		}

	}
}

func main() {
	server := NewServer(":6379")

	go func() {
		newMap, err := engine.New(16)
		if err != nil {
			server.ln.Close()
			// how can i return user an error without msg.conn?
		}
		for msg := range server.msgch {
			fmt.Printf("received massage from connection (%s): %s", msg.from, string(msg.payload))

			if len(msg.payload) == 0 {
				continue
			}
			text := strings.TrimSpace(string(msg.payload))

			parts := strings.Fields(text)
			if len(parts) == 0 {
				continue
			}

			command := strings.ToUpper(parts[0])
			switch command {
			case "SET":
				if len(parts) < 3 {
					msg.conn.Write([]byte("ERR: SET requires key and value.\n"))
					continue
				}
				key := parts[1]
				value := parts[2]
				newMap.Set(key, value)
				fmt.Printf("key '%s' was set with value '%s'\n", key, value)
			case "GET":
				if len(parts) < 2 {
					msg.conn.Write([]byte("ERR: GET requires a key.\n"))
					continue
				}
				key := parts[1]
				res, status := newMap.Get(key)
				if status == false {
					fmt.Println("Key not found!")
					continue
				}
				fmt.Printf("Value: %s\n", res)
			case "REMOVE":
				if len(parts) < 2 {
					msg.conn.Write([]byte("ERR: REMOVE requires a key"))
					continue
				}
				key := parts[1]
				newMap.Remove(key)
				fmt.Printf("key '%s' was sucessfully removed!\n", key)
			case "CLOSE":
				var answer string
				fmt.Printf("Do you really want to delete your map? you cannot recovery it (y/n) ")
				fmt.Scanln(&answer)
				switch answer {
				case "y":
					newMap.Close()
					fmt.Println("Map destroyed sucessfully!")
				case "n":
					continue
				default:
					fmt.Println("Wrong input, backing to hub.")
					continue
				}
			default:
				fmt.Println("case default")
			}
		}
	}()

	log.Fatal(server.Start())

	// kv, err := engine.New(1024)
	// if err != nil {
	// 	panic(err)
	// }
	// defer kv.Close()

	// kv.Set("status", "connected")
	// if val, ok := kv.Get("status"); ok {
	// 	fmt.Println("Value from C engine:", val)
	// }
}
