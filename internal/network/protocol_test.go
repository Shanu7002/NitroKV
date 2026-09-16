package network

import (
	"bufio"
	"net"
	"os"
	"strings"
	"testing"
)

func exchange(pm *ProtocolManager, cmd string, from string) string {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	reader := bufio.NewReader(clientConn)
	msg := Message{
		Payload: []byte(cmd + "\n"),
		Conn:    serverConn,
		From:    from,
	}

	go func() {
		pm.HandleCommand(msg)
		serverConn.Close()
	}()

	response, _ := reader.ReadString('\n')
	return strings.TrimSpace(response)
}

func TestCompleteWorkflow(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)
	os.Mkdir("data", 0755)

	pm := NewProtocolManager()
	userA := "127.0.0.1:1001"
	userB := "127.0.0.1:1002"

	res := exchange(pm, "REGISTER prod_db", userA)
	if !strings.Contains(res, "OK") {
		t.Errorf("Registration failed: %s", res)
	}

	exchange(pm, "LOGIN prod_db", userA)
	res = exchange(pm, "SET key1 value1", userA)
	if !strings.Contains(res, "OK") {
		t.Errorf("Set failed: %s", res)
	}

	res = exchange(pm, "SET key2 value2", userB)
	if !strings.Contains(res, "ERR: Not logged in") {
		t.Errorf("Security breach! UserB set data without login: %s", res)
	}

	exchange(pm, "LOGIN prod_db", userA)
	res = exchange(pm, `SET "User Name", "Senior Engineer"`, userA)
	if !strings.Contains(res, "OK") {
		t.Errorf("Regex SET failed: %s", res)
	}

	res = exchange(pm, `GET "User Name"`, userA)
	if !strings.Contains(res, "Senior Engineer") {
		t.Errorf("Regex GET failed, expected 'Senior Engineer', got: %s", res)
	}

	res = exchange(pm, `REMOVE "User Name"`, userA)
	if !strings.Contains(res, "removed") {
		t.Errorf("Remove failed: %s", res)
	}

	if _, err := os.Stat("data/prod_db.log"); os.IsNotExist(err) {
		t.Fatal("Persistence error: prod_db.log was not created")
	}
}

func TestRecovery(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)
	os.Mkdir("data", 0755)

	logContent := "SET \"restored_key\", value_recovered\n"
	os.WriteFile("data/old_db.log", []byte(logContent), 0644)

	pm := NewProtocolManager()

	pm.RestoreAll(Message{Conn: &net.TCPConn{}})

	user := "127.0.0.1:9999"
	exchange(pm, "LOGIN old_db", user)
	res := exchange(pm, `GET "restored_key"`, user)

	if !strings.Contains(res, "value_recovered") {
		t.Errorf("Recovery failed. Expected 'value_recovered', got: %s", res)
	}
}
