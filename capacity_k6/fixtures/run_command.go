package main

import (
	"flag"
	"fmt"
)

// El lanzador de carga solo conoce dotenv, un manifest local y el proceso k6.
// No importa database/sql ni llama a seed/cleanup/validateDatabase.
func executeRun(args []string) error {
	flags := flag.NewFlagSet("run", flag.ContinueOnError)
	envFile := flags.String("env-file", ".env", "dotenv con BASE_URL; DATABASE_URL solo se usa en mantenimiento")
	manifestPath := flags.String("manifest", "fixtures.json", "archivo de precondiciones ya generado")
	runDir := flags.String("run-dir", "runs", "registros de ejecucion para limpieza posterior")
	var load runFlags
	load.register(flags)
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("argumentos inesperados")
	}
	connections, err := readConnections(*envFile)
	if err != nil {
		return err
	}
	load.resolve(flags)
	if err = load.settings.validate(connections.BaseURL); err != nil {
		return err
	}
	return runK6(*manifestPath, *runDir, RuntimeConfig{RunSettings: load.settings, BaseURL: connections.BaseURL, EnvFile: *envFile}, load.k6Args())
}
