#include <iostream>
#include <vector>

using namespace std;

struct PointAssignment {
    int position;
    long long newValue;
};

struct RangeRequest {
    int firstPosition;
    int lastPosition;
};

class FenwickTree {
public:
    explicit FenwickTree(int elementCount)
        : tree_(static_cast<size_t>(elementCount) + 1U, 0) {}

    int size() const {
        return static_cast<int>(tree_.size()) - 1;
    }

    void addValueAt(int position, long long difference) {
        int currentIndex = position;
        const int elementCount = size();

        while (currentIndex <= elementCount) {
            tree_[currentIndex] += difference;
            currentIndex += lowestSetBit(currentIndex);
        }
    }

    long long sumThrough(int position) const {
        long long total = 0;
        int currentIndex = position;

        while (currentIndex > 0) {
            total += tree_[currentIndex];
            currentIndex -= lowestSetBit(currentIndex);
        }
        return total;
    }

    long long sumBetween(int firstPosition, int lastPosition) const {
        const long long sumBeforeFirst = sumThrough(firstPosition - 1);
        const long long sumThroughLast = sumThrough(lastPosition);
        return sumThroughLast - sumBeforeFirst;
    }

private:
    vector<long long> tree_;

    static int lowestSetBit(int value) {
        return value & -value;
    }
};

class MutableNumberArray {
public:
    explicit MutableNumberArray(const vector<long long>& initialValues)
        : values_(initialValues), fenwick_(
              static_cast<int>(initialValues.size()) - 1) {}

    void buildIndex() {
        for (int position = 1; position < static_cast<int>(values_.size());
             ++position) {
            fenwick_.addValueAt(position, values_[position]);
        }
    }

    void assign(int position, long long newValue) {
        const long long oldValue = values_[position];
        const long long change = newValue - oldValue;

        values_[position] = newValue;
        fenwick_.addValueAt(position, change);
    }

    long long rangeSum(int firstPosition, int lastPosition) const {
        return fenwick_.sumBetween(firstPosition, lastPosition);
    }

private:
    vector<long long> values_;
    FenwickTree fenwick_;
};

bool readInitialValues(int elementCount, vector<long long>& values) {
    values.assign(static_cast<size_t>(elementCount) + 1U, 0);

    for (int position = 1; position <= elementCount; ++position) {
        if (!(cin >> values[position])) {
            return false;
        }
    }
    return true;
}

void executeAssignment(MutableNumberArray& numbers) {
    PointAssignment assignment;
    cin >> assignment.position >> assignment.newValue;
    numbers.assign(assignment.position, assignment.newValue);
}

void executeRangeRequest(const MutableNumberArray& numbers) {
    RangeRequest request;
    cin >> request.firstPosition >> request.lastPosition;
    cout << numbers.rangeSum(request.firstPosition,
                             request.lastPosition)
         << '\n';
}

void executeOperations(MutableNumberArray& numbers, int operationCount) {
    for (int operationIndex = 0; operationIndex < operationCount;
         ++operationIndex) {
        int operationType;
        cin >> operationType;

        if (operationType == 1) {
            executeAssignment(numbers);
        } else if (operationType == 2) {
            executeRangeRequest(numbers);
        }
    }
}

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int elementCount;
    int operationCount;
    if (!(cin >> elementCount >> operationCount)) {
        return 0;
    }

    vector<long long> initialValues;
    if (!readInitialValues(elementCount, initialValues)) {
        return 0;
    }

    MutableNumberArray numbers(initialValues);
    numbers.buildIndex();
    executeOperations(numbers, operationCount);
    return 0;
}
