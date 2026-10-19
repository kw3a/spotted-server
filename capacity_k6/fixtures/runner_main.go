//go:build capacity_runner

// Este archivo se compila explicitamente con los archivos del lanzador HTTP.
// El ejecutable resultante no incluye seeders, SQL ni la limpieza.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "run" {
		args = args[1:]
	}
	if err := executeRun(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		var child *exec.ExitError
		if errors.As(err, &child) && child.ExitCode() > 0 {
			os.Exit(child.ExitCode())
		}
		os.Exit(1)
	}
}
