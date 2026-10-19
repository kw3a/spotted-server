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
        long infinity = Long.MAX_VALUE / 4;
        long[] distance = new long[n];
        Arrays.fill(distance, infinity);
        PriorityQueue<long[]> queue = new PriorityQueue<>(Comparator.comparingLong(a -> a[0]));
        distance[0] = 0;
        queue.add(new long[]{0, 0});
        while (!queue.isEmpty()) {
            long[] current = queue.poll();
            long cost = current[0];
            int u = (int) current[1];
            if (cost != distance[u]) continue;
            for (long[] edge : graph.get(u)) {
                int v = (int) edge[0];
                long candidate = cost + edge[1];
                if (candidate < distance[v]) {
                    distance[v] = candidate;
                    queue.add(new long[]{candidate, v});
                }
            }
        }
        System.out.println(distance[n - 1] == infinity ? -1 : distance[n - 1]);
    }
}
