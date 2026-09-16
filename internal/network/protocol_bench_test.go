package network

import (
	"fmt"
	"net"
	"os"
	"testing"
)

// mockConn simula uma conexão de rede que não faz nada.
type mockConn struct {
	net.Conn
}

func (m *mockConn) Write(b []byte) (int, error) { return len(b), nil }
func (m *mockConn) Close() error                { return nil }

// setupBench cria um ambiente isolado no /tmp para não sujar ou apagar seu projeto.
func setupBench(b *testing.B) (*ProtocolManager, string, string) {
	// 1. Criamos a pasta temporária
	tempDir, err := os.MkdirTemp("", "nitro_bench_*")
	if err != nil {
		b.Fatal(err)
	}

	// 2. Guardamos onde estamos agora (seu repositório)
	oldWd, _ := os.Getwd()

	// 3. Entramos na pasta temporária
	if err := os.Chdir(tempDir); err != nil {
		b.Fatal(err)
	}

	// 4. Criamos a estrutura de dados necessária
	os.MkdirAll("data", 0755)

	pm := NewProtocolManager()
	dbName := "bench_db"

	// Registra o DB usando o mock
	pm.handleRegister(Message{Conn: &mockConn{}}, []string{"REGISTER", dbName})

	// Retornamos o PM, o diretório temporário para limpeza e o original para retorno
	return pm, tempDir, oldWd
}

// 1. TESTE DE ESCRITA SEQUENCIAL (Gargalo de Disco)
func BenchmarkHandleSet_Disk(b *testing.B) {
	pm, tempDir, oldWd := setupBench(b)

	// Limpeza: Volta para a pasta do projeto e deleta os logs de teste
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

// 2. TESTE DE CONCORRÊNCIA (Disputa de Mutex)
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

// 3. TESTE DE LEITURA EM MEMÓRIA (Performance do Motor C)
func BenchmarkHandleGet_Memory(b *testing.B) {
	pm, tempDir, oldWd := setupBench(b)
	defer os.Chdir(oldWd)
	defer os.RemoveAll(tempDir)

	msg := Message{
		From: "127.0.0.1:1",
		Conn: &mockConn{},
	}
	pm.sessions[msg.From] = "bench_db"

	// Preparação: Insere uma chave para ser lida repetidamente
	pm.handleSet(msg, "SET bench_key val", []string{"SET", "bench_key", "val"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.handleGet(msg, "GET bench_key", []string{"GET", "bench_key"})
	}
}
