#include <algorithm>
#include <iostream>
#include <vector>

using namespace std;

/*
 * Resuelve consultas de suma en un arreglo que admite actualizaciones.
 * Las posiciones de entrada son 1-indexadas. Una operación "1 p x" cambia
 * el valor de la posición p a x; una operación "2 l r" imprime la suma del
 * intervalo inclusivo [l, r].
 *
 * Un árbol de Fenwick (BIT) representa las sumas de bloques cuyo tamaño es
 * una potencia de dos. El último bit encendido de un índice determina el
 * tamaño del bloque cubierto por ese índice. Así, sumar un prefijo visita
 * O(log n) bloques y actualizar un valor modifica O(log n) bloques.
 * Para asignar un valor nuevo guardamos también el arreglo actual y
 * propagamos solamente la diferencia entre el nuevo valor y el anterior.
 * Una suma de intervalo es prefijo(r) - prefijo(l - 1).
 *
 * Complejidad por operación: O(log n). Construcción: O(n log n).
 */
class FenwickTree {
public:
    explicit FenwickTree(int size) : tree(size + 1, 0) {}

    void add(int index, long long delta) {
        // index += index & -index avanza al siguiente bloque responsable.
        for (int i = index; i < static_cast<int>(tree.size()); i += i & -i) {
            tree[i] += delta;
        }
    }

    long long prefixSum(int index) const {
        long long result = 0;
        // Quitar el último bit encendido avanza al bloque anterior del prefijo.
        for (int i = index; i > 0; i -= i & -i) {
            result += tree[i];
        }
        return result;
    }

    long long rangeSum(int left, int right) const {
        return prefixSum(right) - prefixSum(left - 1);
    }

private:
    vector<long long> tree;
};

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    int n, q;
    if (!(cin >> n >> q)) return 0;

    vector<long long> values(n + 1);
    FenwickTree fenwick(n);
    for (int i = 1; i <= n; ++i) {
        cin >> values[i];
        fenwick.add(i, values[i]);
    }

    while (q--) {
        int type;
        cin >> type;
        if (type == 1) {
            int position;
            long long newValue;
            cin >> position >> newValue;
            long long difference = newValue - values[position];
            values[position] = newValue;
            fenwick.add(position, difference);
        } else {
            int left, right;
            cin >> left >> right;
            cout << fenwick.rangeSum(left, right) << '\n';
        }
    }
    return 0;
}
