package main

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed solutions/cpp/*.cpp solutions/java/*/Main.java
var solutionFiles embed.FS

type Programs struct{ Source, PartialSource, SyntaxErrorSource string }

func programsFor(key, languageID string) Programs {
	path := func(name string) string {
		if languageID == "54" {
			return "solutions/cpp/" + name + ".cpp"
		}
		if languageID == "62" {
			return "solutions/java/" + name + "/Main.java"
		}
		panic("unsupported solution language: " + languageID)
	}
	read := func(name string) string {
		data, err := solutionFiles.ReadFile(path(name))
		if err != nil {
			panic(err)
		}
		return string(data)
	}
	valid := read(key)
	partial := valid
	switch key {
	case "lis":
		if languageID == "54" {
			partial = strings.Replace(valid, "lower_bound", "upper_bound", 1)
		} else {
			partial = strings.Replace(valid, "tails[mid] < value", "tails[mid] <= value", 1)
		}
	case "knapsack":
		partial = strings.Replace(valid, "int limit = capacity; limit >= weight; --limit", "int limit = weight; limit <= capacity; ++limit", 1)
	case "dijkstra":
		partial = read("bfs")
	default:
		panic(fmt.Sprintf("unsupported algorithm: %s", key))
	}
	// Retirar el punto y coma de una instruccion garantiza error de sintaxis.
	syntax := strings.Replace(valid, "return 0;", "return 0", 1)
	if languageID == "62" {
		syntax = strings.Replace(valid, "int position = 0, length = 0;", "int position = 0, length = 0", 1)
	}
	return Programs{valid, partial, syntax}
}
