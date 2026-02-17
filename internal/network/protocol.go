package network

import (
	"bufio"
	"fmt"
	"nitrokv/internal/engine"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

/*
	General helper

msg -> Message struct that have: 1- Who sent the message (from), 2- Actual message (payload), 3- to return message (Conn)
text -> Payload to string
parts -> text divided by fields. parts[0] always the cmd
cmd -> command. Set, Get, Remove etc
*/

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
	f, err := os.OpenFile("data/"+dbName+".log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Persistence Error for %s: %v\n", dbName, err)
		return
	}
	defer f.Close()

	switch cmd {
	case "SET":
		fmt.Fprintf(f, "SET \"%s\", %s\n", parts[1], parts[2])
	case "REMOVE":
		fmt.Fprintf(f, "REMOVE \"%s\"\n", parts[1])
	}

	f.Sync()
}

func (p *ProtocolManager) HandleHelp(msg Message) {
	helpMenu := `
--- NitroKV Help Menu ---
Commands:
  REGISTER <db>      : Creates a new isolated database.
  LOGIN <db>         : Enters a database session.
  SET "key", val     : Stores a value (supports quotes for spaces).
  GET "key"          : Retrieves a value.
  REMOVE "key"       : Deletes a key.
  RESTORE            : Reloads all databases from .log files.
  RESTORE <db>       : Reloads specific database from .log files.
  CLOSE              : Destroys current DB in memory.
  QUIT <db>          : Closes your connection.
  HELP               : Shows this menu.
-------------------------
`
	fmt.Fprint(msg.Conn, helpMenu)
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
		p.handleSet(msg, text, parts)
	case "GET":
		if len(parts) < 2 {
			fmt.Fprintf(msg.Conn, "ERR: GET requires a key.\n")
			return
		}
		res, status := p.handleGet(msg, text, parts)
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

		p.handleRemove(msg, text, parts)
	case "QUIT":
		if len(parts) < 2 {
			fmt.Fprintln(msg.Conn, "ERR: QUIT requires a database name.")
			return
		}
		p.handleQuit(msg, parts)
	case "CLOSE":
		p.handleClose(msg)
	case "RESTORE":
		if len(parts) == 1 {
			p.RestoreAll(msg)
			return
		}
		if len(parts) >= 2 {
			status := p.RestoreUnique(msg, parts)

			if status == false {
				fmt.Fprintf(msg.Conn, "ERR: DB not found!\n")
			}
		}
	case "HELP":
		p.HandleHelp(msg)
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

func (p *ProtocolManager) handleSet(msg Message, text string, parts []string) {
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

	// if the input text has this format -> SET "key key key", value value
	re := regexp.MustCompile(`(?i)^set\s+"([^"]+)",\s*(.+)$`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 3 {
		key := matches[1]
		value := strings.TrimSpace(matches[2])

		targetDB.Set(key, value)
		p.persist(dbName, "SET", matches)
		fmt.Fprintf(msg.Conn, "OK: key '%s' was set with value '%s' in '%s'\n", key, value, dbName)
		return
	}

	// if the input text has this format -> SET key value
	cmd := parts[0]
	key := parts[1]
	value := parts[2]

	targetDB.Set(key, value)
	p.persist(dbName, strings.ToUpper(cmd), parts)
	fmt.Fprintf(msg.Conn, "OK: key '%s' was set with value '%s' in '%s'\n", key, value, dbName)
}

func (p *ProtocolManager) handleGet(msg Message, text string, parts []string) (string, bool) {
	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintf(msg.Conn, "ERR: Not logged in. Use LOGIN <db_name>\n")
		return "login", false
	}

	var key string
	targetDB := p.dbs[dbName]

	// if the input text has this format -> GET "key key key"
	re := regexp.MustCompile(`(?i)^get\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(text)
	if len(matches) == 2 {
		key = matches[1]

		fmt.Printf(key)
		if res, ok := targetDB.Get(key); ok {
			return res, true
		}

	}
	// if the input text has this format -> GET key
	key = parts[1]

	if res, ok := targetDB.Get(key); ok {
		return res, true
	}
	return "", false
}

func (p *ProtocolManager) handleRemove(msg Message, text string, parts []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	dbName, loggedIn := p.sessions[msg.From]
	if !loggedIn {
		fmt.Fprintln(msg.Conn, "ERR: Not logged in.")
		return
	}

	targetDB, exists := p.dbs[dbName]
	if !exists {
		fmt.Fprintln(msg.Conn, "ERR: Database not found.")
		return
	}

	var key string
	re := regexp.MustCompile(`(?i)^remove\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(text)

	if len(matches) == 2 {
		key = matches[1]
	} else {
		key = parts[1]
	}

	if _, ok := targetDB.Get(key); !ok {
		fmt.Fprintf(msg.Conn, "ERR: Key '%s' not found!\n", key)
		return
	}

	targetDB.Remove(key)

	p.persist(dbName, "REMOVE", []string{"REMOVE", key})

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

func (p *ProtocolManager) RestoreAll(msg Message) error {
	files, err := filepath.Glob("data/*.log")
	if err != nil {
		return err
	}

	for _, filename := range files {
		dbName := strings.TrimPrefix(filename, "data/")
		dbName = strings.TrimSuffix(dbName, ".log")

		db, _ := engine.New(16)

		file, _ := os.Open(filename)
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			actualText := scanner.Text()
			parts := strings.Fields(actualText)
			if len(parts) < 2 {
				continue
			}
			cmd := strings.ToUpper(parts[0])

			switch cmd {
			case "SET":
				re := regexp.MustCompile(`(?i)^set\s+"([^"]+)",\s*(.+)$`)
				matches := re.FindStringSubmatch(actualText)
				key, value := matches[1], matches[2]

				db.Set(key, value)
			case "REMOVE":
				re := regexp.MustCompile(`(?i)^remove\s+"([^"]+)"`)
				matches := re.FindStringSubmatch(actualText)
				key := matches[1]

				db.Remove(key)
			}
		}
		file.Close()

		p.mu.Lock()
		p.dbs[dbName] = db
		p.mu.Unlock()

		fmt.Fprintf(msg.Conn, "Restored database: %s\n", dbName)
		fmt.Printf("Restored database: %s\n", dbName)
	}
	return nil
}

func (p *ProtocolManager) RestoreUnique(msg Message, parts []string) bool {
	files, err := filepath.Glob("data/*.log")
	if err != nil {
		return false
	}

	dbName := parts[1]
	for _, filename := range files {
		actuallDbName := strings.TrimPrefix(filename, "data/")
		actuallDbName = strings.TrimSuffix(actuallDbName, ".log")

		if actuallDbName == dbName {
			db, _ := engine.New(16)

			file, _ := os.Open(filename)
			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				actualText := scanner.Text()
				parts := strings.Fields(actualText)
				if len(parts) < 2 {
					continue
				}
				cmd := strings.ToUpper(parts[0])

				switch cmd {
				case "SET":
					re := regexp.MustCompile(`(?i)^set\s+"([^"]+)",\s*(.+)$`)
					matches := re.FindStringSubmatch(actualText)
					key, value := matches[1], matches[2]

					db.Set(key, value)
				case "REMOVE":
					re := regexp.MustCompile(`(?i)^remove\s+"([^"]+)"`)
					matches := re.FindStringSubmatch(actualText)
					key := matches[1]

					db.Remove(key)
				}
			}
			file.Close()

			p.mu.Lock()
			p.dbs[dbName] = db
			p.mu.Unlock()

			fmt.Fprintf(msg.Conn, "OK: Database '%s' success retored.\n", dbName)
			fmt.Printf("Restored database: %s\n", dbName)
			return true
		}
	}
	return false
}
