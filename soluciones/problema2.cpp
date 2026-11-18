#include <algorithm>
#include <functional>
#include <iostream>
#include <limits>
#include <queue>
#include <tuple>
#include <utility>
#include <vector>

using namespace std;

/*
 * Camino más barato entre 1 y n con permiso para usar un cupón en, como
 * máximo, una arista. El cupón reduce a la mitad el costo entero (floor).
 * El grafo no dirigido puede contener pesos grandes, por eso las distancias
 * y sumas usan long long.
 *
 * Dijkstra se amplía con una dimensión de estado: used=0 significa que el
 * cupón aún está disponible y used=1 que ya se aplicó. Desde cada estado se
 * puede recorrer una arista pagando su costo normal; desde used=0 también se
 * puede recorrerla pagando weight/2 y pasar a used=1. Esto equivale a buscar
 * en un grafo de 2n estados, con aristas normales dentro de cada capa y las
 * aristas con descuento entre capas.
 *
 * Complejidad: O((n + m) log n), con una constante pequeña por las dos capas.
 */
struct Edge {
    int to;
    long long weight;
};

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int n, m;
    if (!(cin >> n >> m)) return 0;

    vector<vector<Edge>> graph(n + 1);
    for (int i = 0; i < m; ++i) {
        int u, v;
        long long w;
        cin >> u >> v >> w;
        graph[u].push_back({v, w});
        graph[v].push_back({u, w});
    }

    const long long INF = numeric_limits<long long>::max() / 4;
    vector<vector<long long>> distance(2, vector<long long>(n + 1, INF));
    using State = tuple<long long, int, int>; // costo, nodo, cupón usado
    priority_queue<State, vector<State>, greater<State>> pending;

    distance[0][1] = 0;
    pending.push({0, 1, 0});

    while (!pending.empty()) {
        auto [cost, node, used] = pending.top();
        pending.pop();
        if (cost != distance[used][node]) continue; // entrada obsoleta

        for (const Edge& edge : graph[node]) {
            // Opción 1: cruzar la arista pagando el precio completo.
            long long normalCost = cost + edge.weight;
            if (normalCost < distance[used][edge.to]) {
                distance[used][edge.to] = normalCost;
                pending.push({normalCost, edge.to, used});
            }

            // Opción 2: gastar el cupón en esta arista. Nunca se vuelve a
            // ofrecer una vez que el estado ya indica used == 1.
            if (used == 0) {
                long long discountedCost = cost + edge.weight / 2;
                if (discountedCost < distance[1][edge.to]) {
                    distance[1][edge.to] = discountedCost;
                    pending.push({discountedCost, edge.to, 1});
                }
            }
        }
    }

    // El cupón es opcional: el mejor resultado puede estar en cualquiera
    // de las dos capas, por ejemplo si todas las rutas cortas empatan.
    long long answer = min(distance[0][n], distance[1][n]);
    if (answer == INF) cout << -1 << '\n';
    else cout << answer << '\n';
    return 0;
}
