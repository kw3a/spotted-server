#include <iostream>
#include <unordered_map>
#include <vector>

using namespace std;

class PrefixFrequencyTable {
public:
    explicit PrefixFrequencyTable(size_t expectedEntries) {
        frequencies_.max_load_factor(0.70F);
        frequencies_.reserve(expectedEntries);
        frequencies_[0] = 1;
    }

    long long occurrencesOf(long long prefixValue) const {
        unordered_map<long long, long long>::const_iterator found =
            frequencies_.find(prefixValue);
        if (found == frequencies_.end()) {
            return 0;
        }
        return found->second;
    }

    void record(long long prefixValue) {
        ++frequencies_[prefixValue];
    }

private:
    unordered_map<long long, long long> frequencies_;
};

struct SubarrayInput {
    int length;
    long long target;
    vector<long long> values;
};

bool readInput(SubarrayInput& input) {
    if (!(cin >> input.length >> input.target)) {
        return false;
    }

    input.values.resize(input.length);
    for (int index = 0; index < input.length; ++index) {
        if (!(cin >> input.values[index])) {
            return false;
        }
    }
    return true;
}

long long countMatchingSubarrays(const SubarrayInput& input) {
    PrefixFrequencyTable previousPrefixes(
        static_cast<size_t>(input.length) * 2U + 1U);

    long long currentPrefix = 0;
    long long matchingSubarrays = 0;

    for (size_t index = 0; index < input.values.size(); ++index) {
        currentPrefix += input.values[index];

        const long long neededPreviousPrefix = currentPrefix - input.target;
        matchingSubarrays +=
            previousPrefixes.occurrencesOf(neededPreviousPrefix);

        previousPrefixes.record(currentPrefix);
    }

    return matchingSubarrays;
}

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    SubarrayInput input;
    if (!readInput(input)) {
        return 0;
    }

    const long long answer = countMatchingSubarrays(input);
    cout << answer << '\n';
    return 0;
}
