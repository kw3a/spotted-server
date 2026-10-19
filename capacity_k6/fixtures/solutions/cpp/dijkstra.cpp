#include <functional>
#include <iostream>
#include <limits>
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
    const long long infinity = numeric_limits<long long>::max() / 4;
    vector<long long> distance(n, infinity);
    priority_queue<pair<long long, int>, vector<pair<long long, int>>, greater<pair<long long, int>>> queue;
    distance[0] = 0;
    queue.push({0, 0});
    while (!queue.empty()) {
        auto current = queue.top();
        queue.pop();
        long long cost = current.first;
        int u = current.second;
        if (cost != distance[u]) continue;
        for (auto edge : graph[u]) {
            int v = edge.first;
            long long candidate = cost + edge.second;
            if (candidate < distance[v]) {
                distance[v] = candidate;
                queue.push({candidate, v});
            }
        }
    }
    cout << (distance[n - 1] == infinity ? -1 : distance[n - 1]) << '\n';
    return 0;
}
