package network

import (
	"bufio"
	"net"
	"os"
	"strings"
	"testing"
)

func TestRegisterAndLogin(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	os.Mkdir("data", 0755)

	pm := NewProtocolManager()

	tests := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "Register new database",
			command:  "REGISTER test_db\n",
			expected: "OK: Database 'test_db' registered",
		},
		{
			name:     "Register duplicate database",
			command:  "REGISTER test_db\n",
			expected: "Database 'test_db' is already taken",
		},
		{
			name:     "Login to database",
			command:  "LOGIN test_db\n",
			expected: "OK: Using database 'test_db'",
		},
		{
			name:     "Login to non-existent database",
			command:  "LOGIN unknown_db\n",
			expected: "ERR: Database 'unknown_db' not found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientConn, serverConn := net.Pipe()
			defer clientConn.Close()
			defer serverConn.Close()

			reader := bufio.NewReader(clientConn)

			msg := Message{
				Payload: []byte(tt.command),
				Conn:    serverConn,
				From:    "127.0.0.1:1234",
			}

			go func() {
				pm.HandleCommand(msg)
				serverConn.Close()
			}()

			response, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" && !strings.Contains(err.Error(), "closed") {
				t.Fatalf("Failed to read response: %v", err)
			}

			actual := strings.TrimSpace(response)
			if actual != tt.expected {
				t.Errorf("Expected: %q, Got: %q", tt.expected, actual)
			}
		})
	}
}
