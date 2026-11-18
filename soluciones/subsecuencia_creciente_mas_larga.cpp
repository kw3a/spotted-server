#include <iostream>
#include <vector>

using namespace std;

struct SequenceInput {
    int length;
    vector<long long> values;
};

class IncreasingSubsequenceLengths {
public:
    explicit IncreasingSubsequenceLengths(size_t maximumLength) {
        minimumEndingValue_.reserve(maximumLength);
    }

    void include(long long value) {
        const size_t position = firstPositionNotLessThan(value);

        if (position == minimumEndingValue_.size()) {
            minimumEndingValue_.push_back(value);
        } else {
            minimumEndingValue_[position] = value;
        }
    }

    size_t longestLength() const {
        return minimumEndingValue_.size();
    }

private:
    vector<long long> minimumEndingValue_;

    size_t firstPositionNotLessThan(long long value) const {
        size_t beginning = 0;
        size_t ending = minimumEndingValue_.size();

        while (beginning < ending) {
            const size_t middle = beginning + (ending - beginning) / 2;

            if (minimumEndingValue_[middle] < value) {
                beginning = middle + 1;
            } else {
                ending = middle;
            }
        }

        return beginning;
    }
};

bool readSequence(SequenceInput& input) {
    if (!(cin >> input.length)) {
        return false;
    }

    input.values.resize(static_cast<size_t>(input.length));
    for (int index = 0; index < input.length; ++index) {
        if (!(cin >> input.values[index])) {
            return false;
        }
    }
    return true;
}

size_t findLongestStrictlyIncreasingLength(const SequenceInput& input) {
    IncreasingSubsequenceLengths subsequences(input.values.size());

    for (size_t index = 0; index < input.values.size(); ++index) {
        subsequences.include(input.values[index]);
    }

    return subsequences.longestLength();
}

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    SequenceInput input;
    if (!readSequence(input)) {
        return 0;
    }

    const size_t answer = findLongestStrictlyIncreasingLength(input);
    cout << answer << '\n';
    return 0;
}
