#include <algorithm>
#include <iostream>
#include <vector>
using namespace std;
int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);
    int n;
    cin >> n;
    vector<long long> tails;
    for (int i = 0; i < n; ++i) {
        long long value;
        cin >> value;
        auto position = lower_bound(tails.begin(), tails.end(), value);
        if (position == tails.end()) tails.push_back(value);
        else *position = value;
    }
    cout << tails.size() << '\n';
    return 0;
}
