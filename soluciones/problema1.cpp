#include <iostream>
#include <unordered_map>
#include <vector>

using namespace std;

/*
 * Cuenta los subarreglos cuya suma es exactamente K.
 *
 * La presencia de números negativos hace que una ventana deslizante normal
 * no sea suficiente: al avanzar el extremo derecho, la suma puede aumentar
 * o disminuir. En su lugar usamos sumas prefijas.
 *
 * Si P[j] es la suma de los primeros j elementos, el segmento que empieza
 * después del prefijo i y termina en j suma K cuando P[j] - P[i] = K.
 * Por tanto, para cada P[j] basta contar cuántos prefijos anteriores valen
 * P[j] - K. El mapa guarda exactamente esos conteos.
 *
 * Complejidad esperada: O(n) tiempo y O(n) memoria.
 */
int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int n;
    long long target;
    if (!(cin >> n >> target)) return 0;

    unordered_map<long long, long long> prefixFrequency;
    prefixFrequency.reserve(static_cast<size_t>(n) * 2 + 1);
    prefixFrequency.max_load_factor(0.7f);

    // El prefijo vacío permite contar segmentos que empiezan en la posición 0.
    prefixFrequency[0] = 1;

    long long prefixSum = 0;
    long long answer = 0;
    for (int i = 0; i < n; ++i) {
        long long value;
        cin >> value;
        prefixSum += value;

        auto it = prefixFrequency.find(prefixSum - target);
        if (it != prefixFrequency.end()) {
            answer += it->second;
        }

        // Se inserta después de consultar para que el segmento tenga al menos
        // un elemento y no se cuente el intervalo vacío cuando K = 0.
        ++prefixFrequency[prefixSum];
    }

    cout << answer << '\n';
    return 0;
}
