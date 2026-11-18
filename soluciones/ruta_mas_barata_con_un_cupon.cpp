#include <algorithm>
#include <functional>
#include <iostream>
#include <limits>
#include <queue>
#include <vector>

using namespace std;

struct Road {
    int destination;
    long long price;

    Road(int destinationCity, long long roadPrice)
        : destination(destinationCity), price(roadPrice) {}
};

struct SearchState {
    long long accumulatedPrice;
    int city;
    int couponWasUsed;

    SearchState(long long price, int currentCity, int used)
        : accumulatedPrice(price), city(currentCity), couponWasUsed(used) {}
};

struct MoreExpensiveStateFirst {
    bool operator()(const SearchState& first,
                    const SearchState& second) const {
        if (first.accumulatedPrice != second.accumulatedPrice) {
            return first.accumulatedPrice > second.accumulatedPrice;
        }
        if (first.city != second.city) {
            return first.city > second.city;
        }
        return first.couponWasUsed > second.couponWasUsed;
    }
};

class RoadNetwork {
public:
    explicit RoadNetwork(int cityCount)
        : roads_(static_cast<size_t>(cityCount) + 1U) {}

    void addTwoWayRoad(int firstCity, int secondCity, long long price) {
        roads_[firstCity].push_back(Road(secondCity, price));
        roads_[secondCity].push_back(Road(firstCity, price));
    }

    int cityCount() const {
        return static_cast<int>(roads_.size()) - 1;
    }

    const vector<Road>& roadsLeaving(int city) const {
        return roads_[city];
    }

private:
    vector<vector<Road> > roads_;
};

class DiscountedRouteFinder {
public:
    explicit DiscountedRouteFinder(const RoadNetwork& network)
        : network_(network),
          infinity_(numeric_limits<long long>::max() / 4),
          bestPrice_(2, vector<long long>(
                           static_cast<size_t>(network.cityCount()) + 1U,
                           infinity_)) {}

    long long minimumPrice() {
        priority_queue<SearchState,
                        vector<SearchState>,
                        MoreExpensiveStateFirst> states;

        bestPrice_[0][1] = 0;
        states.push(SearchState(0, 1, 0));

        while (!states.empty()) {
            SearchState current = states.top();
            states.pop();

            if (isOutdated(current)) {
                continue;
            }

            exploreOutgoingRoads(current, states);
        }

        return cheapestDestinationPrice();
    }

private:
    const RoadNetwork& network_;
    const long long infinity_;
    vector<vector<long long> > bestPrice_;

    bool isOutdated(const SearchState& state) const {
        return state.accumulatedPrice !=
               bestPrice_[state.couponWasUsed][state.city];
    }

    void exploreOutgoingRoads(
        const SearchState& current,
        priority_queue<SearchState,
                       vector<SearchState>,
                       MoreExpensiveStateFirst>& states) {
        const vector<Road>& outgoing = network_.roadsLeaving(current.city);

        for (size_t index = 0; index < outgoing.size(); ++index) {
            const Road& road = outgoing[index];
            relaxWithoutDiscount(current, road, states);

            if (current.couponWasUsed == 0) {
                relaxWithDiscount(current, road, states);
            }
        }
    }

    void relaxWithoutDiscount(
        const SearchState& current,
        const Road& road,
        priority_queue<SearchState,
                       vector<SearchState>,
                       MoreExpensiveStateFirst>& states) {
        const long long candidatePrice =
            current.accumulatedPrice + road.price;
        const int destinationLayer = current.couponWasUsed;

        if (candidatePrice < bestPrice_[destinationLayer][road.destination]) {
            bestPrice_[destinationLayer][road.destination] = candidatePrice;
            states.push(SearchState(candidatePrice,
                                    road.destination,
                                    destinationLayer));
        }
    }

    void relaxWithDiscount(
        const SearchState& current,
        const Road& road,
        priority_queue<SearchState,
                       vector<SearchState>,
                       MoreExpensiveStateFirst>& states) {
        const long long discountedPrice = road.price / 2;
        const long long candidatePrice =
            current.accumulatedPrice + discountedPrice;

        if (candidatePrice < bestPrice_[1][road.destination]) {
            bestPrice_[1][road.destination] = candidatePrice;
            states.push(SearchState(candidatePrice, road.destination, 1));
        }
    }

    long long cheapestDestinationPrice() const {
        const int destination = network_.cityCount();
        const long long priceWithoutCoupon = bestPrice_[0][destination];
        const long long priceWithCoupon = bestPrice_[1][destination];
        return min(priceWithoutCoupon, priceWithCoupon);
    }
};

bool readNetwork(RoadNetwork& network, int roadCount) {
    for (int index = 0; index < roadCount; ++index) {
        int firstCity;
        int secondCity;
        long long price;

        if (!(cin >> firstCity >> secondCity >> price)) {
            return false;
        }
        network.addTwoWayRoad(firstCity, secondCity, price);
    }
    return true;
}

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int numberOfCities;
    int numberOfRoads;
    if (!(cin >> numberOfCities >> numberOfRoads)) {
        return 0;
    }

    RoadNetwork network(numberOfCities);
    if (!readNetwork(network, numberOfRoads)) {
        return 0;
    }

    DiscountedRouteFinder routeFinder(network);
    const long long answer = routeFinder.minimumPrice();
    const long long unreachable = numeric_limits<long long>::max() / 4;

    if (answer == unreachable) {
        cout << -1 << '\n';
    } else {
        cout << answer << '\n';
    }
    return 0;
}
