package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func runK6(manifestPath, runDir string, config RuntimeConfig, args []string) error {
	manifestPath, err := filepath.Abs(manifestPath)
	if err != nil {
		return err
	}
	if _, err = readManifest(manifestPath); err != nil {
		return err
	}

	entropy := make([]byte, 8)
	if _, err = rand.Read(entropy); err != nil {
		return err
	}
	id := hex.EncodeToString(entropy)
	prefix, _ := runPrefix(id)
	config.RunID, config.FixturesFile = id, manifestPath
	entry, release, err := createLaunch(config, ".bin", "main.js")
	if err != nil {
		return err
	}
	defer release()
	record := RunRecord{RunID: id, Prefix: prefix, StartedAt: time.Now().UTC(), Config: config}
	recordPath := filepath.Join(runDir, id+".json")
	if err = writeJSON(recordPath, record); err != nil {
		return err
	}
	envFile := config.EnvFile
	if envFile == "" {
		envFile = ".env"
	}
	fmt.Printf("Run %s; recuperacion: ./cleanup.sh --env-file %q --run-id %s\n", id, envFile, id)
	commandArgs := append([]string{"run"}, args...)
	commandArgs = append(commandArgs, entry)
	command := exec.Command("k6", commandArgs...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	// No exponer la conexion SQL a k6 ni heredar opciones K6_* del shell.
	for _, item := range os.Environ() {
		name := strings.SplitN(item, "=", 2)[0]
		if name != "DATABASE_URL" && !strings.HasPrefix(name, "K6_") {
			command.Env = append(command.Env, item)
		}
	}
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	finished := make(chan struct{})
	runErr := command.Start()
	if runErr == nil {
		go func() {
			select {
			case sig := <-signals:
				_ = command.Process.Signal(sig)
				select {
				case <-finished:
				case <-time.After(15 * time.Second):
					_ = command.Process.Kill()
				}
			case <-finished:
			}
		}()
		runErr = command.Wait()
	}
	close(finished)
	record.FinishedAt = time.Now().UTC()
	if err = writeJSON(recordPath, record); err != nil {
		return errors.Join(runErr, err)
	}
	fmt.Printf("Carga finalizada. Limpiar sus datos con: ./cleanup.sh --env-file %q --run-id %s\n", envFile, id)
	return runErr
}
