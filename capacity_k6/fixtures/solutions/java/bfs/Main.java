import java.util.*;
public class Main {
    static class FastScanner {
        private final java.io.InputStream in = System.in;
        private final byte[] buffer = new byte[1 << 16];
        private int position = 0, length = 0;
        private int read() throws java.io.IOException {
            if (position >= length) {
                length = in.read(buffer);
                position = 0;
                if (length <= 0) return -1;
            }
            return buffer[position++];
        }
        long nextLong() throws java.io.IOException {
            int c;
            do { c = read(); } while (c <= 32 && c != -1);
            int sign = 1;
            if (c == '-') { sign = -1; c = read(); }
            long value = 0;
            while (c > 32 && c != -1) {
                value = value * 10 + c - '0';
                c = read();
            }
            return value * sign;
        }
        int nextInt() throws java.io.IOException { return (int) nextLong(); }
    }
    public static void main(String[] args) throws Exception {
        FastScanner in = new FastScanner();
        int n = in.nextInt(), m = in.nextInt();
        List<List<long[]>> graph = new ArrayList<>();
        for (int i = 0; i < n; i++) graph.add(new ArrayList<>());
        for (int i = 0; i < m; i++) {
            int u = in.nextInt() - 1, v = in.nextInt() - 1;
            graph.get(u).add(new long[]{v, in.nextLong()});
        }
        long[] distance = new long[n];
        Arrays.fill(distance, -1);
        ArrayDeque<Integer> queue = new ArrayDeque<>();
        distance[0] = 0;
        queue.add(0);
        while (!queue.isEmpty()) {
            int u = queue.remove();
            for (long[] edge : graph.get(u)) {
                int v = (int) edge[0];
                if (distance[v] == -1) {
                    distance[v] = distance[u] + edge[1];
                    queue.add(v);
                }
            }
        }
        System.out.println(distance[n - 1]);
    }
}
