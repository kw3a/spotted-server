#include <iostream>
#include <queue>
#include <utility>
#include <vector>
using namespace std;
int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);
    int n, m;
    cin >> n >> m;
    vector<vector<pair<int, long long>>> graph(n);
    for (int i = 0; i < m; ++i) {
        int u, v;
        long long weight;
        cin >> u >> v >> weight;
        graph[u - 1].push_back({v - 1, weight});
    }
    vector<long long> distance(n, -1);
    queue<int> queue;
    distance[0] = 0;
    queue.push(0);
    while (!queue.empty()) {
        int u = queue.front();
        queue.pop();
        for (auto edge : graph[u]) {
            int v = edge.first;
            if (distance[v] == -1) {
                distance[v] = distance[u] + edge.second;
                queue.push(v);
            }
        }
    }
    cout << distance[n - 1] << '\n';
    return 0;
}
