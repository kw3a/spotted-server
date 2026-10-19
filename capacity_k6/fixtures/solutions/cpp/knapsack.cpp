#include <algorithm>
#include <iostream>
#include <vector>
using namespace std;
int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);
    int n, capacity;
    cin >> n >> capacity;
    vector<long long> best(capacity + 1, 0);
    for (int i = 0; i < n; ++i) {
        int weight;
        long long value;
        cin >> weight >> value;
        for (int limit = capacity; limit >= weight; --limit) {
            best[limit] = max(best[limit], best[limit - weight] + value);
        }
    }
    cout << best[capacity] << '\n';
    return 0;
}
