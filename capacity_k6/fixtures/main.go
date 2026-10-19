// CLI de precondiciones y ciclo de vida; no agrega endpoints al servidor.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func execute(args []string) error {
	if len(args) > 0 && args[0] == "run" {
		return executeRun(args[1:])
	}
	if len(args) == 0 {
		return fmt.Errorf("uso: ./maintenance.sh {seed|cleanup|verify-judge0} [flags]")
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	manifestPath := flags.String("manifest", "fixtures.json", "archivo de precondiciones")
	envFile := flags.String("env-file", ".env", "dotenv con DATABASE_URL y BASE_URL")
	connectTimeout := flags.Duration("connect-timeout", 10*time.Second, "limite de conexion a MySQL")
	ioTimeout := flags.Duration("db-io-timeout", 10*time.Second, "limite de cada lectura/escritura MySQL")
	operationTimeout := flags.Duration("operation-timeout", 5*time.Minute, "limite total de seed o cleanup")
	id := flags.String("run-id", "", "identificador de la ejecucion a limpiar")
	judgeURL := flags.String("judge-url", "http://127.0.0.1:2358", "API de Judge0 para verificacion directa")
	judgeConfig := flags.String("judge-config", "../judge0/judge0.conf", "config de autenticacion de Judge0; vacio para API sin autenticacion")
	judgeReport := flags.String("report", "judge0.verification.json", "informe de verificacion directa")
	refresh := flags.Bool("refresh-problems", false, "actualizar problemas y casos propios sin participaciones; solo seed")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("argumentos inesperados")
	}
	if args[0] != "seed" && args[0] != "run" && args[0] != "cleanup" && args[0] != "verify-judge0" {
		return fmt.Errorf("comando desconocido: %s", args[0])
	}
	if *refresh && args[0] != "seed" {
		return fmt.Errorf("--refresh-problems solo se admite con seed")
	}
	if args[0] == "verify-judge0" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		return verifyJudge(ctx, *judgeURL, *judgeConfig, *judgeReport)
	}
	connections, err := readConnections(*envFile)
	if err != nil {
		return err
	}
	if *connectTimeout <= 0 || *ioTimeout <= 0 || *operationTimeout <= 0 {
		return fmt.Errorf("--connect-timeout, --db-io-timeout y --operation-timeout deben ser positivos")
	}
	if connections.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL falta en %s", *envFile)
	}
	db, err := openDatabase(connections.DatabaseURL, *connectTimeout, *ioTimeout)
	if err != nil {
		return err
	}
	defer db.Close()
	base, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pingCtx, pingCancel := context.WithTimeout(base, *connectTimeout)
	fmt.Fprintln(os.Stderr, "Conectando a MySQL usando DATABASE_URL del archivo dotenv...")
	err = db.PingContext(pingCtx)
	pingCancel()
	if err != nil {
		return fmt.Errorf("conexion a MySQL fallida (limite %s): %w", *connectTimeout, err)
	}
	fmt.Fprintln(os.Stderr, "Conexion a MySQL establecida.")
	ctx, cancel := context.WithTimeout(base, *operationTimeout)
	defer cancel()
	switch args[0] {
	case "seed":
		progress := func(format string, values ...any) { fmt.Fprintf(os.Stderr, format+"\n", values...) }
		progress("Creando o comprobando precondiciones (limite total %s)...", *operationTimeout)
		m, err := seedDataset(ctx, db, *refresh, progress)
		if err != nil {
			return fmt.Errorf("seed fallido: %w", err)
		}
		if err = writeJSON(*manifestPath, m); err != nil {
			return err
		}
		fmt.Printf("Precondiciones listas: 10 usuarios, 10 companias, 10 ofertas, 10 quizzes, 30 problemas, 60 ejemplos y 150 casos secretos. Manifest: %s\n", *manifestPath)
		return nil
	default:
		count, err := cleanup(ctx, db, *id)
		if err != nil {
			return err
		}
		fmt.Printf("Eliminados %d usuarios y sus datos de run %s; fixtures conservados.\n", count, *id)
		return nil
	}
}
func main() {
	if err := execute(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		var exit *exec.ExitError
		if errors.As(err, &exit) && exit.ExitCode() > 0 {
			os.Exit(exit.ExitCode())
		}
		os.Exit(1)
	}
}
