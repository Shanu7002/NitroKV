package network

import (
	"fmt"
	"net"
	"os"
	"testing"
)

type mockConn struct {
	net.Conn
}

func (m *mockConn) Write(b []byte) (int, error) { return len(b), nil }
func (m *mockConn) Close() error                { return nil }

func setupBench(b *testing.B) (*ProtocolManager, string, string) {
	tempDir, err := os.MkdirTemp("", "nitro_bench_*")
	if err != nil {
		b.Fatal(err)
	}

	oldWd, _ := os.Getwd()

	if err := os.Chdir(tempDir); err != nil {
		b.Fatal(err)
	}

	os.MkdirAll("data", 0755)

	pm := NewProtocolManager()
	dbName := "bench_db"

	pm.handleRegister(Message{Conn: &mockConn{}}, []string{"REGISTER", dbName})

	return pm, tempDir, oldWd
}

func BenchmarkHandleSet_Disk(b *testing.B) {
	pm, tempDir, oldWd := setupBench(b)

	defer os.Chdir(oldWd)
	defer os.RemoveAll(tempDir)

	msg := Message{
		From: "127.0.0.1:1",
		Conn: &mockConn{},
	}
	pm.sessions[msg.From] = "bench_db"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key_%d", i)
		pm.handleSet(msg, "SET "+key+" value", []string{"SET", key, "value"})
	}
}

func BenchmarkHandleSet_Parallel(b *testing.B) {
	pm, tempDir, oldWd := setupBench(b)
	defer os.Chdir(oldWd)
	defer os.RemoveAll(tempDir)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		addr := fmt.Sprintf("127.0.0.1:%d", os.Getpid())
		msg := Message{
			From: addr,
			Conn: &mockConn{},
		}

		pm.mu.Lock()
		pm.sessions[addr] = "bench_db"
		pm.mu.Unlock()

		for pb.Next() {
			pm.handleSet(msg, "SET pkey val", []string{"SET", "pkey", "val"})
		}
	})
}

func BenchmarkHandleGet_Memory(b *testing.B) {
	pm, tempDir, oldWd := setupBench(b)
	defer os.Chdir(oldWd)
	defer os.RemoveAll(tempDir)

	msg := Message{
		From: "127.0.0.1:1",
		Conn: &mockConn{},
	}
	pm.sessions[msg.From] = "bench_db"

	pm.handleSet(msg, "SET bench_key val", []string{"SET", "bench_key", "val"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.handleGet(msg, "GET bench_key", []string{"GET", "bench_key"})
	}
}
