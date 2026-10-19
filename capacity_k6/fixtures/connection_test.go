package main

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Transporte en memoria: acepta la conexion pero nunca envia el saludo MySQL.
// No necesita sockets, servicios ni una base de datos real.
func stalledDSN(t *testing.T) string {
	t.Helper()
	mysql.RegisterDialContext("capacity-stalled", func(context.Context, string) (net.Conn, error) {
		client, server := net.Pipe()
		// Respaldo independiente para que una regresion tampoco cuelgue el test.
		timer := time.AfterFunc(3*time.Second, func() { server.Close() })
		t.Cleanup(func() { timer.Stop(); client.Close(); server.Close() })
		return client, nil
	})
	return "user:secret@capacity-stalled(test)/db"
}

func TestDatabaseHandshakeTimeout(t *testing.T) {
	db, err := openDatabase(stalledDSN(t), 80*time.Millisecond, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	started := time.Now()
	err = db.PingContext(context.Background())
	if err == nil {
		t.Fatal("accepted server without a MySQL greeting")
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("handshake exceeded timeout: %s", elapsed)
	}
}

func TestSeedReportsConnectionFailureAndExits(t *testing.T) {
	dir := t.TempDir()
	env := filepath.Join(dir, ".env")
	if err := os.WriteFile(env, []byte("DATABASE_URL="+stalledDSN(t)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output, err := os.CreateTemp(dir, "stderr")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	original := os.Stderr
	os.Stderr = output
	defer func() { os.Stderr = original }()
	manifestPath := filepath.Join(dir, "fixtures.json")
	started := time.Now()
	err = execute([]string{"seed", "--env-file", env, "--manifest", manifestPath, "--connect-timeout", "80ms", "--db-io-timeout", "1m"})
	if err == nil || !strings.Contains(err.Error(), "conexion a MySQL fallida") {
		t.Fatalf("expected connection error: %v", err)
	}
	if time.Since(started) > 2*time.Second {
		t.Fatal("seed did not exit promptly")
	}
	text, readErr := os.ReadFile(output.Name())
	if readErr != nil || !strings.Contains(string(text), "Conectando a MySQL") || strings.Contains(string(text), "secret") {
		t.Fatalf("missing or unsafe progress: %s (%v)", text, readErr)
	}
	if _, err = os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Fatal("failed seed wrote a manifest")
	}
}

func TestDatabaseInvalidDSNDoesNotExposeCredentials(t *testing.T) {
	_, err := openDatabase("password-secret-invalid-dsn", time.Second, time.Second)
	if err == nil || err.Error() != "DATABASE_URL no tiene formato DSN MySQL valido" {
		t.Fatalf("unexpected error: %v", err)
	}
}
