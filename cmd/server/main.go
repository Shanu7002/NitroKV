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
			return
		}

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
			conn:    conn,
		}

	}
}

func main() {
	server := NewServer(":6379")

	newMap, err := engine.New(16)
	if err != nil {
		log.Fatal()
	}

	go func() {
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
					fmt.Fprintln(msg.conn, "ERR: SET requires key and value.")
					continue
				}

				key, value := parts[1], parts[2]
				newMap.Set(key, value)
				fmt.Fprintf(msg.conn, "OK: key '%s' was set with value '%s'\n", key, value)
			case "GET":
				if len(parts) < 2 {
					fmt.Fprintf(msg.conn, "ERR: GET requires a key.\n")
					continue
				}
				key := parts[1]
				res, status := newMap.Get(key)
				if status == false {
					fmt.Fprintf(msg.conn, "ERR: Key not found!\n")
					continue
				}
				fmt.Fprintf(msg.conn, "OK: Value: %s\n", res)
			case "REMOVE":
				if len(parts) < 2 {
					fmt.Fprintf(msg.conn, "ERR: REMOVE requires a key")
					continue
				}
				key := parts[1]
				newMap.Remove(key)
				fmt.Fprintf(msg.conn, "OK: key '%s' was sucessfully removed!\n", key)
			case "CLOSE":
				fmt.Fprintf(msg.conn, "OK: Server-wide map destroyed. Connection closing.\n")
				newMap.Close()
			case "QUIT":
				fmt.Fprintf(msg.conn, "Goodbye!\n")
				msg.conn.Close()
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
