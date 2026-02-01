package network

import (
	"fmt"
	"nitrokv/internal/engine"
	"os"
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

func (p *ProtocolManager) persist(dbName, cmd string, parts []string) {
	if cmd != "SET" && cmd != "REMOVE" {
		return
	}
	f, err := os.OpenFile(dbName+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Persistence Error for %s: %v\n", dbName, err)
		return
	}
	defer f.Close()

	switch cmd {
	case "SET":
		fmt.Fprintf(f, "SET %s %s\n", parts[1], parts[2])
	case "REMOVE":
		fmt.Fprintf(f, "REMOVE %s\n", parts[1])
	}

	f.Sync()
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
		if len(parts) < 2 {
			fmt.Fprintln(msg.Conn, "ERR: REGISTER requires a name.")
			return
		}
		p.handleRegister(msg, parts)
	case "LOGIN":
		if len(parts) < 2 {
			fmt.Fprintln(msg.Conn, "ERR: LOGIN requires a name.")
			return
		}
		p.handleLogin(msg, parts)
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
		if res == "login" {
			return
		}
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
	case "QUIT":
		if len(parts) < 2 {
			fmt.Fprintln(msg.Conn, "ERR: QUIT requires a database name.")
			return
		}
		p.handleQuit(msg, parts)
	case "CLOSE":
		p.handleClose(msg)
	default:
		fmt.Println("Sorry, this function do not exist.")
	}
}

func (p *ProtocolManager) handleRegister(msg Message, parts []string) {
	dbName := parts[1]

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exist := p.dbs[dbName]; exist {
		fmt.Fprintf(msg.Conn, "Database '%s' is already taken\n", dbName)
		return
	}
	db, err := engine.New(16)
	if err != nil {
		fmt.Fprintln(msg.Conn, "ERR: Failed to create database")
		return
	}

	p.dbs[dbName] = db
	fmt.Fprintf(msg.Conn, "OK: Database '%s' registered\n", dbName)
}

func (p *ProtocolManager) handleLogin(msg Message, parts []string) {
	dbName := parts[1]
	connection := msg.From

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exist := p.dbs[dbName]; !exist {
		fmt.Fprintf(msg.Conn, "ERR: Database '%s' not found.\n", dbName)
		return
	}

	p.sessions[connection] = dbName
	fmt.Fprintf(msg.Conn, "OK: Using database '%s'\n", dbName)
}

func (p *ProtocolManager) handleSet(msg Message, parts []string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintln(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>")
		return
	}

	targetDB, exists := p.dbs[dbName]
	if !exists {
		fmt.Fprintf(msg.Conn, "ERR: Database %s no longer exists\n", dbName)
		return
	}

	cmd, key, value := parts[0], parts[1], parts[2]

	targetDB.Set(key, value)
	p.persist(dbName, strings.ToUpper(cmd), parts)
	fmt.Fprintf(msg.Conn, "OK: key '%s' was set with value '%s' in '%s'\n", key, value, dbName)
}

func (p *ProtocolManager) handleGet(msg Message, parts []string) (string, bool) {
	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintf(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>\n")
		return "login", false
	}

	targetDB := p.dbs[dbName]
	key := parts[1]

	if res, ok := targetDB.Get(key); ok {
		return res, true
	}
	return "", false
}

func (p *ProtocolManager) handleRemove(msg Message, parts []string) {
	key := parts[1]

	p.mu.Lock()
	defer p.mu.Unlock()

	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintln(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>.")
		return
	}
	targetDB := p.dbs[dbName]

	targetDB.Remove(key)
	fmt.Fprintf(msg.Conn, "OK: %s removed from %s\n", key, dbName)
}

func (p *ProtocolManager) handleQuit(msg Message, parts []string) {
	dbNameMsg := parts[1]
	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintln(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>")
		return
	}
	if dbNameMsg == dbName {
		fmt.Fprintf(msg.Conn, "OK: %s connection was closed.\n", dbName)
		fmt.Fprintf(msg.Conn, "Goodbye!\n")
		msg.Conn.Close()
		return
	}
	fmt.Fprintf(msg.Conn, "ERR: you are not logged in %s database\n", dbNameMsg)
}

func (p *ProtocolManager) handleClose(msg Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintln(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>")
		return
	}

	if targetDB, exists := p.dbs[dbName]; exists {
		targetDB.Close()
		delete(p.dbs, dbName)
	}

	for addr, name := range p.sessions {
		if name == dbName {
			delete(p.sessions, addr)
		}
	}

	fmt.Fprintf(msg.Conn, "OK: Server-wide database %s destroyed. Connection closing.\n", dbName)
	msg.Conn.Close()
}
