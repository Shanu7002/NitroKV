package network

import (
	"fmt"
	"nitrokv/internal/engine"
	"strings"
	"sync"
)

type ProtocolManager struct {
	dbs      map[string]*engine.Engine
	sessions map[string]string
	mu       sync.RWMutex
}

func NewProtocolManager() *ProtocolManager {
	return &ProtocolManager{
		dbs:      make(map[string]*engine.Engine),
		sessions: make(map[string]string),
	}
}

func (p *ProtocolManager) HandleCommand(msg Message) {
	text := strings.TrimSpace(string(msg.Payload))
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	command := strings.ToUpper(parts[0])
	switch command {
	case "REGISTER":
	case "LOGIN":
	case "SET":
		if len(parts) < 3 {
			fmt.Fprintln(msg.Conn, "ERR: SET requires key and value.")
			return
		}
		p.handleSet(msg, parts)

	case "GET":
		if len(parts) < 2 {
			fmt.Fprintf(msg.Conn, "ERR: GET requires a key.\n")
			return
		}
		res, status := p.handleGet(msg, parts)
		if status == false {
			fmt.Fprintf(msg.Conn, "ERR: Key not found!\n")
			return
		}
		fmt.Fprintf(msg.Conn, "OK: Value: %s\n", res)
	case "REMOVE":
		if len(parts) < 2 {
			fmt.Fprintf(msg.Conn, "ERR: REMOVE requires a key\n")
			return
		}

		if _, ok := p.handleGet(msg, parts); ok {
			p.handleRemove(msg, parts)
			// fmt.Fprintf(msg.Conn, "OK: key '%s' was sucessfully removed!\n", key)
		} else {
			fmt.Fprintf(msg.Conn, "ERR: Key not found!\n")
			return
		}
	case "CLOSE":
		fmt.Fprintf(msg.Conn, "OK: Server-wide map destroyed. Connection closing.\n")
		p.handleClose(msg)
		msg.Conn.Close()
	case "QUIT":
		fmt.Fprintf(msg.Conn, "Goodbye!\n")
		msg.Conn.Close()
	default:
		fmt.Println("Sorry, this function do not exist.")
	}
}

func (p *ProtocolManager) handleRegister(msg Message, parts string) {

}

func (p *ProtocolManager) handleLogin(msg Message, parts string) {

}

func (p *ProtocolManager) handleSet(msg Message, parts []string) {
	p.mu.RLock()
	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintf(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>")
		return
	}
	p.mu.RUnlock()

	targetDB := p.dbs[dbName]
	key, value := parts[1], parts[2]

	targetDB.Set(key, value)
	fmt.Fprintf(msg.Conn, "OK: key '%s' was set with value '%s' in '%s'\n", key, value, dbName)
}

func (p *ProtocolManager) handleGet(msg Message, parts []string) (string, bool) {
	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintf(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>")
		return "", false
	}

	targetDB := p.dbs[dbName]
	key := parts[1]

	if res, ok := targetDB.Get(key); ok {
		return res, true
	}
	return "", false
}

func (p *ProtocolManager) handleRemove(msg Message, parts []string) {

}

func (p *ProtocolManager) handleQuit(msg Message) {

}

func (p *ProtocolManager) handleClose(msg Message) {

}
