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
        int n = in.nextInt(), capacity = in.nextInt();
        long[] best = new long[capacity + 1];
        for (int i = 0; i < n; i++) {
            int weight = in.nextInt();
            long value = in.nextLong();
            for (int limit = capacity; limit >= weight; --limit) {
                best[limit] = Math.max(best[limit], best[limit - weight] + value);
            }
        }
        System.out.println(best[capacity]);
    }
}
