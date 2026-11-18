#include <algorithm>
#include <iostream>
#include <vector>

using namespace std;

/*
 * Encuentra la longitud de una subsecuencia estrictamente creciente más
 * larga, sin exigir que sus elementos sean contiguos.
 *
 * El método O(n^2) de programación dinámica prueba cada pareja de posiciones
 * y no alcanza para los límites grandes del problema. En su lugar, tails[k]
 * guarda el menor valor final posible entre las subsecuencias crecientes de
 * longitud k+1 procesadas hasta el momento. Mantener el final mínimo es útil
 * porque deja más margen para extender esa subsecuencia con valores futuros.
 *
 * Para cada x, lower_bound localiza la primera posición con tails[pos] >= x.
 * Si no existe, x extiende la subsecuencia más larga conocida. Si existe,
 * reemplazar ese final por x conserva la longitud y mejora su posibilidad de
 * extensión. Los reemplazos no necesariamente forman la subsecuencia real;
 * aun así, el invariante garantiza que tails.size() es la respuesta.
 *
 * Se usa lower_bound (y no upper_bound) para que la subsecuencia sea estricta:
 * dos valores iguales no aumentan la longitud.
 * Complejidad: O(n log n) tiempo y O(n) memoria.
 */
int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int n;
    if (!(cin >> n)) return 0;

    vector<long long> tails;
    tails.reserve(n);

    for (int i = 0; i < n; ++i) {
        long long value;
        cin >> value;

        auto place = lower_bound(tails.begin(), tails.end(), value);
        if (place == tails.end()) {
            // Este valor puede extender todas las subsecuencias conocidas.
            tails.push_back(value);
        } else {
            // Conserva la longitud, pero deja un final más pequeño para ella.
            *place = value;
        }
    }

    cout << tails.size() << '\n';
    return 0;
}
