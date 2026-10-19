package main

import (
	"fmt"
	"strings"
)

// Entradas grandes reproducibles, con respuestas conocidas por construccion.
func sequenceSample(values []int, answer int) Sample {
	var b strings.Builder
	fmt.Fprintf(&b, "%d\n", len(values))
	for i, v := range values {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprint(&b, v)
	}
	b.WriteByte('\n')
	return Sample{b.String(), fmt.Sprintf("%d\n", answer)}
}

func algorithms() []Algorithm {
	increasing := make([]int, 10000)
	duplicates := make([]int, 5000)
	for i := range increasing {
		increasing[i] = i - 5000
	}
	for i := range duplicates {
		duplicates[i] = i / 5
	}
	var bag strings.Builder
	fmt.Fprintln(&bag, "300 2000")
	for i := 1; i <= 300; i++ {
		fmt.Fprintf(&bag, "10 %d\n", i)
	}
	const vertices = 1500
	var graph strings.Builder
	fmt.Fprintf(&graph, "%d %d\n", vertices, 3*vertices-5)
	for step := 1; step <= 3; step++ {
		for i := 1; i+step <= vertices; i++ {
			weight := 7
			if step > 1 {
				weight = step * 10
			}
			fmt.Fprintf(&graph, "%d %d %d\n", i, i+step, weight)
		}
	}
	fmt.Fprintf(&graph, "1 %d 50000\n", vertices)
	lis := Algorithm{
		Key:                 "lis",
		Title:               "Subsecuencia creciente mas larga",
		Description:         "Dada una secuencia, calcula la longitud de su subsecuencia estrictamente creciente mas larga. Puedes omitir elementos, conservando su orden; valores iguales no forman un incremento. Entrada: n en la primera linea y n enteros en la segunda. Restricciones: 1 <= n <= 10000; -10^9 <= a[i] <= 10^9. Salida: una longitud entera y un salto de linea. Se espera O(n log n).",
		PartialAccepted:     2,
		PartialCaseAccepted: [5]bool{true, false, false, true, false},
		Examples:            []Sample{{"8\n10 9 2 5 3 7 101 18\n", "4\n"}, {"6\n-3 -1 -2 0 2 1\n", "4\n"}},
		Cases: []Sample{
			{"6\n9 7 5 3 1 -1\n", "1\n"},
			{"8\n7 7 7 7 7 7 7 7\n", "1\n"},
			{"8\n3 1 2 2 4 3 5 5\n", "4\n"},
			sequenceSample(increasing, 10000), sequenceSample(duplicates, 1000),
		},
	}
	knapsack := Algorithm{
		Key:                 "knapsack",
		Title:               "Mochila 0/1",
		Description:         "Selecciona objetos para maximizar su valor total sin superar la capacidad W. Cada objeto puede elegirse como maximo una vez y no se puede fraccionar. Entrada: n W, seguidos de n lineas peso valor. Restricciones: 1 <= n <= 300; 1 <= W <= 2000; 1 <= peso <= 10000; 1 <= valor <= 10000. Salida: el valor maximo como entero y un salto de linea. Se espera programacion dinamica O(n W), con memoria O(W).",
		PartialAccepted:     2,
		PartialCaseAccepted: [5]bool{true, false, false, true, false},
		Examples:            []Sample{{"4 7\n1 1\n3 4\n4 5\n5 7\n", "9\n"}, {"3 8\n3 30\n4 50\n5 60\n", "90\n"}},
		Cases: []Sample{
			{"3 5\n2 6\n3 10\n5 12\n", "16\n"},
			{"1 10\n3 5\n", "5\n"},
			{"2 6\n3 8\n4 9\n", "9\n"},
			{"2 4\n5 100\n6 200\n", "0\n"},
			{bag.String(), "40100\n"},
		},
	}
	shortest := Algorithm{
		Key:                 "dijkstra",
		Title:               "Camino minimo ponderado",
		Description:         "Calcula el coste minimo para viajar del vertice 1 al vertice n en un grafo dirigido con pesos no negativos. Puede haber ciclos, aristas paralelas y pesos cero. Entrada: n m, seguidos de m lineas u v w que representan una arista dirigida u -> v. Restricciones: 1 <= n <= 5000; 0 <= m <= 20000; 1 <= u,v <= n; 0 <= w <= 10^9. Salida: el coste minimo, o -1 si n es inalcanzable, seguido de un salto de linea. Se espera Dijkstra con cola de prioridad, O((n + m) log n).",
		PartialAccepted:     3,
		PartialCaseAccepted: [5]bool{true, false, true, false, true},
		Examples:            []Sample{{"4 5\n1 2 2\n1 3 10\n2 3 3\n3 4 1\n2 4 9\n", "6\n"}, {"3 1\n1 2 7\n", "-1\n"}},
		Cases: []Sample{
			{"4 4\n1 2 1\n1 3 1\n2 4 1\n3 4 1\n", "2\n"},
			{"3 3\n1 3 20\n1 2 2\n2 3 3\n", "5\n"},
			{"5 2\n1 2 7\n3 5 1\n", "-1\n"},
			{graph.String(), fmt.Sprintf("%d\n", 7*(vertices-1))},
			{"5 6\n1 2 0\n2 1 0\n2 3 4\n1 4 2\n4 3 3\n3 5 2\n", "6\n"},
		},
	}
	return []Algorithm{lis, knapsack, shortest}
}
